#!/usr/bin/env bash

# Settings
WORKDIR="deployment"
NODES=5
MAX_PEERS=4
CHAIN_ID=12345
BOOTNODE="enode://8f9079250bf8a523f46a438560433a962ce2d3f9ba20e99145ac1b52fba27216eef5b6774d6cb9dd83fe564d968ed1725c80e031f988d316be05881419bce6e2@127.0.0.1:30301"
GETH_PORT=30303
RPC_PORT=8551
ETH_BALANCE=100000000000000000000 # 100 ETH
IFACE="lo0"
EVENTS=1

# Prepare deployment environment
mkdir -p "$WORKDIR" && cd "$WORKDIR" || exit
killall bootnode geth hidra >/dev/null 2>&1

# For each node
for i in $(seq 1 $NODES); do
  NODE_DIR="$(pwd)/N$i"
  mkdir -p "N$i/keystore"

  # Create Ethereum account
  if ! [ $(ls -A "$NODE_DIR"/keystore) ]; then
    ../bin/geth --datadir "$NODE_DIR" account new --password secret.txt >/dev/null 2>&1
  fi

  # Remove blockchain and HIDRA logs
  rm -rf "$NODE_DIR"/geth >/dev/null 2>&1
  rm -f "$NODE_DIR"/geth.ipc "$NODE_DIR"/history "$NODE_DIR"/hidra.out "$NODE_DIR"/hidra.err

  # Initialize genesis block
  ../bin/geth --datadir "$NODE_DIR" init genesis.json >/dev/null 2>&1

  # Create HIDRA log file
  touch "$NODE_DIR"/hidra.out
done

# Compile HIDRA
#env GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o ../bin/hidra ../main.go # Linux
env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o ../bin/hidra ../main.go # MacOS
#env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o ../bin/hidra ../main.go # Windows

# Execute bootnode
nohup ../bin/bootnode -nodekey bootnode.key >/dev/null 2>&1 &

# Execute sealer node
SEALER_ACCOUNT=$(../bin/geth --datadir "$(pwd)/N1" --verbosity 0 account list | grep -o "[0-9a-fA-F]\{40\}" | head -1)
nohup ../bin/geth --datadir "$(pwd)/N1" --networkid "$CHAIN_ID" --syncmode full \
  --bootnodes "$BOOTNODE" --maxpeers "$MAX_PEERS" --port "$GETH_PORT" --authrpc.port "$RPC_PORT" \
  --unlock "0x$SEALER_ACCOUNT" --password secret.txt --miner.etherbase "0x$SEALER_ACCOUNT" --mine >/dev/null 2>&1 &
#../bin/hidra monitor $! "$(pwd)/N1"/hidra.out &

# Next ports
GETH_PORT=$((GETH_PORT + 1))
RPC_PORT=$((RPC_PORT + 1))

sleep 3

# For each regular node
for i in $(seq 2 $NODES); do
  NODE_DIR="$(pwd)/N$i"

  # Send ETH from the sealer
  ACCOUNT=$(../bin/geth --datadir "$NODE_DIR" --verbosity 0 account list | grep -o "[0-9a-fA-F]\{40\}" | head -1)
  TXN="eth.sendTransaction({from: \"0x$SEALER_ACCOUNT\", to: \"0x$ACCOUNT\", value: $ETH_BALANCE})"
  ../bin/geth --datadir "$(pwd)/N1" attach --exec "$TXN" >/dev/null 2>&1

  # Execute regular node
  nohup ../bin/geth --datadir "$NODE_DIR" --networkid "$CHAIN_ID" --syncmode full \
    --bootnodes "$BOOTNODE" --maxpeers "$MAX_PEERS" --port "$GETH_PORT" --authrpc.port "$RPC_PORT" \
    --unlock "0x$ACCOUNT" --password secret.txt >/dev/null 2>&1 &
  #../bin/hidra monitor $! "$NODE_DIR"/hidra.out &

  # Next ports
  GETH_PORT=$((GETH_PORT + 1))
  RPC_PORT=$((RPC_PORT + 1))
done

# HIDRA
NODE_PORT=30303
for i in $(seq 1 $NODES); do
  NODE_DIR="$(pwd)/N$i"

  # Configure node directory
  if [[ $OSTYPE == "darwin"* ]]; then
    sed -i '' -e '/ETH_NODE_DIR/d' .env
    sed -i '' -e '4s|^|ETH_NODE_DIR=\"'$NODE_DIR'\"\'$'\n|g' .env
  else
    sed -i -e '/ETH_NODE_DIR/d' .env
    sed -i -e '4s|^|ETH_NODE_DIR=\"'$NODE_DIR'\"\'$'\n|g' .env
  fi

  # Deploy HIDRA smart contract
  if [ $i == 1 ]; then
    ../bin/hidra deploy >/dev/null 2>&1
    sleep 5
  fi

  # Register HIDRA user
  ../bin/hidra register "$NODE_PORT" >/dev/null 2>&1
  sleep 3

  # Run HIDRA
  nohup ../bin/hidra run "$IFACE" >"$NODE_DIR"/hidra.out 2>"$NODE_DIR"/hidra.err &
  #../bin/hidra monitor $! "$NODE_DIR"/hidra.out &
  sleep 1

  # Next port
  NODE_PORT=$((NODE_PORT + 1))
done

sleep 30

# Deploy HIDRA applications/containers
for i in $(seq 1 $EVENTS); do
  sleep 30
  ../bin/hidra application deploy >/dev/null 2>&1
done

# Remove the HIDRA application/container deployed
#../bin/hidra application remove 1
