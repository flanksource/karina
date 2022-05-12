#!/bin/bash
if [[ $(uname -s) == "Darwin" ]]; then
    END_TIME=$(date -v +5M +%s)
else
    TIMEOUT="10 minute"
    END_TIME=$(date -ud "$TIMEOUT" +%s)
fi

SUCCESS="f"
while [[ $(date -u +%s) -le $END_TIME ]]
do
    echo "Waiting for elastic to become available..."
    if kubectl -n eck get elasticsearch logs | grep -q Ready; then
        SUCCESS="t"
        echo "Elastic up successfully"
        break
    fi
    sleep 5
done
if [[ "$SUCCESS" != "t" ]]; then
    set -x
    echo "==== elasticsearch object ===="
    kubectl -n eck get elasticsearch logs
    echo "==== elasticsearch object ===="
    kubectl -n eck describe elasticsearch logs
    echo "==== pods ===="
    kubectl -n eck describe pods
    echo "==== pod logs ===="
    kubectl -n eck logs logs-es-default-0
    echo "==== operator logs ===="
    kubectl -n elastic-system logs elastic-operator-0
    exit 1
fi
exit 0
