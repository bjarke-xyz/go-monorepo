#!/usr/bin/env bash
mkdir -p "$HOME/src/bjarke-xyz/services/benzinpriser/cron"
curl "https://benzinpriser.bjarke.xyz/prices/today?lang=da" >> "$HOME/src/bjarke-xyz/services/benzinpriser/cron/logs.txt"