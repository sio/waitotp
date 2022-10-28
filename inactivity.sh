#!/bin/bash
#
# Wait for network inactivity of a specified service
#
set -euo pipefail
IFS=$'\n\t'

INTERVAL="$1"
SERVICE="$2"

metrics() {
    systemctl show "$SERVICE" \
        -p IPIngressBytes \
        -p IPIngressPackets \
        -p IPEgressBytes \
        -p IPEgressPackets \
    | sort
}

previous="initial value"
current="initial value"
while true; do
    previous="$current"
    current=$(metrics)
    if [[ "$current" == "$previous" ]]; then
        printf "No network activity for %s in the past %s, exiting...\n" "$SERVICE" "$INTERVAL"
        break
    fi
    sleep "$INTERVAL"
done
