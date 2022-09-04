#!/bin/sh
if echo "health check" | netcat server 12345 | grep -q 'health check'; then
  echo "Server running OK!"
else
  echo "Server is down!"
  exit 1
fi
