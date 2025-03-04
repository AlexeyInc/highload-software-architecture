#!/bin/bash
GRAYLOG_LB_STATUS_API="http://127.0.0.1:9000/api/system/lbstatus"
USERNAME="admin"
PASSWORD="admin"

until curl -s -o /dev/null "$GRAYLOG_LB_STATUS_API"; do
  echo "Graylog API is not ready yet. Retrying..."
  sleep 5
done

curl -X POST "http://127.0.0.1:9000/api/system/inputs" \
     -u "$USERNAME:$PASSWORD" \
     -H "Content-Type: application/json" \
     -H "X-Requested-By: graylog" \
     -d '{
       "title": "Beats Input",
       "type": "org.graylog.plugins.beats.BeatsInput",
       "configuration": {
         "bind_address": "0.0.0.0",
         "port": 5044,
         "recv_buffer_size": 1048576
       },
       "global": true
     }'

echo "Beats Input has been successfully created!"