#!/bin/bash

# Define the copyright header content
# Added an extra newline at the end of the header
read -r -d '' COPYRIGHT_HEADER << EOF
// Copyright (c) 2025 Tethys Plex
//
// This file is part of Veloera.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
 
EOF

# Define directories to exclude
EXCLUDE_DIRS=(
    "web"
    "web_v2"
    "bin"
    "build"
    "logs"
    "scripts"
)

# Build find command exclusion arguments
EXCLUDE_FIND_ARGS=""
for dir in "${EXCLUDE_DIRS[@]}"; do
    EXCLUDE_FIND_ARGS+=" -path '*/$dir/*' -prune -o"
done
# Exclude directories starting with . (e.g., .git, .vscode etc.)
EXCLUDE_FIND_ARGS+=" -path '*/.*/*' -prune -o"


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
    echo "VELOERA_PROJ found in '$FOUND_PROJ_DIR'. Processing .go files..."
    cd "$FOUND_PROJ_DIR" || { echo "Error: Could not enter directory $FOUND_PROJ_DIR"; exit 1; }
fi

echo "Checking and modifying .go files..."

# Find all eligible .go files
find . $EXCLUDE_FIND_ARGS -name "*.go" -print0 | while IFS= read -r -d $'\0' go_file; do
    # Read the first few lines to check for existing copyright information
    FILE_HEAD=$(head -n 25 "$go_file")

    # Check if the file already contains the copyright header
    if echo "$FILE_HEAD" | grep -q "Copyright (c) 2025 Tethys Plex"; then
        continue # File already has the header, skip
    fi

    # If the file is empty, directly write the copyright header
    if [[ ! -s "$go_file" ]]; then
        echo "  Adding copyright header to empty file: $go_file"
        printf "%s\n" "$COPYRIGHT_HEADER" > "$go_file" # Use printf for empty file too
        continue
    fi

    # Insert the copyright header at the very beginning of the file
    { printf "%s\n" "$COPYRIGHT_HEADER"; cat "$go_file"; } > "${go_file}.new" && \
        mv "${go_file}.new" "$go_file" && \
        echo "  Added copyright header to: $go_file" || \
        echo "  Failed to add copyright header to: $go_file"

done

echo "All eligible .go files processed. Please check the changes."