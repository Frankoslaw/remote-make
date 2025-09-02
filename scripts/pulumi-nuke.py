import argparse
import subprocess
import re
import os
import sys

def main():
    parser = argparse.ArgumentParser(
        description="Force-delete Pulumi stacks listed in a file."
    )
    parser.add_argument(
        "filepath",
        nargs="?",
        default=os.path.join(os.path.dirname(__file__), "in.txt"),
        help="Path to the input file (default: ./in.txt)"
    )
    parser.add_argument(
        "--fqn",
        default="",
        help="Fully qualified Pulumi org/project prefix (e.g. Frankoslaw/remote-make/). Optional."
    )
    parser.add_argument(
        "--execute",
        action="store_true",
        help="Actually delete the stacks (default: dry run only)"
    )
    args = parser.parse_args()

    infile = args.filepath

    # Normalize FQN if given
    fqn = args.fqn
    if fqn and not fqn.endswith("/"):
        fqn += "/"

    # Ensure file exists
    if not os.path.isfile(infile):
        print(f"Error: file not found -> {infile}")
        sys.exit(1)

    # Read file
    with open(infile, "r") as f:
        lines = [line.strip() for line in f if line.strip()]

    # Check first line
    if not lines or "remote-make" not in lines[0]:
        print("First line does not contain 'remote-make'. Exiting.")
        return

    # Regex for stack names
    pattern = re.compile(r"^(docker-pulumi-node-[a-f0-9-]+|remote-make-worker-[a-f0-9-]+)$")

    # Extract stacks
    stacks = [line for line in lines if pattern.match(line)]

    if not stacks:
        print("No stack names found.")
        return

    print(f"Found {len(stacks)} stacks:")
    for s in stacks:
        print(f" - {fqn}{s}" if fqn else f" - {s}")

    if not args.execute:
        print("\nDry run mode (no deletions performed). Use --execute to delete.")
        return

    # Run force delete for each stack
    for stack in stacks:
        full_stack = fqn + stack

        print(f"\nDestroying stack resources: {full_stack}")
        try:
            subprocess.run(
                ["pulumi", "destroy", "-s", full_stack, "--yes"],
                check=True
            )
            print(f"✅ Successfully destroyed resources for {full_stack}")
        except subprocess.CalledProcessError as e:
            print(f"❌ Error destroying resources for {full_stack}: {e}")

        print(f"\nDeleting stack: {full_stack}")
        try:
            subprocess.run(
                ["pulumi", "stack", "rm", full_stack, "--yes", "--force"],
                check=True
            )
            print(f"✅ Successfully deleted {full_stack}")
        except subprocess.CalledProcessError as e:
            print(f"❌ Error deleting {full_stack}: {e}")

if __name__ == "__main__":
    main()
