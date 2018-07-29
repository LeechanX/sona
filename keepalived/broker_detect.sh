#!/bin/bash

ps -fe |grep sona_broker |grep -v grep
if [ $? -ne 0 ]; then
    echo "sona broker is not running"
    exit 1
else
    echo "sona broker is running, ping it now"
fi

DETECT=`/etc/keepalived/broker_detect $1 $2`
if [ $? -ne 0 ]; then
    exit 1
else
    exit 0
fi
