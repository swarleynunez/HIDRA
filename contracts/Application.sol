pragma solidity ^0.6.6;

library Application {
    struct ClusterApplication {
        address owner; // Node which owns the application
        string info; // Encoded
        uint256 registeredAt; // Unix time
        uint256 unregisteredAt; // Unix time
    }
}
