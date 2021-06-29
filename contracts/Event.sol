pragma solidity ^0.6.6;

library Event {
    struct ClusterEvent {
        Type eType;
        address sender;
        address solver;
        Reply[] replies; // Set of node replies
        uint256 sentAt; // Unix time
        uint256 solvedAt; // Unix time
    }

    struct Type {
        uint64 rcid;
        string metadata; // Encoded dynamic metadata
    }

    struct Reply {
        address replier;
        string nodeState; // Encoded
        address[] voters;
        uint256 repliedAt; // Unix time
    }
}
