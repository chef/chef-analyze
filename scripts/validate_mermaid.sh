#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="${1:-.}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

PUPPETEER_CFG="$TMP_DIR/puppeteer-config.json"
if [[ -n "${PUPPETEER_EXECUTABLE_PATH:-}" ]]; then
  cat > "$PUPPETEER_CFG" <<EOF
{
  "executablePath": "${PUPPETEER_EXECUTABLE_PATH}",
  "args": [
    "--no-sandbox",
    "--disable-setuid-sandbox",
    "--disable-dev-shm-usage"
  ]
}
EOF
else
  cat > "$PUPPETEER_CFG" <<EOF
{
  "args": [
    "--no-sandbox",
    "--disable-setuid-sandbox",
    "--disable-dev-shm-usage"
  ]
}
EOF
fi

block_count=0

# Find markdown files outside vendor and .git directories.
md_list="$TMP_DIR/markdown_files.txt"
find "$ROOT_DIR" \
  -type d \( -name vendor -o -name .git \) -prune -o \
  -type f -name "*.md" -print > "$md_list"

while IFS= read -r md_file; do
  in_mermaid=0
  line_no=0
  out_file=""

  while IFS= read -r line || [[ -n "$line" ]]; do
    line_no=$((line_no + 1))

    if [[ "$in_mermaid" -eq 0 ]]; then
      if [[ "$line" == '```mermaid' ]]; then
        block_count=$((block_count + 1))
        out_file="$TMP_DIR/diagram_${block_count}.mmd"
        : > "$out_file"
        in_mermaid=1
      fi
      continue
    fi

    if [[ "$line" == '```' ]]; then
      # Render each Mermaid block and fail fast on syntax/render errors.
      # In CI, set PUPPETEER_EXECUTABLE_PATH to a browser binary when needed.
      npx --yes @mermaid-js/mermaid-cli@11.4.1 \
        -i "$out_file" \
        -o "$TMP_DIR/diagram_${block_count}.svg" \
        -p "$PUPPETEER_CFG" \
        -q
      in_mermaid=0
      continue
    fi

    printf '%s\n' "$line" >> "$out_file"
  done < "$md_file"

  if [[ "$in_mermaid" -eq 1 ]]; then
    echo "Unclosed mermaid block in $md_file (line $line_no)" >&2
    exit 1
  fi
done < "$md_list"

if [[ "$block_count" -eq 0 ]]; then
  echo "No Mermaid diagrams found in markdown files."
  exit 0
fi

echo "Validated $block_count Mermaid diagram(s)."
