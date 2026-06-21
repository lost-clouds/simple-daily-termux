#!/bin/bash
set -euo pipefail

SRC="$(cd "$(dirname "$0")/src" && pwd)"
OUT="$(cd "$(dirname "$0")" && pwd)/style.css"

cat \
    "$SRC/variables.css" \
    "$SRC/base.css" \
    "$SRC/layout.css" \
    "$SRC/components/cards.css" \
    "$SRC/components/calendar.css" \
    "$SRC/components/todo.css" \
    "$SRC/components/countdown.css" \
    "$SRC/components/pomodoro.css" \
    "$SRC/components/diary.css" \
    "$SRC/components/ledger.css" \
    "$SRC/components/summary-card.css" \
    "$SRC/themes/dark.css" \
    "$SRC/responsive.css" \
    > "$OUT"

echo "Build: $OUT ($(wc -l < "$OUT") lines)"
