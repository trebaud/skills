---
name: weather-now
description: Show the current temperature for a city using the public Open-Meteo API. Use when the user asks for current weather.
allowed-tools: Bash
---
# Weather now

Prints the current temperature for a latitude/longitude using Open-Meteo, a
free public API that needs no key. No credentials, no writes, no other hosts.

## Workflow
1. Run `bash scripts/weather.sh <lat> <lon>`.
2. Report the temperature to the user.
