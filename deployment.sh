#!/usr/bin/env bash
docker-compose build && docker-compose push
docker stack deploy --compose-file=docker-compose.yml --with-registry-auth benzinpriser
