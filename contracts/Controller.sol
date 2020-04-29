pragma solidity ^0.6.6;

import "./Faucet.sol";
import "./Node.sol";
import "./EventHandler.sol";


contract Controller {
    // Faucet smart contract instance
    Faucet private faucet;

    // Global network state
    NetworkState.GlobalState private state;

    // Node list
    mapping(address => address) private nodes;

    // Event list
    mapping(uint64 => EventHandler.Event) private events;

    // Solidity events
    event SendEvent(uint64 id, uint64 typeId, address sender, uint64 createdAt);

    // Constructor
    constructor() public {
        // Deploy the faucet
        faucet = new Faucet();

        // Initialize global network state
        state = NetworkState.GlobalState(0, 1);
    }

    // Functions
    function registerNode() public {
        require(
            !isNodeRegistered(msg.sender),
            "The node is already registered"
        );

        nodes[msg.sender] = address(new Node(msg.sender)); // The "new" keyword creates a smart contract

        state.nodeCount++;
    }

    function isNodeRegistered(address nodeAddr) public view returns (bool) {
        if (nodes[nodeAddr] != address(0)) return true;
        return false;
    }

    function sendEvent(uint64 _typeId, uint64 _createdAt) public {
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("sendEvent") // Required reputation to send events
            ),
            "The node has not enough reputation"
        );

        // Create and save the new event
        events[state.nextEventId] = EventHandler.Event(
            _typeId,
            msg.sender,
            _createdAt,
            address(0), // Empty at the beginning
            0 // Empty at the beginning
        );

        // Emits a solidity event
        emit SendEvent(state.nextEventId, _typeId, msg.sender, _createdAt);

        state.nextEventId++;
    }

    function hasNodeReputation(address nodeAddr, uint64 reputation)
        public
        view
        returns (bool)
    {
        require(isNodeRegistered(nodeAddr), "The node is not registered");

        // Get node reputation
        Node nodeContract = Node(nodes[nodeAddr]);

        if (nodeContract.getReputation() >= reputation) return true;
        return false;
    }
}
