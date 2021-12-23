// SPDX-License-Identifier: MIT
pragma solidity ^0.6.6;

// Distributed Container Registry
library DCR {
    struct Application {
        address owner; // Node which owns the application
        string info; // Encoded
        uint64[] rcids;
        uint256 registeredAt; // Unix time
        uint256 unregisteredAt; // Unix time
    }

    struct Container {
        uint64 appid; // Application identifier
        string info; // Encoded
        bool autodeployed;
        ContainerInstance[] instances;
        uint256 registeredAt; // Unix time
        uint256 unregisteredAt; // Unix time
    }

    struct ContainerInstance {
        address host; // Node which runs the container
        uint256 startedAt; // Unix time
    }
}
