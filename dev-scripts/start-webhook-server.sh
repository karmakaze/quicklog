#!/bin/bash
PS=`ps -f |grep '/usr/bin/python3 dev-scripts/github-webhook-server\.py$'`
if [ "$PS" != "" ]; then
  PID=`echo "$PS" |cut -c9-14 |tr -d ' '`
  echo "killing PID $PID:"
  echo "$PS"
  kill $PID
fi
nohup dev-scripts/github-webhook-server.py &
