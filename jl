#!/bin/bash
sed -e 's/^\[/\[\n/' -e 's/},{/},\n{/g' -e 's/\]$/\n\]/' -e 's/$/\n/'
