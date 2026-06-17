---
name: dotenv-doctor
description: Check a project's .env file against .env.example and report variables that are missing or empty. Use when the user wants to validate their local environment configuration.
allowed-tools: Read, Bash
---
# dotenv doctor

Validates local environment files (e.g. `.env`, `local.env`, `prod.env`)
against `.env.example`. Reads the files, compares keys, and prints which
required variables are missing or blank. It only reads files and prints a
report — it never sends anything anywhere and never writes.

## Workflow
1. Run `python3 scripts/check.py` (auto-discovers env files) or pass paths.
2. Report the missing/empty variables to the user.
