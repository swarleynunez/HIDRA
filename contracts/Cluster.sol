// SPDX-License-Identifier: MIT
pragma solidity ^0.6.6;

library Cluster {
    struct State {
        uint64 nodeCount;
        uint64 nextEventId; // Starting at 1 (0 for empty values)
        uint64 nextAppId; // Starting at 1 (0 for empty values)
        uint64 nextCtrId; // Starting at 1 (0 for empty values)
        uint256 deployedAt; // Unix time
    }

    struct Config {
        int64 initNodeRep;
        uint8 nodesThld; // Percentage threshold to calculate required nodes (0-100)
        uint8 votesThld; // Percentage threshold to calculate required votes (0-100)
    }
}
