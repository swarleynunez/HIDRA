pragma solidity ^0.6.6;


library NetworkState {
    // Global network state
    struct GlobalState {
        uint64 nodeCount;
        uint64 nextEventId;
    }

    // Node state structure (at a specific time)
    struct NodeState {
        string cpuPercent;
        uint64 memUsage; // In bytes
        string memPercent;
        uint64 diskUsage; // In bytes
        string diskPercent;
        uint64 NetPacketsSent; // Counting all NICs (also the variables below)
        uint64 NetBytesSent;
        uint64 NetPacketsRecv;
        uint64 NetBytesRecv;
    }

    // Functions
}
