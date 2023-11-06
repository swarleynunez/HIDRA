// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

library Cluster {
    struct State {
        uint64 nodeCount;
        uint64 nextEventId; // Starting at 1 (0 for empty values)
        uint64 nextAppId; // Starting at 1 (0 for empty values)
        uint64 nextCtrId; // Starting at 1 (0 for empty values)
        uint deployedAt; // Unix time
    }

    struct Config {
        //int64 initNodeRep;
        uint8 nodesTh; // Percentage threshold to calculate required nodes (0-100)
        uint8 votesTh; // Percentage threshold to calculate required votes (0-100)
        uint64 maxRepScores; // Maximum number of reputation scores to be counted for each node
    }
}
