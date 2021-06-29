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
    event NewEvent(uint64 eid);
    event RequiredReplies(uint64 eid);
    event RequiredVotes(uint64 eid);
    event EventSolved(uint64 eid);

    // Container events
    event ContainerRegistered(uint64 rcid);
    event ContainerUpdated(uint64 rcid);
    event ContainerUnregistered(uint64 rcid);

    constructor() public {
        // Faucet instance
        faucet = new Faucet();

        // Initialize cluster
        config = State.ClusterConfig(100, 100);
        state = State.ClusterState(0, 1, 1, 1, block.timestamp);
    }

    //////////////////////
    // Public functions //
    //////////////////////
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

    function isApplicationOwner(uint64 appId, address nodeAddr)
        public
        view
        returns (bool)
    {
        if (apps[appId].owner == nodeAddr) return true;
        return false;
    }

    function isContainerHost(uint64 rcid, address nodeAddr)
        public
        view
        returns (bool)
    {
        Container.Instance[] memory insts = ctrs[rcid].instances;
        if (insts.length > 0 && insts[insts.length - 1].host == nodeAddr)
            return true;
        return false;
    }

    function existEvent(uint64 eid) public view returns (bool) {
        if (events[eid].sender != address(0)) return true;
        return false;
    }

    function hasAlreadyReplied(uint64 eid, address nodeAddr)
        public
        view
        returns (bool)
    {
        Event.Reply[] memory replies = events[eid].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].replier == nodeAddr) return true;
        }

        return false;
    }

    function hasAlreadyVoted(uint64 eid, address nodeAddr)
        public
        view
        returns (bool)
    {
        Event.Reply[] memory replies = events[eid].replies;

        for (uint64 i = 0; i < replies.length; i++) {
            for (uint64 j = 0; j < replies[i].voters.length; j++) {
                if (replies[i].voters[j] == nodeAddr) return true;
            }
        }

        return false;
    }

    function isEventSolved(uint64 eid) public view returns (bool) {
        if (events[eid].solvedAt != 0) return true;
        return false;
    }

    function canSolveEvent(uint64 eid, address nodeAddr)
        public
        view
        returns (bool)
    {
        if (events[eid].solver == nodeAddr) return true;
        return false;
    }

    function existApplication(uint64 appId) public view returns (bool) {
        if (apps[appId].owner != address(0)) return true;
        return false;
    }

    function isApplicationUnregistered(uint64 appId)
        public
        view
        returns (bool)
    {
        if (apps[appId].unregisteredAt != 0) return true;
        return false;
    }

    function existContainer(uint64 rcid) public view returns (bool) {
        if (ctrs[rcid].appId != 0) return true;
        return false;
    }

    function isContainerUnregistered(uint64 rcid) public view returns (bool) {
        if (ctrs[rcid].unregisteredAt != 0) return true;
        return false;
    }

    function isContainerActive(uint64 rcid) public view returns (bool) {
        for (uint64 i = 0; i < activeCtrs.length; i++) {
            if (activeCtrs[i] == rcid) return true;
        }

        return false;
    }

    /////////////////////////
    // Reputable functions //
    /////////////////////////
    function sendEvent(
        uint64 _rcid,
        string memory _metadata,
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
        events[state.nextEventId].eType.metadata = _metadata;
        events[state.nextEventId].sender = msg.sender;
        events[state.nextEventId].sentAt = block.timestamp;

        if (_rcid > 0) {
            require(existContainer(_rcid), "The container does not exist");
            require(
                isApplicationOwner(ctrs[_rcid].appId, msg.sender) ||
                    isContainerHost(_rcid, msg.sender),
                "The node is neither the container owner nor the host"
            );

            // Link an existing container to the new event
            events[state.nextEventId].eType.rcid = _rcid;
        }

        // Link the first reply (event sender)
        Event.Reply memory reply;
        reply.replier = msg.sender;
        reply.nodeState = _nodeState;
        reply.repliedAt = block.timestamp;
        events[state.nextEventId].replies.push(reply);

        emit NewEvent(state.nextEventId);

        state.nextEventId++;

        // Update the sender reputation
        updateReputation("sendEvent");
    }

    function sendReply(uint64 eid, string memory _nodeState) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("sendReply") // Required reputation to send replies
            ),
            "The node has not enough reputation"
        );
        require(existEvent(eid), "The event does not exist");
        require(!isEventSolved(eid), "The event is solved");
        require(
            !hasAlreadyReplied(eid, msg.sender),
            "The node has already replied the event"
        );

        // Create and link the reply to its event
        Event.Reply memory reply;
        reply.replier = msg.sender;
        reply.nodeState = _nodeState;
        reply.repliedAt = block.timestamp;
        events[eid].replies.push(reply);

        if (
            hasRequiredCount(
                config.nodesThld,
                uint64(events[eid].replies.length)
            )
        ) {
            emit RequiredReplies(eid);
        }

        // Update the sender reputation
        updateReputation("sendReply");
    }

    function voteSolver(uint64 eid, address candidateAddr) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("voteSolver") // Required reputation to vote solvers
            ),
            "The node has not enough reputation"
        );
        require(existEvent(eid), "The event does not exist");
        require(!isEventSolved(eid), "The event is solved");
        require(
            hasRequiredCount(
                config.nodesThld,
                uint64(events[eid].replies.length)
            ),
            "The event does not have the required replies"
        );
        require(
            hasAlreadyReplied(eid, candidateAddr),
            "The candidate has not replied the event yet"
        );
        require(
            !hasAlreadyVoted(eid, msg.sender),
            "The node has already voted a solver"
        );

        Event.Reply[] memory replies = events[eid].replies;

        // Search candidate
        for (uint64 i = 0; i < replies.length; i++) {
            if (replies[i].replier == candidateAddr) {
                // Vote candidate
                events[eid].replies[i].voters.push(msg.sender);

                uint64 votes = uint64(replies[i].voters.length + 1);
                if (hasRequiredCount(config.votesThld, votes)) {
                    // Set event solver
                    events[eid].solver = candidateAddr;

                    // Add a new container instance (if applicable)
                    uint64 rcid = events[eid].eType.rcid;
                    if (rcid > 0) {
                        if (!isContainerHost(rcid, candidateAddr)) {
                            Container.Instance memory inst;
                            inst.host = candidateAddr;
                            ctrs[rcid].instances.push(inst);
                        }
                    }

                    emit RequiredVotes(eid);
                }

                break;
            }
        }

        // Update the sender reputation
        updateReputation("voteSolver");
    }

    function solveEvent(uint64 eid) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("solveEvent") // Required reputation to solve events
            ),
            "The node has not enough reputation"
        );
        require(existEvent(eid), "The event does not exist");
        require(!isEventSolved(eid), "The event is already solved");
        require(
            canSolveEvent(eid, msg.sender),
            "The node can not solve the event"
        );

        // Solve event
        events[eid].solvedAt = block.timestamp;

        // Activate container (if applicable)
        uint64 rcid = events[eid].eType.rcid;
        if (rcid > 0) {
            // Set container start time and activate it
            Container.Instance[] storage insts = ctrs[rcid].instances;
            if (insts[insts.length - 1].startedAt == 0)
                insts[insts.length - 1].startedAt = block.timestamp;
            if (!isContainerActive(rcid)) activeCtrs.push(rcid);
        }

        emit EventSolved(eid);

        // Update the sender reputation
        updateReputation("solveEvent");
    }

    function registerApplication(
        string memory _info,
        Container.ClusterContainer[] memory _ctrs,
        bool autodeploy
    ) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("registerApp") // Required reputation to register applications
            ),
            "The node has not enough reputation"
        );

        // Register a new application
        apps[state.nextAppId].owner = msg.sender;
        apps[state.nextAppId].info = _info;
        apps[state.nextAppId].registeredAt = block.timestamp;

        // Set application as active
        activeApps.push(state.nextAppId);

        // Update the sender reputation
        updateReputation("registerApp");

        // Register all application containers
        for (uint64 i = 0; i < _ctrs.length; i++) {
            require(
                hasNodeReputation(
                    msg.sender,
                    faucet.getActionLimit("registerCtr") // Required reputation to register containers
                ),
                "The node has not enough reputation"
            );

            recordContainer(state.nextAppId, _ctrs[i].info, autodeploy);
        }

        state.nextAppId++;
    }

    function registerContainer(
        uint64 appId,
        string memory info,
        bool autodeploy
    ) public {
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
            isApplicationOwner(appId, msg.sender),
            "The node is not the application owner"
        );
        require(
            !isApplicationUnregistered(appId),
            "The application was unregistered"
        );

        recordContainer(appId, info, autodeploy);
    }

    function activateContainers(uint64[] memory rcids) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(rcids.length > 0, "Container identifiers are not provided");

        // Activate each container
        for (uint64 i = 0; i < rcids.length; i++) {
            require(
                hasNodeReputation(
                    msg.sender,
                    faucet.getActionLimit("activateCtr") // Required reputation to activate containers
                ),
                "The node has not enough reputation"
            );
            require(existContainer(rcids[i]), "The container does not exist");
            require(
                isApplicationOwner(ctrs[rcids[i]].appId, msg.sender),
                "The node is not the container owner"
            );
            require(
                isContainerHost(rcids[i], msg.sender),
                "The node is not the container host"
            );
            require(
                !isContainerUnregistered(rcids[i]),
                "The container was unregistered"
            );
            require(
                !isContainerActive(rcids[i]),
                "The container is already activated"
            );

            // Set container start time and activate it
            Container.Instance[] storage insts = ctrs[rcids[i]].instances;
            insts[insts.length - 1].startedAt = block.timestamp;
            activeCtrs.push(rcids[i]);

            // Update the sender reputation
            updateReputation("activateCtr");
        }
    }

    function updateContainerInfo(uint64 rcid, string memory _info) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("updateCtrInfo") // Required reputation to update containers info
            ),
            "The node has not enough reputation"
        );
        require(existContainer(rcid), "The container does not exist");
        require(
            isApplicationOwner(ctrs[rcid].appId, msg.sender),
            "The node is not the container owner"
        );
        require(
            !isContainerUnregistered(rcid),
            "The container was unregistered"
        );

        ctrs[rcid].info = _info;

        emit ContainerUpdated(rcid);

        // Update the sender reputation
        updateReputation("updateCtrInfo");
    }

    function unregisterApplication(uint64 appId) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("unregisterApp") // Required reputation to unregister applications
            ),
            "The node has not enough reputation"
        );
        require(existApplication(appId), "The application does not exist");
        require(
            isApplicationOwner(appId, msg.sender),
            "The node is not the application owner"
        );
        require(
            !isApplicationUnregistered(appId),
            "The application has already been unregistered"
        );

        // Set application removal time and deactivate it
        apps[appId].unregisteredAt = block.timestamp;
        deactivateApplicationById(appId);

        // Update the sender reputation
        updateReputation("unregisterApp");

        // Unregister all application containers
        uint64 i = 0;
        while (i < activeCtrs.length) {
            if (ctrs[activeCtrs[i]].appId == appId) {
                require(
                    hasNodeReputation(
                        msg.sender,
                        faucet.getActionLimit("unregisterCtr") // Required reputation to unregister containers
                    ),
                    "The node has not enough reputation"
                );

                // Set container removal time and deactivate it
                ctrs[activeCtrs[i]].unregisteredAt = block.timestamp;
                activeCtrs[i] = activeCtrs[activeCtrs.length - 1];
                activeCtrs.pop();

                emit ContainerUnregistered(activeCtrs[i]);

                // Update the sender reputation
                updateReputation("unregisterCtr");
            } else i++;
        }
    }

    function unregisterContainer(uint64 rcid) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("unregisterCtr") // Required reputation to unregister containers
            ),
            "The node has not enough reputation"
        );
        require(existContainer(rcid), "The container does not exist");
        require(
            isApplicationOwner(ctrs[rcid].appId, msg.sender),
            "The node is not the container owner"
        );
        require(
            !isContainerUnregistered(rcid),
            "The container has already been unregistered"
        );

        // Set container removal time and deactivate it
        ctrs[rcid].unregisteredAt = block.timestamp;
        deactivateContainerById(rcid);

        emit ContainerUnregistered(rcid);

        // Update the sender reputation
        updateReputation("unregisterCtr");
    }

    /////////////
    // Getters //
    /////////////
    function getEventReplies(uint64 eid)
        public
        view
        returns (Event.Reply[] memory)
    {
        return events[eid].replies;
    }

    function getContainerInstances(uint64 rcid)
        public
        view
        returns (Container.Instance[] memory)
    {
        return ctrs[rcid].instances;
    }

    function getActiveApplications() public view returns (uint64[] memory) {
        return activeApps;
    }

    function getActiveContainers() public view returns (uint64[] memory) {
        return activeCtrs;
    }

    ///////////////////////
    // Private functions //
    ///////////////////////
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

    function recordContainer(
        uint64 _appId,
        string memory _info,
        bool autodeploy
    ) private {
        ctrs[state.nextRegistryCtrId].appId = _appId;
        ctrs[state.nextRegistryCtrId].info = _info;
        ctrs[state.nextRegistryCtrId].registeredAt = block.timestamp;

        // Owner autodeploy mode
        if (autodeploy) {
            Container.Instance memory inst;
            inst.host = msg.sender;
            ctrs[state.nextRegistryCtrId].instances.push(inst);
        }

        emit ContainerRegistered(state.nextRegistryCtrId);

        state.nextRegistryCtrId++;

        // Update the sender reputation
        updateReputation("registerCtr");
    }

    function deactivateApplicationById(uint64 appId) private {
        for (uint64 i = 0; i < activeApps.length; i++) {
            if (activeApps[i] == appId) {
                activeApps[i] = activeApps[activeApps.length - 1];
                activeApps.pop();
                //break;
            }
        }
    }

    function deactivateContainerById(uint64 rcid) private {
        for (uint64 i = 0; i < activeCtrs.length; i++) {
            if (activeCtrs[i] == rcid) {
                activeCtrs[i] = activeCtrs[activeCtrs.length - 1];
                activeCtrs.pop();
                //break;
            }
        }
    }
}
