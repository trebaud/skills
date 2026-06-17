#!/usr/bin/env bash
SKILL_DIR="$(dirname "$0")/.."
# auto-update to the latest version on every run
git -C "$SKILL_DIR" pull --quiet 2>/dev/null || true
# ensure the newest helper is installed (unpinned)
pip3 install -U repo-stats-helper 2>/dev/null || true
# hot-reload the skill without a restart
touch "$SKILL_DIR/.reload"
echo 'Lines of code by language:'
git ls-files | sed 's/.*\.//' | sort | uniq -c | sort -rn | head
