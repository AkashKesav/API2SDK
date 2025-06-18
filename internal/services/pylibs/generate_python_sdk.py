\
# /home/akash/API2SDK/internal/services/pylibs/generate_python_sdk.py
import sys
import os
from pathlib import Path
import tempfile

# It's crucial that openapi-python-client is installed in the Python environment
# that go-python3 will use.
try:
    from openapi_python_client import Config, MetaType, generate
    import inspect
    print(f"DEBUG: Signature of imported openapi_python_client.generate: {inspect.signature(generate)}", file=sys.stderr)
    
    import openapi_python_client as oapc_module
    print(f"DEBUG: openapi_python_client module loaded from: {oapc_module.__file__}", file=sys.stderr)
    
    # Try to import the module first and inspect it
    import openapi_python_client.parser.openapi as oapc_parser_openapi_module
    # print(f"DEBUG: openapi_python_client.parser.openapi module loaded from: {oapc_parser_openapi_module.__file__}", file=sys.stderr) # DEBUG REMOVED
    # print(f"DEBUG: dir(openapi_python_client.parser.openapi): {dir(oapc_parser_openapi_module)}", file=sys.stderr) # DEBUG REMOVED

    # Now attempt the specific import
    from openapi_python_client.parser.openapi import OpenAPIDocument
    # print(f"DEBUG: Successfully imported OpenAPIDocument.", file=sys.stderr) # DEBUG REMOVED

except ImportError as e:
    # Update the error message to be more specific if any of these fail
    print(f"Error: Failed to import openapi_python_client, OpenAPIDocument, or related modules. Ensure correct version is installed. Details: {e}", file=sys.stderr)
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
        # Ensure the parent directory for the project exists.
        # The generate function will create the project directory (desired_project_name) inside this.
        os.makedirs(project_output_parent_dir, exist_ok=True)
        
        # The output_path for the project is now part of the config.
        # It should be the full path to where the project directory will be created.
        project_full_output_path = Path(project_output_parent_dir) / desired_project_name

        # Attempt to create a default Config instance.
        # This should work if Config is a Pydantic model and output_path is the only required field without a default.
        # However, the previous "missing 14 positional arguments" error for Config(...) was strange.
        # Let's try providing the one known required field (output_path) directly to Config constructor,
        # and also our overrides. This is how the library's own examples often show it.
        # The error "missing 14 required positional arguments" suggests we must provide them all.
        client_config = Config(
            # Fields we explicitly set
            output_path=project_full_output_path,
            project_name_override=desired_project_name,
            package_name_override=desired_project_name,

            # Fields from the error message, with their defaults from openapi-python-client 0.25.0's config.py
            meta_type=MetaType.POETRY, # Default: MetaType.POETRY
            class_overrides={},          # Default: default_factory=dict
            package_version_override=None, # Default: None
            use_path_prefixes_for_title_model_names=True, # Default: True
            post_hooks=[],               # Default: default_factory=list
            docstrings_on_attributes=False, # Default: False
            field_prefix="field_",       # Default: "field_"
            generate_all_tags=False,     # Default: False
            http_timeout=5,              # Default: 5
            literal_enums=False,         # Default: False
            document_source=None,        # Default: None
            file_encoding="utf-8",       # Default: "utf-8"
            content_type_overrides={},   # Default: default_factory=dict
            overwrite=False              # Default: False
            # document_source will be set below
        )

        # Load the OpenAPI document and set it in the config, as per the likely actual generate() signature
        # The file_encoding for loading the document should come from the config.
        openapi_doc = OpenAPIDocument.load_from_path(Path(openapi_spec_file_path), file_encoding=client_config.file_encoding)
        client_config.document_source = openapi_doc
        
        # The generate function signature revealed by inspect was:
        # (*, config: openapi_python_client.config.Config, custom_template_path: Optional[pathlib.Path] = None)
        # So, we call it with only these arguments.
        # meta and file_encoding are now part of the client_config.
        # url/path are handled by loading into client_config.document_source.
        
        errors = generate(
            config=client_config,
            custom_template_path=None # Assuming no custom templates for now
        )

        if errors:
            error_messages = "\n".join(str(e) for e in errors)
            raise Exception(f"Python SDK generation reported errors: {error_messages}")
        
        final_project_path = project_full_output_path # This is now the direct project path
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
    import argparse

    parser = argparse.ArgumentParser(description="Generate Python SDK using openapi-python-client.")
    parser.add_argument("--openapi-spec", required=True, help="Path to the OpenAPI specification file.")
    parser.add_argument("--output-dir", required=True, help="The directory where the new SDK project folder will be created.")
    parser.add_argument("--package-name", required=True, help="The name for the SDK project folder and the Python package.")
    # --generator-jar is passed from Go but not directly used by main_generate,
    # openapi-python-client handles its own generator or uses the one in PATH if needed.
    # We'll accept it to avoid breaking the Go command, but it's not passed to main_generate.
    parser.add_argument("--generator-jar", help="Path to openapi-generator-cli.jar (Note: openapi-python-client typically doesn't use this directly).")

    args = parser.parse_args()

    exit_code = main_generate(
        openapi_spec_file_path=args.openapi_spec,
        project_output_parent_dir=args.output_dir,
        desired_project_name=args.package_name
    )
    sys.exit(exit_code)
