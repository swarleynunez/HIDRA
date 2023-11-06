// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

// Distributed Reputation System
library DRS {
    struct ReputableAction {
        int64 limit;
        int64 variation; // Reward or penalty
    }
}
