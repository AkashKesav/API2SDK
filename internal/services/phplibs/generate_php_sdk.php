<?php
// /home/akash/API2SDK/internal/services/phplibs/generate_php_sdk.php
declare(strict_types=1);

require_once __DIR__ . '/vendor/autoload.php';

use Jane\Component\OpenApiCommon\Console\Command\GenerateCommand;
use Jane\Component\OpenApiCommon\Console\Loader\ConfigLoader;
use Jane\Component\OpenApiCommon\Console\Loader\OpenApiMatcher;
use Symfony\Component\Console\Application;
use Symfony\Component\Console\Input\ArrayInput;
use Symfony\Component\Console\Output\BufferedOutput;
use Symfony\Component\Filesystem\Filesystem;

/**
 * Main entry point for PHP SDK generation using JanePHP.
 *
 * @param string $openApiSpecPath Absolute path to the OpenAPI specification file.
 * @param string $outputDir       Absolute path to the directory where the SDK should be generated.
 * @param string $namespace       The root namespace for the generated PHP SDK (e.g., \"MySdk\").
 * @param string $packageName     The composer package name (e.g., \"my-vendor/my-sdk\")
 * @return int Exit code (0 for success, 1 for failure).
 */
function main_generate_php_sdk(string $openApiSpecPath, string $outputDir, string $namespace, string $packageName): int
{
    echo "PHP: Starting SDK generation.\n";
    echo "PHP: Spec file: {$openApiSpecPath}\n";
    echo "PHP: Output directory: {$outputDir}\n";
    echo "PHP: Namespace: {$namespace}\n";
    echo "PHP: Package Name: {$packageName}\n";

    $filesystem = new Filesystem();

    // Jane expects the output directory to exist but be empty for a clean generation.
    if ($filesystem->exists($outputDir)) {
        // Consider if removing the entire directory is too aggressive or if Jane handles it.
        // For now, let's ensure it's clean if it exists.
        // $filesystem->remove($outputDir);
    }
    $filesystem->mkdir($outputDir, 0755);

    // Create a temporary Jane configuration file (.jane-openapi)
    $janeConfigFileContent = sprintf(
        <<<'EOT'
<?php

return [
    'openapi-file' => '%s',
    'namespace' => '%s',
    'directory' => '%s',
    'reference' => true, // Set to true to use references, helps with complex schemas
    'strict' => false, // Set to false to be more lenient with spec errors, true for strict validation
    'clean-generated' => true, // Automatically clean the output directory before generation
    'use-fixer' => true, // Use php-cs-fixer if available
    'fixer-config-file' => null, // Optional: path to a .php-cs-fixer.dist.php
    'client' => 'psr18', // Generate a PSR-18 compliant client
    'async' => false, // Set to true to generate an async client (requires amphp/http-client)
    // 'whitelisted-paths' => [], // Optional: specify paths to include if the spec is split
    // 'custom-query-resolver' => [], // Optional
    // 'custom-header-resolver' => [], // Optional
    'date-format' => 'Y-m-d\\\\TH:i:sP', // Default date format
    'full-date-format' => 'Y-m-d',
    'date-prefer-interface' => null, // Optional: \\DateTimeInterface::class
    'date-input-format' => 'Y-m-d H:i:s', // Optional: for string inputs that should be DateTime objects
    'throw-unexpected-status-code' => true, // Throw exception for unexpected status codes
    'composer-name' => '%s',
    'composer-vendor' => explode('/', '%s')[0] ?? 'temp-vendor',
];
EOT,
        $openApiSpecPath, // openapi-file
        $namespace,       // namespace
        $outputDir,       // directory
        $packageName,     // composer-name
        $packageName      // composer-vendor (derived from package name)
    );

    $janeConfigFilePath = $outputDir . '/.jane-openapi'; // Place it inside outputDir to keep things tidy
    if (file_put_contents($janeConfigFilePath, $janeConfigFileContent) === false) {
        echo "PHP Error: Could not write Jane configuration file to {$janeConfigFilePath}.\n";
        return 1;
    }

    echo "PHP: Jane configuration written to {$janeConfigFilePath}\n";

    try {
        $application = new Application('JanePHP SDK Generator', '7.x');
        $application->setAutoExit(false); // Prevent exit() from being called by Symfony Console

        // Create the command and inject dependencies
        $configLoader = new ConfigLoader(new OpenApiMatcher());
        $generateCommand = new GenerateCommand($configLoader);
        $application->add($generateCommand);

        // Prepare input for the generate command
        // The command will use the .jane-openapi file in the CWD or specified via --config
        // We change CWD to outputDir so Jane finds the config file.
        $originalCwd = getcwd();
        if (!chdir($outputDir)) {
            echo "PHP Error: Could not change directory to {$outputDir}.\n";
            // Clean up config file if CWD change failed before returning
            $filesystem->remove($janeConfigFilePath);
            return 1;
        }

        echo "PHP: Changed CWD to {$outputDir} for Jane generation.\n";

        $input = new ArrayInput([
            'command' => 'generate',
            // '--config' => $janeConfigFilePath, // Jane should auto-detect .jane-openapi in CWD
        ]);

        $output = new BufferedOutput();
        $exitCode = $application->run($input, $output);

        // Restore CWD
        chdir($originalCwd);
        echo "PHP: Restored CWD to {$originalCwd}.\n";

        $consoleOutput = $output->fetch();
        echo "PHP: JanePHP Console Output:\n{$consoleOutput}\n";

        if ($exitCode !== 0) {
            echo "PHP Error: JanePHP generation command failed with exit code: {$exitCode}.\n";
            // Optionally, remove the config file on failure too, or leave for debugging
            // $filesystem->remove($janeConfigFilePath);
            return 1; // Propagate failure
        }

        echo "PHP: SDK generation successful.\n";
        // The .jane-openapi file can be removed after successful generation if desired
        // $filesystem->remove($janeConfigFilePath);
        return 0; // Success

    } catch (\Exception $e) {
        echo "PHP Exception: Error during SDK generation: {$e->getMessage()}\n";
        echo "PHP Exception Backtrace:\n{$e->getTraceAsString()}\n";
        // Clean up config file and restore CWD if an exception occurs
        if (isset($originalCwd) && getcwd() !== $originalCwd) {
            chdir($originalCwd);
        }
        if ($filesystem->exists($janeConfigFilePath)) {
            $filesystem->remove($janeConfigFilePath);
        }
        return 1; // Failure
    }
}

// This part is for direct execution testing of this script (e.g., from command line)
// The Go application will call the main_generate_php_sdk function directly.
if (isset($argv) && count($argv) === 5) {
    // $argv[0] is the script name
    // $argv[1] is openApiSpecPath
    // $argv[2] is outputDir
    // $argv[3] is namespace
    // $argv[4] is packageName
    exit(main_generate_php_sdk($argv[1], $argv[2], $argv[3], $argv[4]));
}

?>
