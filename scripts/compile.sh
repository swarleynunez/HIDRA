#!/usr/bin/env bash

# Compile Solidity contracts (solc 0.8.21)
./bin/solc --evm-version paris --optimize --optimize-runs=1000000 --abi --bin -o bin/contracts --overwrite contracts/Controller.sol >/dev/null 2>&1

# Generate Go bindings
./bin/abigen --abi bin/contracts/Controller.abi --bin bin/contracts/Controller.bin --type Controller --pkg bindings --out core/bindings/controller.go
./bin/abigen --abi bin/contracts/Faucet.abi --type Faucet --pkg bindings --out core/bindings/faucet.go
./bin/abigen --abi bin/contracts/Node.abi --type Node --pkg bindings --out core/bindings/node.go
