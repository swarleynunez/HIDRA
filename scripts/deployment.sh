#!/usr/bin/env bash

# Settings
WORKDIR="deployment"
PEERS=10
BOOTNODE="enode://8f9079250bf8a523f46a438560433a962ce2d3f9ba20e99145ac1b52fba27216eef5b6774d6cb9dd83fe564d968ed1725c80e031f988d316be05881419bce6e2@127.0.0.1:30301"
GETH_PORT=30303
RPC_PORT=8551
ETH_BALANCE=1000000000000000000000000 # 1000000 ETH

# Prepare deployment environment
mkdir -p "../$WORKDIR" && cd "../$WORKDIR" || exit
killall bootnode geth hidra >/dev/null 2>&1

# For each peer
for i in $(seq 1 $PEERS); do
  PEER_DIR="$(pwd)/N$i"
  mkdir -p "N$i/keystore"

  # Create Ethereum account
  if ! [ "$(ls -A "$PEER_DIR"/keystore)" ]; then
    ../bin/geth --datadir "$PEER_DIR" account new --password secret.txt >/dev/null 2>&1
  fi

  # Remove blockchain and HIDRA logs
  rm -rf "$PEER_DIR"/geth >/dev/null 2>&1
  rm -f "$PEER_DIR"/geth.ipc "$PEER_DIR"/history "$PEER_DIR"/hidra.out "$PEER_DIR"/hidra.err

  # Initialize genesis block
  ../bin/geth --datadir "$PEER_DIR" init genesis.json >/dev/null 2>&1
done

# Execute bootnode
nohup ../bin/bootnode -nodekey bootnode.key >/dev/null 2>&1 &

# Execute sealer
SEALER_ACCOUNT=$(../bin/geth --datadir "$(pwd)/N1" --verbosity 0 account list | grep -o "[0-9a-fA-F]\{40\}" | head -1)
nohup ../bin/geth --datadir "$(pwd)/N1" --networkid 12345 --bootnodes $BOOTNODE --maxpeers $((PEERS -1)) \
  --port $GETH_PORT --authrpc.port $RPC_PORT --syncmode full --unlock "0x$SEALER_ACCOUNT" --password secret.txt \
  --mine >/dev/null 2>&1 &

# Next ports
GETH_PORT=$((GETH_PORT + 1))
RPC_PORT=$((RPC_PORT + 1))

sleep 3

# For each peer
for i in $(seq 2 $PEERS); do
  PEER_DIR="$(pwd)/N$i"

  # Send ETH from the sealer/faucet
  ACCOUNT=$(../bin/geth --datadir "$PEER_DIR" --verbosity 0 account list | grep -o "[0-9a-fA-F]\{40\}" | head -1)
  TXN="eth.sendTransaction({from: \"0x$SEALER_ACCOUNT\", to: \"0x$ACCOUNT\", value: $ETH_BALANCE})"
  ../bin/geth --datadir "$(pwd)/N1" attach --exec "$TXN" >/dev/null 2>&1

  # Execute regular peer
  nohup ../bin/geth --datadir "$PEER_DIR" --networkid 12345 --bootnodes $BOOTNODE --maxpeers $((PEERS -1)) \
    --port $GETH_PORT --authrpc.port $RPC_PORT --syncmode full --unlock "0x$ACCOUNT" --password secret.txt \
    >/dev/null 2>&1 &

  # Next ports
  GETH_PORT=$((GETH_PORT + 1))
  RPC_PORT=$((RPC_PORT + 1))
done

# Compile HIDRA
# env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o ../bin/hidra ../main.go # Linux
env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o ../bin/hidra ../main.go # MacOS
# env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o ../bin/hidra ../main.go # Windows

# For each peer
for i in $(seq 1 $PEERS); do
  PEER_DIR="$(pwd)/N$i"

  # Configure peer directory
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' -e '/ETH_NODE_DIR/d' .env
    sed -i '' -e '4s|^|ETH_NODE_DIR=\"'$PEER_DIR'\"\'$'\n|g' .env
  else
    sed -i -e '/ETH_NODE_DIR/d' .env
    sed -i -e '4s|^|ETH_NODE_DIR=\"'$PEER_DIR'\"\'$'\n|g' .env
  fi

  # Deploy HIDRA smart contract
  if [ "$i" == 1 ]; then
    ../bin/hidra deploy >/dev/null 2>&1
    sleep 15
  fi

  # Register HIDRA user
  ../bin/hidra register >/dev/null 2>&1
  sleep 1

  # Run HIDRA
  nohup ../bin/hidra run >"$PEER_DIR"/hidra.out 2>"$PEER_DIR"/hidra.err &
  sleep 1
done

# Deploy a HIDRA application/container
../bin/hidra application deploy

# Remove the HIDRA application/container deployed
../bin/hidra application remove 1
