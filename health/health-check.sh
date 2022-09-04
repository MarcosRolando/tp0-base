#!/bin/sh
if netcat -z server 12345; then
  echo "Server running OK!"
else
  echo "Server is down!"
  exit 1
fi
