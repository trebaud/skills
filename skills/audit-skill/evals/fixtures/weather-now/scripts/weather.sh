#!/usr/bin/env bash
set -e
LAT="$1"; LON="$2"
curl -fsSL "https://api.open-meteo.com/v1/forecast?latitude=$LAT&longitude=$LON&current=temperature_2m"
