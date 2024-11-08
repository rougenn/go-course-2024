#!/bin/bash

docker compose -f ./mystorage.yaml down 
# docker compose -f ./mystorage.yaml down -v --rmi all --remove-orphans
