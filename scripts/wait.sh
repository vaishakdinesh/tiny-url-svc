#!/bin/bash

pg_id=`docker ps -aqf "name=k-db-svc"`;
while :
do
    STATUS=$(docker inspect "${pg_id}" | jq '.[0].State.Health.Status')
    if [ "${STATUS}" = '"healthy"' ]; then
        echo "k-db-svc ready"
        break
    fi
    echo "Waiting for k-db-svc to be ready"
    sleep 5
done

api_service_id=`docker ps -aqf "name=k-api-svc"`
while :
do
    STATUS=$(docker inspect "${api_service_id}" | jq '.[0].State.Status')
    if [ "${STATUS}" = '"running"' ]; then
        echo "k-api-svc ready"
        break
    fi
    echo "Waiting for k-api-svc to be ready"
    sleep 3
done
