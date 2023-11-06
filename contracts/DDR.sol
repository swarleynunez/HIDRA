// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

// Distributed Device Registry
library DDR {
    struct NodeData {
        address controller; // Controller contract address
        string specs; // Encoded
        //int64 reputation;
        uint registeredAt; // Unix time
    }
}
