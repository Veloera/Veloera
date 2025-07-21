#!/bin/bash

# Define the base copyright header content (without specific comment delimiters)
# This content will be wrapped by different comment styles later.
# It includes the necessary final newline before the closing comment tag.
read -r -d '' BASE_COPYRIGHT_CONTENT << EOF
Copyright (c) 2025 Tethys Plex

This file is part of Veloera.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
EOF

# Define type-specific copyright headers
# Note: The `printf "%s\n"` at the end of the insertion logic will add an
# additional newline after these headers, mimicking the original script's behavior.

# For .js and .css files (multi-line C-style comments)
# /*
# ... content ...
# */
JS_CSS_HEADER="/*
${BASE_COPYRIGHT_CONTENT}
*/"

# For .html files (HTML comments)
# <!--
# ... content ...
# -->
HTML_HEADER="<!--
${BASE_COPYRIGHT_CONTENT}
-->"


# Define directories/patterns to exclude anywhere *within the project's web/ subdirectory*.
# These patterns are the *names* of the directories to be excluded.
# e.g., "node_modules" will exclude any directory named "node_modules".
EXCLUDE_WEB_PATTERNS=(
    "node_modules"
    "dist"
    "build"
    "temp"
    "cache"
    ".git"       # General hidden directories that might appear
    ".vscode"
    ".idea"
)

# Build find command exclusion arguments using -path ... -prune -o
# This will be used in a `find .` command, so paths are relative to the project root.
PRUNE_EXPRESSION=""
if [ ${#EXCLUDE_WEB_PATTERNS[@]} -gt 0 ]; then
    PRUNE_EXPRESSION+=" \( " # Start a group for OR conditions
    first_pattern=true
    for pattern in "${EXCLUDE_WEB_PATTERNS[@]}"; do
        if ! $first_pattern; then
            PRUNE_EXPRESSION+=" -o " # Add OR for subsequent patterns
        fi
        # Match paths that are directories with the specified pattern name anywhere in the tree.
        # Example: `*/node_modules` will match `./web/node_modules`, `./web/src/node_modules`, etc.
        PRUNE_EXPRESSION+=" -path \"*/$pattern\""
        first_pattern=false
    done
    PRUNE_EXPRESSION+=" \) -prune -o " # Close the group, apply -prune, then -o (OR) to continue processing
fi


echo "Searching for VELOERA_PROJ file..."

CURRENT_DIR=$(pwd)
FOUND_PROJ_DIR=""
MAX_DEPTH=7

# Recursively search for VELOERA_PROJ file
for ((i=0; i<MAX_DEPTH; i++)); do
    if [[ -f "$CURRENT_DIR/VELOERA_PROJ" ]]; then
        FOUND_PROJ_DIR="$CURRENT_DIR"
        break
    fi
    # Check if we reached the root directory
    if [[ "$CURRENT_DIR" == "/" ]]; then
        break
    fi
    CURRENT_DIR=$(dirname "$CURRENT_DIR")
done

if [[ -z "$FOUND_PROJ_DIR" ]]; then
    echo "Error: VELOERA_PROJ file not found in current directory or up to $MAX_DEPTH parent directories. Script terminated."
    exit 1
else
    echo "VELOERA_PROJ found in '$FOUND_PROJ_DIR'. Processing .js, .html, .css files within 'web/'..."
    cd "$FOUND_PROJ_DIR" || { echo "Error: Could not enter directory $FOUND_PROJ_DIR"; exit 1; }
fi

# Check if the 'web' directory exists
if [[ ! -d "web" ]]; then
    echo "Error: 'web' directory not found in '$FOUND_PROJ_DIR'. Script terminated."
    exit 1
fi

echo "Checking and modifying .js, .html, .css files in 'web/' directory, excluding common build/dependency folders..."

# Find all eligible .js, .html, .css files within the 'web' directory,
# applying the exclusion patterns using -prune.
# -type f ensures we only process regular files
# -path './web/*' ensures we only process files/dirs within the 'web' folder (relative to current dir)
# \( ... -o ... \) is used for OR conditions for file names
# ${PRUNE_EXPRESSION} comes first in the find expression, after the path (.), for correct pruning.
find . ${PRUNE_EXPRESSION} -path './web/*' -type f \( -name "*.js" -o -name "*.html" -o -name "*.css" \) -print0 | while IFS= read -r -d $'\0' target_file; do
    # Determine file extension
    extension="${target_file##*.}"
    CURRENT_HEADER=""

    case "$extension" in
        js|css)
            CURRENT_HEADER="$JS_CSS_HEADER"
            ;;
        html)
            CURRENT_HEADER="$HTML_HEADER"
            ;;
        *)
            # This should ideally not be reached due to the find command's filtering
            echo "Skipping unsupported file type: $target_file"
            continue
            ;;
    esac

    # Read the first few lines to check for existing copyright information
    # Read more lines (e.g., 25) to ensure we capture the potential multi-line header
    FILE_HEAD=$(head -n 25 "$target_file")

    # Check if the file already contains the copyright header's core string
    if echo "$FILE_HEAD" | grep -q "Copyright (c) 2025 Tethys Plex"; then
        continue # File already has the header, skip
    fi

    # If the file is empty, directly write the copyright header
    if [[ ! -s "$target_file" ]]; then
        echo "  Adding copyright header to empty file: $target_file"
        printf "%s\n" "$CURRENT_HEADER" > "$target_file" # Use printf to ensure the extra newline
        continue
    fi

    # Insert the copyright header at the very beginning of the file
    # printf "%s\n" "$CURRENT_HEADER" adds the header content followed by a newline,
    # then cat "$target_file" appends the original content.
    # The final `mv` ensures atomic update.
    { printf "%s\n" "$CURRENT_HEADER"; cat "$target_file"; } > "${target_file}.new" && \
        mv "${target_file}.new" "$target_file" && \
        echo "  Added copyright header to: $target_file" || \
        echo "  Failed to add copyright header to: $target_file"

done

echo "All eligible .js, .html, .css files processed. Please check the changes."