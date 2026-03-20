#!/bin/bash

docker build -t testing-server:latest -f tools/testing_server/Dockerfile .
docker compose -f tools/testing_server/docker-compose-test.yaml up
docker compose -f tools/testing_server/docker-compose-test.yaml down