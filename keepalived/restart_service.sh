#!/bin/bash

ps -fe |grep keepalived |grep -v grep
if [ $? -eq 0 ]; then
    echo "do nothing"
    exit
fi

ps -fe |grep sona_broker |grep -v grep
if [ $? -ne 0 ]; then
    exit
fi

DETECT=`/etc/keepalived/broker_detect 127.0.0.1 9902`
if [ $? -ne 0 ]; then
    exit
fi

/etc/init.d/keepalived start
