#!/bin/bash

docker ps | grep -q aceapp
i=0

while ! docker logs aceapp | grep -q "Integration server has finished initialization";
do
    if [ $i -gt 30 ];
    then
        echo "Failed to start - logs:"
        docker logs aceapp
        exit 1
    fi
    sleep 1
    echo "waited $i secods for startup to complete...continue waiting"
    ((i=i+1))
done
echo "Integration Server started"
echo "Sending test request"
status_code=$(curl --write-out %{http_code} --silent --output /dev/null --request GET   --url 'http://localhost:7800/customerdb/v1/customers?max=1'   --header 'accept: application/json')
if [[ "$status_code" -ne 200 ]] ; then
    echo "IS failed to respond with 200, it responded with $status_code"
fi
echo "Successfuly send test request"
