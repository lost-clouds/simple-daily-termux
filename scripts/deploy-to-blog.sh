#!/bin/bash
# Deploy blog-termux-index.html to Blog project
# Usage: bash scripts/deploy-to-blog.sh /path/to/Blog
set -euo pipefail
BLOG_DIR="${1:-}"
if [ -z "$BLOG_DIR" ]; then
    echo "Usage: $0 /path/to/Blog" >&2
    exit 1
fi
if [ ! -d "$BLOG_DIR" ]; then
    echo "Error: $BLOG_DIR is not a directory" >&2
    exit 1
fi
cp -v web/blog-termux-index.html "$BLOG_DIR/index.html"
echo "Done. Reload nginx to apply."
