pragma solidity ^0.6.6;

library State {
    // Global cluster state
    struct ClusterState {
        uint64 nodeCount;
        uint64 nextEventId; // Starting at 1 (0 for empty values)
        uint64 nextAppId; // Starting at 1 (0 for empty values)
        uint64 nextRegistryCtrId; // Starting at 1 (0 for empty values)
        uint256 initDate; // Unix time
    }

    // Global cluster parameters
    struct ClusterConfig {
        uint8 nodesThld; // Percentage threshold to calculate required nodes (0-100)
        uint8 votesThld; // Percentage threshold to calculate required votes (0-100)
    }
}
