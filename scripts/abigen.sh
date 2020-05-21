#!/usr/bin/env bash

abigen --abi ./bin/contracts/Controller.abi --bin ./bin/contracts/Controller.bin --type Controller --pkg contracts --out ./core/contracts/controller.go
abigen --abi ./bin/contracts/Faucet.abi --type Faucet --pkg contracts --out ./core/contracts/faucet.go
abigen --abi ./bin/contracts/Node.abi --type Node --pkg contracts --out ./core/contracts/node.go
