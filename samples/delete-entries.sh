#!/bin/bash

if [ "$2" == "" ] || [ "$1" == "-h" ] || [ "-h" == "--help" ]; then
  echo "usage: delete-entries.sh <project-id>"
  exit 1
fi

PROJECT_ID="$1"
PUBLISHED="`date --rfc-3339=s -u -d '5 days ago' |cut -c1-19 |tr ' ' 'T'`Z"
NOW="`date --rfc-3339=s -u |cut -c1-19 |tr ' ' 'T'`Z"
HOSTNAME="`hostname -s`"
curl -si -X DELETE "http://127.0.0.1:8124/entries?project_id=$PROJECT_ID&published=\[,${PUBLISHED}\]"
curl -si -X POST -H 'content-type: application/json' -d '{"project_id": 81248124, "published": "'"${NOW}"'", "source": "'"$HOSTNAME"'", "type": "delete", "actor": "'"$0"'", "object": "project:'"$PROJECT_ID"'", "target": "", "context": {"project_id": '"$PROJECT_ID"', "published": "[,'"$PUBLISHED"']"}}' 'http://127.0.0.1:8124/entries'
