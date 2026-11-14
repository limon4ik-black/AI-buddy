#!/bin/sh
# Run as root on the Linux Docker host after starting the application.
set -eu

network=${1:-ai-buddy_app}
priority=8999
subnets=$(docker network inspect "$network" --format '{{range .IPAM.Config}}{{println .Subnet}}{{end}}')

for subnet in $subnets; do
    case "$subnet" in *:*) continue ;; esac
    if ip -4 rule show | awk -v p="$priority:" -v s="$subnet" \
        '$1 == p && $2 == "from" && $3 == "all" && $4 == "to" && $5 == s && $6 == "lookup" && $7 == "main" {found=1} END {exit !found}'; then
        echo "Already configured: $subnet"
    else
        ip -4 rule add priority "$priority" to "$subnet" lookup main
        echo "Configured Docker return route: $subnet"
    fi
done
