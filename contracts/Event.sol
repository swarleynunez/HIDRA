pragma solidity ^0.6.6;

library Event {
    // Cluster event structure
    struct ClusterEvent {
        string dynType; // Encoded dynamic event type
        address sender;
        uint256 createdAt; // Unix time
        address solver;
        uint256 solvedAt; // Unix time
        ClusterReply[] replies; // Set of node replies
    }

    // Node replies to an event
    struct ClusterReply {
        address replier;
        string nodeState; // Encoded
        uint256 createdAt; // Unix time
        address[] voters;
    }
}
