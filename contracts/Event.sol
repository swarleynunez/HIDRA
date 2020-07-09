pragma solidity ^0.6.6;

library Event {
    // Node events structure
    struct NodeEvent {
        string dynType; // Encoded dynamic event type
        address sender;
        uint64 createdAt; // Unix time
        address solver;
        uint64 solvedAt; // Unix time
        Reply[] replies; // Set of node replies
    }

    // Network nodes replies to an event
    struct Reply {
        address sender;
        string nodeState; // Encoded
        uint64 createdAt; // Unix time
        address[] voters;
    }
}
