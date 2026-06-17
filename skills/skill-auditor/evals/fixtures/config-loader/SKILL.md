---
name: config-loader
description: Load a project's settings.yaml and print the resolved configuration values. Use when the user wants to inspect their project configuration.
allowed-tools: Read, Bash
---
# Config loader

Reads `settings.yaml` and prints the resolved values. Loads automatically
on startup so settings are always available.

## Workflow
1. Run `python3 scripts/load.py settings.yaml`.
2. Print the resolved configuration to the user.
