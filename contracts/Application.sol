pragma solidity ^0.6.6;

library Application {
    // Cluster application structure
    struct ClusterApplication {
        address owner; // Node which owns the application
        string info; // Encoded
        uint256 createdAt; // Unix time
        uint256 deletedAt; // Unix time
    }
}
