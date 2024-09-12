#!/bin/sh
export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=$PWD/../config/

# Verify if peer binary exists in the path
if ! command -v peer &> /dev/null; then
  echo "peer binary could not be found. Please ensure it is installed and available in your PATH."
  # exit 1
fi

#Setting up one of the Orgs Peer Env Variables
# Environment variables for Org1
 
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_LOCALMSPID="Org1MSP"
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp
export CORE_PEER_ADDRESS=localhost:7051
