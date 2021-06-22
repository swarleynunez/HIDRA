pragma solidity ^0.6.6;

library State {
    // Global cluster parameters
    struct ClusterConfig {
        uint8 nodesThld; // Percentage threshold to calculate required nodes (0-100)
        uint8 votesThld; // Percentage threshold to calculate required votes (0-100)
    }

    // Global cluster state
    struct ClusterState {
        uint64 nodeCount;
        uint64 nextEventId;
        uint64 nextAppId;
        uint64 nextRegistryCtrId;
        uint256 initDate; // Unix time
    }
}
