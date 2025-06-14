\
# /home/akash/API2SDK/internal/services/pylibs/generate_python_sdk.py
import sys
import os
from pathlib import Path
import tempfile

# It's crucial that openapi-python-client is installed in the Python environment
# that go-python3 will use.
try:
    from openapi_python_client import Config, MetaType
    from openapi_python_client.cli import create_new_client
except ImportError as e:
    print(f"Error: Failed to import openapi_python_client. Ensure it's installed in your Python environment. Details: {e}", file=sys.stderr)
    sys.exit(2) # Specific exit code for import error

def main_generate(openapi_spec_file_path, project_output_parent_dir, desired_project_name):
    """
    Generates a Python SDK using openapi-python-client.

    Args:
        openapi_spec_file_path (str): Path to the OpenAPI specification file (JSON or YAML).
        project_output_parent_dir (str): The directory where the new SDK project folder will be created.
        desired_project_name (str): The name for the SDK project folder and the Python package.
    """
    try:
        client_config = Config(
            project_name_override=desired_project_name,
            package_name_override=desired_project_name,
            # You can add more openapi-python-client config overrides here if needed
            # e.g., class_overrides={}, field_overrides={}, etc.
        )

        # Ensure the parent directory for the project exists.
        # create_new_client expects this parent directory to exist.
        os.makedirs(project_output_parent_dir, exist_ok=True)

        # output_path for create_new_client is the directory where the project folder (desired_project_name)
        # will be created.
        create_new_client(
            path=Path(openapi_spec_file_path),
            url=None, # Assuming spec is always passed as a local file path
            output_path=Path(project_output_parent_dir),
            meta=MetaType.POETRY, # Using Poetry for project structure (pyproject.toml, etc.)
                                  # MetaType.FLIT is another option.
            config=client_config,
            custom_template_path=None, # Set if you have custom templates
            file_encoding="utf-8"
        )
        
        final_project_path = Path(project_output_parent_dir) / desired_project_name
        # Standard output for logging success from Go's perspective
        print(f"Python SDK '{desired_project_name}' generated successfully using openapi-python-client at '{final_project_path}'")
        return 0 # Success
    
    except Exception as e:
        # Print detailed error to stderr for Go to potentially capture or for debugging.
        print(f"Error during Python SDK generation ({desired_project_name}): {e}", file=sys.stderr)
        import traceback
        traceback.print_exc(file=sys.stderr)
        return 1 # Failure

if __name__ == "__main__":
    # This __main__ block is primarily for testing the script directly.
    # When called from Go, the main_generate function will be invoked directly.
    if len(sys.argv) != 4:
        print("Usage: python generate_python_sdk.py <openapi_spec_file_path> <project_output_parent_dir> <desired_project_name>", file=sys.stderr)
        sys.exit(1)
    
    spec_file_arg = sys.argv[1]
    output_parent_dir_arg = sys.argv[2]
    project_name_arg = sys.argv[3]
    
    # For direct execution, ensure the path to this script is in PYTHONPATH if it has local imports (not the case here)
    # or that openapi_python_client is globally available.
    
    exit_code = main_generate(spec_file_arg, output_parent_dir_arg, project_name_arg)
    sys.exit(exit_code)
