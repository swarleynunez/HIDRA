// SPDX-License-Identifier: MIT
pragma solidity ^0.6.6;

// Distributed Event Logger
library DEL {
    struct Event {
        string eType; // Encoded event type
        address sender;
        address solver;
        EventReply[] replies; // Set of node replies
        uint64 rcid; // Optional, depending on the event type
        uint256 sentAt; // Unix time
        uint256 solvedAt; // Unix time
    }

    struct EventReply {
        address replier;
        string nodeState; // Encoded
        address[] voters;
        uint256 repliedAt; // Unix time
    }
}
