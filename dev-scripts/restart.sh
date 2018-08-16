#!/bin/bash
make build && (killall quicklog; ./quicklog &)
