// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

// Distributed Event Logger
library DEL {
    struct Event {
        string eType; // Encoded event type
        address sender;
        address solver;
        uint64 rcid; // Optional, depending on the event type
        EventReply[] replies; // Set of node replies
        mapping(address => uint64) votes;
        mapping(address => bool) repliers;
        mapping(address => bool) voters;
        bool hasRequiredReplies;
        bool hasRequiredVotes;
        uint sentAt; // Unix time
        uint solvedAt; // Unix time
    }

    struct EventReply {
        address replier;
        ReputationScore[] repScores;
        mapping(address => bool) reputedNodes;
        uint repliedAt; // Unix time
    }

    struct ReputationScore {
        address node;
        string score;
    }
}
