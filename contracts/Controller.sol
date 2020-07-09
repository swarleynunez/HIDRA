pragma solidity ^0.6.6;
pragma experimental ABIEncoderV2;

import "./Faucet.sol";
import "./State.sol";
import "./Node.sol";
import "./Event.sol";
import "./Container.sol";

contract Controller {
    // Faucet smart contract instance
    Faucet public faucet;

    // Global network state
    State.GlobalState public state;

    // Node list
    mapping(address => address) public nodes;

    // Event list
    mapping(uint64 => Event.NodeEvent) public events;

    // Container registry
    mapping(uint64 => Container.NodeContainer) public containers;
    uint64[] private activeCtrs;

    // Main flow events
    event NewEvent(uint64 eventId);
    event RequiredReplies(uint64 eventId);
    event RequiredVotes(uint64 eventId, address solver);
    event EventSolved(uint64 eventId, address sender);

    // Container events
    event NewContainer(uint64 registryCtrId, string ctrId, address host);
    event ContainerRemoved(uint64 registryCtrId);

    // Constructor
    constructor() public {
        // Faucet instance
        faucet = new Faucet();

        // Initialize global network state
        state = State.GlobalState(0, 0, 0);
    }

    // Functions
    function registerNode(string memory specs) public {
        require(
            !isNodeRegistered(msg.sender),
            "The node is already registered"
        );

        nodes[msg.sender] = address(new Node(msg.sender, specs)); // The "new" keyword creates a smart contract

        state.nodeCount++;
    }

    function isNodeRegistered(address nodeAddr) public view returns (bool) {
        if (nodes[nodeAddr] != address(0)) return true;
        return false;
    }

    // Reputable
    function sendEvent(
        string memory _dynType,
        uint64 _createdAt,
        string memory _nodeState
    ) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("sendEvent") // Required reputation to send events
            ),
            "The node has not enough reputation"
        );

        // Create a new event
        events[state.nextEventId].dynType = _dynType;
        events[state.nextEventId].sender = msg.sender;
        events[state.nextEventId].createdAt = _createdAt;

        // Create and link the first reply (event sender)
        Event.Reply memory reply;
        reply.sender = msg.sender;
        reply.nodeState = _nodeState;
        reply.createdAt = _createdAt;
        events[state.nextEventId].replies.push(reply);

        emit NewEvent(state.nextEventId);

        state.nextEventId++;

        // Update the sender reputation
        updateReputation(msg.sender, "sendEvent");
    }

    function hasNodeReputation(address nodeAddr, int64 reputation)
        public
        view
        returns (bool)
    {
        // Node instance
        Node nc = Node(nodes[nodeAddr]);
        if (nc.getReputation() >= reputation) return true;

        return false;
    }

    // Reputable
    function sendReply(
        uint64 eventId,
        string memory _nodeState,
        uint64 _createdAt
    ) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("sendReply") // Required reputation to send replies
            ),
            "The node has not enough reputation"
        );
        require(existEvent(eventId), "The event does not exist");
        require(!isEventSolved(eventId), "The event is solved");
        require(
            !hasAlreadyReplied(eventId, msg.sender),
            "The node has already replied the event"
        );

        // Create and link the reply to its event
        Event.Reply memory reply;
        reply.sender = msg.sender;
        reply.nodeState = _nodeState;
        reply.createdAt = _createdAt;
        events[eventId].replies.push(reply);

        if (hasRequiredCount(uint64(events[eventId].replies.length))) {
            emit RequiredReplies(eventId);
        }

        // Update the sender reputation
        updateReputation(msg.sender, "sendReply");
    }

    function existEvent(uint64 eventId) public view returns (bool) {
        if (events[eventId].sender != address(0)) return true;
        return false;
    }

    function hasAlreadyReplied(uint64 eventId, address nodeAddr)
        public
        view
        returns (bool)
    {
        Event.Reply[] memory replies = events[eventId].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].sender == nodeAddr) return true;
        }

        return false;
    }

    function getEventReplies(uint64 eventId)
        public
        view
        returns (Event.Reply[] memory)
    {
        return events[eventId].replies;
    }

    // Reputable
    function voteSolver(uint64 eventId, address candidateAddr) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("voteSolver") // Required reputation to vote solvers
            ),
            "The node has not enough reputation"
        );
        require(existEvent(eventId), "The event does not exist");
        require(!isEventSolved(eventId), "The event is solved");
        require(
            hasRequiredCount(uint64(events[eventId].replies.length)),
            "The event does not have the required replies"
        );
        require(
            hasAlreadyReplied(eventId, candidateAddr),
            "The candidate has not replied the event"
        );
        require(
            !hasAlreadyVoted(eventId, msg.sender),
            "The node has already voted a solver"
        );

        // Search candidate
        Event.Reply[] memory replies = events[eventId].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].sender == candidateAddr) {
                // Vote candidate (storage)
                events[eventId].replies[i].voters.push(msg.sender);

                uint64 votes = uint64(replies[i].voters.length + 1);
                if (hasRequiredCount(votes)) {
                    emit RequiredVotes(eventId, candidateAddr);
                }

                break;
            }
        }

        // Update the sender reputation
        updateReputation(msg.sender, "voteSolver");
    }

    function hasAlreadyVoted(uint64 eventId, address nodeAddr)
        public
        view
        returns (bool)
    {
        Event.Reply[] memory replies = events[eventId].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            for (uint64 j = 0; j < replies[i].voters.length; j++) {
                if (replies[i].voters[j] == nodeAddr) return true;
            }
        }

        return false;
    }

    // Reputable
    function solveEvent(uint64 eventId, uint64 _solvedAt) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("solveEvent") // Required reputation to solve events
            ),
            "The node has not enough reputation"
        );
        require(existEvent(eventId), "The event does not exist");
        require(!isEventSolved(eventId), "The event is already solved");
        require(
            canSolveEvent(eventId, msg.sender),
            "The node can not solve the event"
        );

        // Update event solver info
        events[eventId].solver = msg.sender;
        events[eventId].solvedAt = _solvedAt;

        emit EventSolved(eventId, events[eventId].sender);

        // Update the sender reputation
        updateReputation(msg.sender, "solveEvent");
    }

    function isEventSolved(uint64 eventId) public view returns (bool) {
        if (events[eventId].solver != address(0)) return true;
        return false;
    }

    function canSolveEvent(uint64 eventId, address nodeAddr)
        public
        view
        returns (bool)
    {
        Event.Reply[] memory replies = events[eventId].replies;
        uint64 votes;

        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].sender == nodeAddr) {
                votes = uint64(replies[i].voters.length);
                break;
            }
        }

        if (hasRequiredCount(votes)) return true;
        return false;
    }

    // Reputable
    function recordContainer(
        string memory _info,
        uint64 _startedAt,
        string memory ctrId
    ) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("recordContainer") // Required reputation to register containers
            ),
            "The node has not enough reputation"
        );

        // Register a new container
        containers[state.nextRegistryCtrId].host = msg.sender;
        containers[state.nextRegistryCtrId].info = _info;
        containers[state.nextRegistryCtrId].startedAt = _startedAt;

        // Set container as active
        activeCtrs.push(state.nextRegistryCtrId);

        emit NewContainer(state.nextRegistryCtrId, ctrId, msg.sender);

        state.nextRegistryCtrId++;

        // Update the sender reputation
        updateReputation(msg.sender, "recordContainer");
    }

    // Reputable
    function removeContainer(uint64 registryCtrId, uint64 _finishedAt) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("removeContainer") // Required reputation to remove containers
            ),
            "The node has not enough reputation"
        );
        require(existContainer(registryCtrId), "The container does not exist");
        require(
            isContainerHost(msg.sender, registryCtrId),
            "The node is not the container host"
        );
        require(
            isContainerActive(registryCtrId),
            "The container is not active"
        );

        // Set container removal time and deactivate it
        containers[registryCtrId].finishedAt = _finishedAt;
        deactivateContainer(registryCtrId);

        emit ContainerRemoved(registryCtrId);

        // Update the sender reputation
        updateReputation(msg.sender, "removeContainer");
    }

    function existContainer(uint64 registryCtrId) public view returns (bool) {
        if (containers[registryCtrId].host != address(0)) return true;
        return false;
    }

    function isContainerActive(uint64 registryCtrId)
        public
        view
        returns (bool)
    {
        for (uint64 i = 0; i < activeCtrs.length; i++) {
            if (activeCtrs[i] == registryCtrId) return true;
        }

        return false;
    }

    function isContainerHost(address nodeAddr, uint64 registryCtrId)
        public
        view
        returns (bool)
    {
        if (containers[registryCtrId].host == nodeAddr) return true;
        return false;
    }

    function getActiveContainers() public view returns (uint64[] memory) {
        return activeCtrs;
    }

    // Private functions
    function hasRequiredCount(uint64 count) private view returns (bool) {
        // TODO (51%, 66%)
        uint64 required = state.nodeCount;
        if (required != 0 && count == required) return true;
        return false;
    }

    function updateReputation(address nodeAddr, string memory action) private {
        // Node instance
        Node nc = Node(nodes[nodeAddr]);

        nc.setVariation(faucet.getActionVariation(action));
    }

    function deactivateContainer(uint64 registryCtrId) private {
        for (uint64 i = 0; i < activeCtrs.length; i++) {
            if (activeCtrs[i] == registryCtrId) {
                activeCtrs[i] = activeCtrs[activeCtrs.length - 1];
                activeCtrs.pop();
            }
        }
    }
}
