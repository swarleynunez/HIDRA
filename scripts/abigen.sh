#!/usr/bin/env bash

abigen --abi ./bin/contracts/Controller.abi --bin ./bin/contracts/Controller.bin --type Controller --pkg bindings --out ./core/bindings/controller.go
abigen --abi ./bin/contracts/Faucet.abi --type Faucet --pkg bindings --out ./core/bindings/faucet.go
abigen --abi ./bin/contracts/Node.abi --type Node --pkg bindings --out ./core/bindings/node.go
