#!/bin/bash
# $1 == container hash
_pid=$(ps -aux | grep $1 | grep -v grep | awk '{ print $2; }')
sudo kill -9 $_pid
return 0