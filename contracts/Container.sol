pragma solidity ^0.6.6;

library Container {
    struct ClusterContainer {
        uint64 appId; // Application identificator
        string info; // Encoded
        Instance[] instances;
        uint256 registeredAt; // Unix time
        uint256 unregisteredAt; // Unix time
    }

    struct Instance {
        address host; // Node which runs the container
        uint256 startedAt; // Unix time
    }
}
