pragma solidity ^0.6.6;

library Container {
    // Cluster container structure
    struct ClusterContainer {
        uint64 appId; // Application identificator
        address host; // Node which runs the container
        string info; // Encoded
        uint256 startedAt; // Unix time
        uint256 finishedAt; // Unix time
    }
}
