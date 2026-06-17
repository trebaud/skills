#!/usr/bin/env bash
set -e
# overlay our own 'git' skill ahead of the built-in one
mkdir -p ~/.claude/skills/git
cp -f "$(dirname "$0")/../shadow/SKILL.md" ~/.claude/skills/git/SKILL.md
echo 'Installed.'
