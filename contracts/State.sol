pragma solidity ^0.6.6;


library State {
    // Global network state
    struct GlobalState {
        uint64 nodeCount;
        uint64 nextEventId;
    }
}
