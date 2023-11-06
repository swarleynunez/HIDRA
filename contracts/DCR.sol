// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

// Distributed Container Registry
library DCR {
    struct Application {
        address owner; // Node which owns the application
        string info; // Encoded
        uint64[] rcids;
        uint registeredAt; // Unix time
        uint unregisteredAt; // Unix time
    }

    struct Container {
        uint64 appid; // Application identifier
        string info; // Encoded
        bool autodeployed;
        ContainerInstance[] instances;
        uint registeredAt; // Unix time
        uint unregisteredAt; // Unix time
    }

    struct ContainerInstance {
        address host; // Node which runs the container
        uint startedAt; // Unix time
    }
}
