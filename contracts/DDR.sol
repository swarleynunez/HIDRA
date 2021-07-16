// SPDX-License-Identifier: MIT
pragma solidity ^0.6.6;

// Distributed Device Registry
library DDR {
    struct NodeData {
        address controller; // Controller contract address
        string specs; // Encoded
        int64 reputation;
        uint256 registeredAt; // Unix time
    }
}
