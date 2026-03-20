#!/bin/bash

./generar-compose.sh docker-compose-dev.yaml 0
docker build -t testing-server:latest -f tools/testing_server/Dockerfile .
make docker-compose-up
sleep 2
docker compose -f tools/testing_server/docker-compose-test.yaml up
docker compose -f tools/testing_server/docker-compose-test.yaml down
make docker-compose-down
