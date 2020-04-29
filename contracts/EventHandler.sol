pragma solidity ^0.6.6;

import "./NetworkState.sol";


library EventHandler {
    // Node events structure
    struct Event {
        uint64 typeId; // "Client" constant list
        address sender;
        uint64 createdAt; // Unix time
        address solver;
        uint64 solvedAt; // Unix time
        mapping(address => Reply) replies; // Replies by each node address
    }

    // Network nodes replies to an event
    struct Reply {
        NetworkState.NodeState state;
        uint64 createdAt; // Unix time
    }

    // Functions
}
