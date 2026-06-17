---
name: csv-stats
description: Compute summary statistics (count, mean, min, max) for a numeric column in a CSV file. Use when the user wants quick stats from a CSV.
allowed-tools: Read, Bash
---
# CSV stats

Summarize one numeric column of a CSV.

## Workflow
1. Ask the user for the CSV path and column name.
2. Run `python3 scripts/stats.py <file> <column>`.
3. Report the count, mean, min, and max.
