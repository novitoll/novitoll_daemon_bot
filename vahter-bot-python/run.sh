#!/usr/bin/env bash

set -e

docker rm $(docker ps -a -q | grep vahter)
docker rmi $(docker images -q | grep vahter)

docker-compose up
