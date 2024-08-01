#!/usr/bin/bash 

#Command to set up the env to build the chaincode
sudo docker compose -f docker-compose-simple.yaml down
sleep 2
sudo docker compose -f docker-compose-simple.yaml up
