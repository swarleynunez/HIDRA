// SPDX-License-Identifier: MIT
pragma solidity ^0.6.6;

// Distributed Reputation System
library DRS {
    struct ReputableAction {
        int64 limit;
        int64 variation; // Reward or penalty
    }
}
