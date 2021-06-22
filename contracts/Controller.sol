pragma solidity ^0.6.6;
pragma experimental ABIEncoderV2;

import "./Faucet.sol";
import "./State.sol";
import "./Node.sol";
import "./Event.sol";
import "./Application.sol";
import "./Container.sol";

contract Controller {
    // Faucet smart contract instance
    Faucet public faucet;

    // Cluster configuration
    State.ClusterConfig public config;
    State.ClusterState public state;

    // Node list
    mapping(address => address) public nodes;

    // Event list
    mapping(uint64 => Event.ClusterEvent) public events;

    // Application and container registries
    mapping(uint64 => Application.ClusterApplication) public apps;
    mapping(uint64 => Container.ClusterContainer) public ctrs;
    uint64[] private activeApps;
    uint64[] private activeCtrs;

    // Main flow events
    event NewEvent(uint64 eventId);
    event RequiredReplies(uint64 eventId);
    event RequiredVotes(uint64 eventId);
    event EventSolved(uint64 eventId);

    // Container events
    event NewContainer(uint64 registryCtrId);
    event ContainerRemoved(uint64 registryCtrId);

    constructor() public {
        // Faucet instance
        faucet = new Faucet();

        // Initialize cluster
        config = State.ClusterConfig(100, 75);
        state = State.ClusterState(0, 0, 0, 0, block.timestamp);
    }

    /////////////////////
    // Public functions//
    /////////////////////
    function registerNode(string memory specs) public {
        require(
            !isNodeRegistered(msg.sender),
            "The node is already registered"
        );

        nodes[msg.sender] = address(new Node(msg.sender, specs));
        // The "new" keyword creates a smart contract

        state.nodeCount++;
    }

    function isNodeRegistered(address nodeAddr) public view returns (bool) {
        if (nodes[nodeAddr] != address(0)) return true;
        return false;
    }

    // Reputable
    function sendEvent(string memory _dynType, string memory _nodeState)
        public
    {
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
        events[state.nextEventId].createdAt = block.timestamp;

        // Create and link the first reply (event sender)
        Event.ClusterReply memory reply;
        reply.replier = msg.sender;
        reply.nodeState = _nodeState;
        reply.createdAt = block.timestamp;
        events[state.nextEventId].replies.push(reply);

        emit NewEvent(state.nextEventId);

        state.nextEventId++;

        // Update the sender reputation
        updateReputation("sendEvent");
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
    function sendReply(uint64 eventId, string memory _nodeState) public {
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
        Event.ClusterReply memory reply;
        reply.replier = msg.sender;
        reply.nodeState = _nodeState;
        reply.createdAt = block.timestamp;
        events[eventId].replies.push(reply);

        if (
            hasRequiredCount(
                config.nodesThld,
                uint64(events[eventId].replies.length)
            )
        ) {
            emit RequiredReplies(eventId);
        }

        // Update the sender reputation
        updateReputation("sendReply");
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
        Event.ClusterReply[] memory replies = events[eventId].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].replier == nodeAddr) return true;
        }

        return false;
    }

    function getEventReplies(uint64 eventId)
        public
        view
        returns (Event.ClusterReply[] memory)
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
            hasRequiredCount(
                config.nodesThld,
                uint64(events[eventId].replies.length)
            ),
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
        Event.ClusterReply[] memory replies = events[eventId].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].replier == candidateAddr) {
                // Vote candidate (storage)
                events[eventId].replies[i].voters.push(msg.sender);

                uint64 votes = uint64(replies[i].voters.length + 1);
                if (hasRequiredCount(config.votesThld, votes)) {
                    emit RequiredVotes(eventId);
                }

                break;
            }
        }

        // Update the sender reputation
        updateReputation("voteSolver");
    }

    function hasAlreadyVoted(uint64 eventId, address nodeAddr)
        public
        view
        returns (bool)
    {
        Event.ClusterReply[] memory replies = events[eventId].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            for (uint64 j = 0; j < replies[i].voters.length; j++) {
                if (replies[i].voters[j] == nodeAddr) return true;
            }
        }

        return false;
    }

    // Reputable
    function solveEvent(uint64 eventId) public {
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
        events[eventId].solvedAt = block.timestamp;

        emit EventSolved(eventId);

        // Update the sender reputation
        updateReputation("solveEvent");
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
        Event.ClusterReply[] memory replies = events[eventId].replies;
        uint64 votes;

        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].replier == nodeAddr) {
                votes = uint64(replies[i].voters.length);
                break;
            }
        }

        if (hasRequiredCount(config.votesThld, votes)) return true;
        return false;
    }

    // Reputable
    function registerApplication(
        string memory _info,
        Container.ClusterContainer[] memory _ctrs
    ) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("registerApp") // Required reputation to register applications
            ),
            "The node has not enough reputation"
        );

        // Record a new application
        apps[state.nextAppId].owner = msg.sender;
        apps[state.nextAppId].info = _info;
        apps[state.nextAppId].createdAt = block.timestamp;

        // Set application as active
        activeApps.push(state.nextAppId);

        // Update the sender reputation
        updateReputation("registerApp");

        // Record containers
        for (uint64 i = 0; i < _ctrs.length; i++) {
            require(
                hasNodeReputation(
                    msg.sender,
                    faucet.getActionLimit("registerCtr") // Required reputation to register containers
                ),
                "The node has not enough reputation"
            );

            recordContainer(state.nextAppId, _ctrs[i].info);
        }

        state.nextAppId++;
    }

    // Reputable
    function registerContainer(uint64 appId, string memory info) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("registerCtr") // Required reputation to register containers
            ),
            "The node has not enough reputation"
        );
        require(existApplication(appId), "The application does not exist");
        require(
            isApplicationOwner(msg.sender, appId),
            "The node is not the application owner"
        );
        require(isApplicationActive(appId), "The application is not active");

        recordContainer(appId, info);
    }

    function existApplication(uint64 appId) public view returns (bool) {
        if (apps[appId].owner != address(0)) return true;
        return false;
    }

    function isApplicationOwner(address nodeAddr, uint64 appId)
        public
        view
        returns (bool)
    {
        if (apps[appId].owner == nodeAddr) return true;
        return false;
    }

    function isApplicationActive(uint64 appId) public view returns (bool) {
        for (uint64 i = 0; i < activeApps.length; i++) {
            if (activeApps[i] == appId) return true;
        }

        return false;
    }

    // Reputable
    function unregisterApplication(uint64 _appId) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("unregisterApp") // Required reputation to unregister applications
            ),
            "The node has not enough reputation"
        );
        require(existApplication(_appId), "The application does not exist");
        require(
            isApplicationOwner(msg.sender, _appId),
            "The node is not the application owner"
        );
        require(isApplicationActive(_appId), "The application is not active");

        // Set application removal time and deactivate it
        apps[_appId].deletedAt = block.timestamp;
        deactivateApplicationById(_appId);

        // Update the sender reputation
        updateReputation("unregisterApp");

        // Deactivate all application containers
        uint64 i = 0;
        while (i < activeCtrs.length) {
            if (ctrs[activeCtrs[i]].appId == _appId) {
                require(
                    hasNodeReputation(
                        msg.sender,
                        faucet.getActionLimit("unregisterCtr") // Required reputation to unregister containers
                    ),
                    "The node has not enough reputation"
                );

                // Set container removal time and deactivate it
                ctrs[activeCtrs[i]].finishedAt = block.timestamp;
                activeCtrs[i] = activeCtrs[activeCtrs.length - 1];
                activeCtrs.pop();

                emit ContainerRemoved(activeCtrs[i]);

                // Update the sender reputation
                updateReputation("unregisterCtr");
            } else i++;
        }
    }

    // Reputable
    function unregisterContainer(uint64 registryCtrId) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("unregisterCtr") // Required reputation to unregister containers
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
        ctrs[registryCtrId].finishedAt = block.timestamp;
        deactivateContainerById(registryCtrId);

        emit ContainerRemoved(registryCtrId);

        // Update the sender reputation
        updateReputation("unregisterCtr");
    }

    function existContainer(uint64 registryCtrId) public view returns (bool) {
        if (ctrs[registryCtrId].host != address(0)) return true;
        return false;
    }

    function isContainerHost(address nodeAddr, uint64 registryCtrId)
        public
        view
        returns (bool)
    {
        if (ctrs[registryCtrId].host == nodeAddr) return true;
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

    function getActiveApplications() public view returns (uint64[] memory) {
        return activeApps;
    }

    function getActiveContainers() public view returns (uint64[] memory) {
        return activeCtrs;
    }

    //////////////////////
    // Private functions//
    //////////////////////
    function hasRequiredCount(uint8 threshold, uint64 count)
        private
        view
        returns (bool)
    {
        uint64 required = (state.nodeCount * threshold) / 100;
        if (required > 0 && count >= required) return true;
        return false;
    }

    function updateReputation(string memory action) private {
        // Node instance
        Node nc = Node(nodes[msg.sender]);

        nc.setVariation(faucet.getActionVariation(action));
    }

    function recordContainer(uint64 _appId, string memory _info) private {
        ctrs[state.nextRegistryCtrId].appId = _appId;
        ctrs[state.nextRegistryCtrId].host = msg.sender;
        ctrs[state.nextRegistryCtrId].info = _info;
        ctrs[state.nextRegistryCtrId].startedAt = block.timestamp;

        // Set container as active
        activeCtrs.push(state.nextRegistryCtrId);

        emit NewContainer(state.nextRegistryCtrId);

        state.nextRegistryCtrId++;

        // Update the sender reputation
        updateReputation("registerCtr");
    }

    function deactivateApplicationById(uint64 appId) private {
        for (uint64 i = 0; i < activeApps.length; i++) {
            if (activeApps[i] == appId) {
                activeApps[i] = activeApps[activeApps.length - 1];
                activeApps.pop();
                break;
            }
        }
    }

    function deactivateContainerById(uint64 registryCtrId) private {
        for (uint64 i = 0; i < activeCtrs.length; i++) {
            if (activeCtrs[i] == registryCtrId) {
                activeCtrs[i] = activeCtrs[activeCtrs.length - 1];
                activeCtrs.pop();
                break;
            }
        }
    }
}
