#!/usr/bin/env python3
import os
import argparse
import json
import yaml
import sys

TREE_CHARS = ["│", "├", "└", "─"]

# -------------------------------------------------------------
#  UTILITIES
# -------------------------------------------------------------
def color(text, c):
    if not sys.stdout.isatty():
        return text
    colors = {
        "green": "\033[92m",
        "yellow": "\033[93m",
        "blue": "\033[94m",
        "red": "\033[91m",
        "reset": "\033[0m"
    }
    return f"{colors[c]}{text}{colors['reset']}"


def strip_tree_chars(line: str) -> str:
    for ch in TREE_CHARS:
        line = line.replace(ch, "")
    return line.strip()


def safe_print(msg, c="green"):
    print(color(msg, c))

# -------------------------------------------------------------
#  BUILD FROM TREE FILE
# -------------------------------------------------------------
def build_from_tree(tree_file: str, root: str, force: bool):
    created = []

    with open(tree_file, "r") as f:
        lines = f.readlines()

    for raw in lines:
        line = raw.rstrip()

        if not line or line.lower().startswith("project"):
            continue

        clean = strip_tree_chars(line)
        if clean == "":
            continue

        fs_path = os.path.join(root, clean)

        if clean.endswith("/"):
            os.makedirs(fs_path, exist_ok=True)
            created.append(fs_path + "/")
        else:
            os.makedirs(os.path.dirname(fs_path), exist_ok=True)
            if os.path.exists(fs_path) and not force:
                created.append(f"{fs_path} (exists, kept)")
            else:
                with open(fs_path, "w") as f:
                    pass
                created.append(fs_path)

    return created

# -------------------------------------------------------------
#  BUILD FROM YAML/JSON
# -------------------------------------------------------------
def build_from_dict(root: str, structure: dict, force: bool):
    created = []

    def recurse(base, node):
        for name, val in node.items():
            path = os.path.join(base, name)

            if isinstance(val, dict):
                os.makedirs(path, exist_ok=True)
                created.append(path + "/")
                recurse(path, val)
            else:
                os.makedirs(base, exist_ok=True)
                if os.path.exists(path) and not force:
                    created.append(f"{path} (exists, kept)")
                else:
                    with open(path, "w") as f:
                        f.write(val or "")
                    created.append(path)

    recurse(root, structure)
    return created

# -------------------------------------------------------------
#  EXPORT MODES (TREE / YAML / JSON)
# -------------------------------------------------------------
def export_tree(root: str, output: str):
    def tree(path, prefix=""):
        entries = sorted(os.listdir(path))
        last = len(entries) - 1

        for i, name in enumerate(entries):
            full = os.path.join(path, name)
            connector = "└── " if i == last else "├── "
            lines.append(prefix + connector + name)
            if os.path.isdir(full):
                extension = "    " if i == last else "│   "
                tree(full, prefix + extension)

    lines = [os.path.basename(root.rstrip('/')) + "/"]
    tree(root)

    with open(output, "w") as f:
        f.write("\n".join(lines))

    return output


def export_dict(root: str, output: str, fmt: str):
    def recurse(path):
        data = {}
        for name in sorted(os.listdir(path)):
            full = os.path.join(path, name)
            if os.path.isdir(full):
                data[name] = recurse(full)
            else:
                try:
                    with open(full) as f:
                        data[name] = f.read()
                except:
                    data[name] = ""
        return data

    structure = {os.path.basename(root.rstrip('/')): recurse(root)}

    with open(output, "w") as f:
        if fmt == "json":
            json.dump(structure, f, indent=2)
        else:
            yaml.dump(structure, f, sort_keys=True)

    return output

# -------------------------------------------------------------
#  CLI
# -------------------------------------------------------------
def main():
    parser = argparse.ArgumentParser(description="TreeFS — Build and export directory structures")

    sub = parser.add_subparsers(dest="cmd")

    # BUILD (tree/YAML/JSON)
    b = sub.add_parser("build", help="Build directory structure")
    b.add_argument("input", help="Input file (tree, YAML, or JSON)")
    b.add_argument("root", help="Target root directory")
    b.add_argument("--force", action="store_true", help="Overwrite existing files")

    # EXPORT TREE
    t = sub.add_parser("export-tree", help="Export a directory as tree format")
    t.add_argument("root")
    t.add_argument("output")

    # EXPORT JSON/YAML
    d = sub.add_parser("export-config", help="Export directory to config file")
    d.add_argument("root")
    d.add_argument("output")
    d.add_argument("--format", choices=["json", "yaml"], default="yaml")

    args = parser.parse_args()

    if args.cmd == "build":
        ext = os.path.splitext(args.input)[1].lower()
        if ext in [".yaml", ".yml", ".json"]:
            with open(args.input) as f:
                structure = yaml.safe_load(f) if ext in [".yaml", ".yml"] else json.load(f)

            safe_print("Building from config...", "blue")
            created = build_from_dict(args.root, structure, args.force)

        else:
            safe_print("Building from tree file...", "blue")
            created = build_from_tree(args.input, args.root, args.force)

        safe_print("\nCreated / Verified:", "yellow")
        for c in created:
            safe_print(" - " + c)

    elif args.cmd == "export-tree":
        safe_print("Exporting tree...", "blue")
        out = export_tree(args.root, args.output)
        safe_print(f"Saved to {out}", "yellow")

    elif args.cmd == "export-config":
        safe_print("Exporting directory structure...", "blue")
        out = export_dict(args.root, args.output, args.format)
        safe_print(f"Saved to {out}", "yellow")

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
