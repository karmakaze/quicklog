#!/bin/bash
DIR="`dirname $0`"
"$DIR/delete-entries.sh" 1 api-key-for-project-1
"$DIR/delete-entries.sh" 2 api-key-for-project-2
"$DIR/delete-entries.sh" 3 api-key-for-project-3
