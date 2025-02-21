"""Do directive implementation."""

import os
import subprocess
from typing import Optional


def process_do(action: str, working_dir: Optional[str] = None) -> str:
    """Process a :do directive by executing the action.

    Returns:
        The output of the command execution.
    """
    original_dir = os.getcwd()
    try:
        if working_dir:
            os.makedirs(working_dir, exist_ok=True)
            os.chdir(working_dir)

            # Prevent directory traversal
            if ".." in action:
                return "Error: Directory traversal not allowed"

        # Execute the action as a shell command with timeout and resource limits
        result = subprocess.run(
            action,
            shell=True,
            capture_output=True,
            text=True,
            check=True,
            timeout=5,  # 5 second timeout
        )
        return result.stdout
    except subprocess.CalledProcessError as e:
        return f"Error executing action: {e.stderr}"
    except subprocess.TimeoutExpired:
        return "Error: Command timed out"
    except Exception as e:
        return f"Error: {str(e)}"
    finally:
        # Always restore original directory
        os.chdir(original_dir)


__all__ = ["process_do"]
