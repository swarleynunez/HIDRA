pragma solidity ^0.6.6;

library Container {
    // Node containers structure
    struct NodeContainer {
        address host; // Node which runs the container
        string info; // Encoded
        uint64 startedAt; // Unix time
        uint64 finishedAt; // Unix time
    }
}
