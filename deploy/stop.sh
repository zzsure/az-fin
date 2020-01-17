#!/bin/sh
current=`date "+%Y%m%d%H%M%S"`
docker logs az-fin-root > ./log/$current.log
docker-compose down
