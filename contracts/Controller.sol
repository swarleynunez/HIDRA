// SPDX-License-Identifier: MIT
pragma solidity ^0.6.6;
pragma experimental ABIEncoderV2;

import "./Faucet.sol";
import "./Cluster.sol";
import "./Node.sol";
import "./DEL.sol";
import "./DCR.sol";

contract Controller {
    // Faucet smart contract instance
    Faucet public faucet;

    // Cluster configuration
    Cluster.State public state;
    Cluster.Config public config;

    // Node list
    mapping(address => address) public nodes;

    // Event list
    mapping(uint64 => DEL.Event) public events;

    // Application and container registries
    mapping(uint64 => DCR.Application) public apps;
    mapping(uint64 => DCR.Container) public ctrs;
    uint64[] private activeApps;
    uint64[] private activeCtrs;

    // Main flow events
    event NewEvent(uint64 eid);
    event RequiredReplies(uint64 eid);
    event RequiredVotes(uint64 eid);
    event EventSolved(uint64 eid);

    // DCR events
    event ApplicationRegistered(uint64 appid);
    event ContainerRegistered(uint64 rcid);
    event ContainerUpdated(uint64 rcid);
    event ContainerUnregistered(uint64 rcid);

    constructor() public {
        // Faucet instance
        faucet = new Faucet();

        // Initialize cluster
        state = Cluster.State(0, 1, 1, 1, block.timestamp);
        config = Cluster.Config(100, 100, 100);
    }

    /////////////
    // Setters //
    /////////////
    function registerNode(string memory specs) public {
        require(
            !isNodeRegistered(msg.sender),
            "The node is already registered"
        );

        nodes[msg.sender] = address(
            new Node(msg.sender, specs, config.initNodeRep)
        );
        // The "new" keyword creates a smart contract

        state.nodeCount++;
    }

    /////////////////////////
    // Reputable functions //
    /////////////////////////
    function sendEvent(
        string memory _eType,
        uint64 _rcid,
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
        events[state.nextEventId].eType = _eType;
        events[state.nextEventId].sender = msg.sender;
        events[state.nextEventId].sentAt = block.timestamp;

        if (_rcid > 0) {
            require(
                existContainer(_rcid),
                "The event's container does not exist"
            );
            require(
                isApplicationOwner(ctrs[_rcid].appid, msg.sender) ||
                isContainerHost(_rcid, msg.sender),
                "The node is neither the event's container owner nor the host"
            );
            require(
                !isContainerAutodeployed(_rcid),
                "The event's container is in autodeploy mode"
            );
            require(
                !isContainerUnregistered(_rcid),
                "The event's container was unregistered"
            );

            // Link an existing container to the new event
            events[state.nextEventId].rcid = _rcid;
        }

        // Link the first reply (event sender)
        DEL.EventReply memory reply;
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
        DEL.EventReply memory reply;
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

        DEL.EventReply[] memory replies = events[eid].replies;

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
                    uint64 rcid = events[eid].rcid;
                    if (rcid > 0) {
                        if (!isContainerHost(rcid, candidateAddr)) {
                            DCR.ContainerInstance memory inst;
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
        uint64 rcid = events[eid].rcid;
        if (rcid > 0) {
            // Activate application
            if (!isApplicationActive(ctrs[rcid].appid))
                activeApps.push(ctrs[rcid].appid);

            // Set container start time and activate it
            DCR.ContainerInstance[] storage insts = ctrs[rcid].instances;
            if (insts[insts.length - 1].startedAt == 0)
                insts[insts.length - 1].startedAt = block.timestamp;
            if (!isContainerActive(rcid)) activeCtrs.push(rcid);
        }

        emit EventSolved(eid);

        // Update the sender reputation
        updateReputation("solveEvent");
    }

    function registerApplication(
        string memory appInfo,
        string[] memory ctrInfos,
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
        apps[state.nextAppId].info = appInfo;
        apps[state.nextAppId].registeredAt = block.timestamp;

        emit ApplicationRegistered(state.nextCtrId);

        // Register all application containers
        for (uint64 i = 0; i < ctrInfos.length; i++) {
            require(
                hasNodeReputation(
                    msg.sender,
                    faucet.getActionLimit("registerCtr") // Required reputation to register containers
                ),
                "The node has not enough reputation"
            );

            recordContainer(state.nextAppId, ctrInfos[i], autodeploy);
        }

        state.nextAppId++;

        // Update the sender reputation
        updateReputation("registerApp");
    }

    function registerContainer(
        uint64 appid,
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
        require(existApplication(appid), "The application does not exist");
        require(
            isApplicationOwner(appid, msg.sender),
            "The node is not the application owner"
        );
        require(
            !isApplicationUnregistered(appid),
            "The application was unregistered"
        );

        recordContainer(appid, info, autodeploy);
    }

    function activateContainer(uint64 rcid) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("activateCtr") // Required reputation to activate containers
            ),
            "The node has not enough reputation"
        );
        require(existContainer(rcid), "The container does not exist");
        require(
            isApplicationOwner(ctrs[rcid].appid, msg.sender),
            "The node is not the container owner"
        );
        require(
            isContainerHost(rcid, msg.sender),
            "The node is not the container host"
        );
        require(
            !isContainerUnregistered(rcid),
            "The container was unregistered"
        );
        require(!isContainerActive(rcid), "The container is already activated");

        // Activate application
        if (!isApplicationActive(ctrs[rcid].appid))
            activeApps.push(ctrs[rcid].appid);

        // Set container start time and activate it
        DCR.ContainerInstance[] storage insts = ctrs[rcid].instances;
        insts[insts.length - 1].startedAt = block.timestamp;
        activeCtrs.push(rcid);

        // Update the sender reputation
        updateReputation("activateCtr");
    }

    function updateContainerInfo(uint64 rcid, string memory _info) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("updateCtr") // Required reputation to update containers
            ),
            "The node has not enough reputation"
        );
        require(existContainer(rcid), "The container does not exist");
        require(
            isApplicationOwner(ctrs[rcid].appid, msg.sender),
            "The node is not the container owner"
        );
        require(
            !isContainerUnregistered(rcid),
            "The container was unregistered"
        );

        ctrs[rcid].info = _info;

        emit ContainerUpdated(rcid);

        // Update the sender reputation
        updateReputation("updateCtr");
    }

    function unregisterApplication(uint64 appid) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("unregisterApp") // Required reputation to unregister applications
            ),
            "The node has not enough reputation"
        );
        require(existApplication(appid), "The application does not exist");
        require(
            isApplicationOwner(appid, msg.sender),
            "The node is not the application owner"
        );
        require(
            !isApplicationUnregistered(appid),
            "The application has already been unregistered"
        );

        // Set application removal time and deactivate it
        apps[appid].unregisteredAt = block.timestamp;
        popArrayItemById(activeApps, appid);

        uint64[] memory rcids = apps[appid].rcids;

        // Unregister all application containers
        for (uint64 i = 0; i < rcids.length; i++) {
            require(
                hasNodeReputation(
                    msg.sender,
                    faucet.getActionLimit("unregisterCtr") // Required reputation to unregister containers
                ),
                "The node has not enough reputation"
            );

            unrecordContainer(rcids[i]);
        }

        // Update the sender reputation
        updateReputation("unregisterApp");
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
            isApplicationOwner(ctrs[rcid].appid, msg.sender),
            "The node is not the container owner"
        );
        require(
            !isContainerUnregistered(rcid),
            "The container has already been unregistered"
        );

        unrecordContainer(rcid);
    }

    ///////////////////////
    // Private functions //
    ///////////////////////
    function updateReputation(string memory action) private {
        Node node = Node(nodes[msg.sender]);
        // Node instance
        node.setVariation(faucet.getActionVariation(action));
    }

    function recordContainer(
        uint64 _appid,
        string memory _info,
        bool autodeploy
    ) private {
        apps[_appid].rcids.push(state.nextCtrId);
        ctrs[state.nextCtrId].appid = _appid;
        ctrs[state.nextCtrId].info = _info;
        ctrs[state.nextCtrId].registeredAt = block.timestamp;

        // Owner autodeploy mode
        if (autodeploy) {
            ctrs[state.nextCtrId].autodeployed = true;
            DCR.ContainerInstance memory inst;
            inst.host = msg.sender;
            ctrs[state.nextCtrId].instances.push(inst);
        }

        emit ContainerRegistered(state.nextCtrId);

        state.nextCtrId++;

        // Update the sender reputation
        updateReputation("registerCtr");
    }

    function unrecordContainer(uint64 rcid) private {
        ctrs[rcid].unregisteredAt = block.timestamp;
        // Set container removal time
        popArrayItemById(apps[ctrs[rcid].appid].rcids, rcid);
        // Unlink container from its application
        popArrayItemById(activeCtrs, rcid);
        // Deactivate container

        emit ContainerUnregistered(rcid);

        // Update the sender reputation
        updateReputation("unregisterCtr");
    }

    function popArrayItemById(uint64[] storage array, uint64 id) private {
        for (uint64 i = 0; i < array.length; i++) {
            if (array[i] == id) {
                array[i] = array[array.length - 1];
                array.pop();
                break;
            }
        }
    }

    /////////////
    // Getters //
    /////////////
    function getEventReplies(uint64 eid)
    public
    view
    returns (DEL.EventReply[] memory)
    {
        return events[eid].replies;
    }

    function getApplicationContainers(uint64 appid)
    public
    view
    returns (uint64[] memory)
    {
        return apps[appid].rcids;
    }

    function getContainerInstances(uint64 rcid)
    public
    view
    returns (DCR.ContainerInstance[] memory)
    {
        return ctrs[rcid].instances;
    }

    function getActiveApplications() public view returns (uint64[] memory) {
        return activeApps;
    }

    function getActiveContainers() public view returns (uint64[] memory) {
        return activeCtrs;
    }

    /////////////
    // Helpers //
    /////////////
    function isNodeRegistered(address nodeAddr) public view returns (bool) {
        if (nodes[nodeAddr] != address(0)) return true;
        return false;
    }

    function hasNodeReputation(address nodeAddr, int64 reputation)
    public
    view
    returns (bool)
    {
        Node node = Node(nodes[nodeAddr]);
        // Node instance
        if (node.getReputation() >= reputation) return true;
        return false;
    }

    function hasRequiredCount(uint8 thld, uint64 count)
    public
    view
    returns (bool)
    {
        uint64 required = (state.nodeCount * thld) / 100;
        if (required > 0 && count >= required) return true;
        return false;
    }

    function isApplicationOwner(uint64 appid, address nodeAddr)
    public
    view
    returns (bool)
    {
        if (apps[appid].owner == nodeAddr) return true;
        return false;
    }

    function isContainerHost(uint64 rcid, address nodeAddr)
    public
    view
    returns (bool)
    {
        DCR.ContainerInstance[] memory insts = ctrs[rcid].instances;
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
        DEL.EventReply[] memory replies = events[eid].replies;
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
        DEL.EventReply[] memory replies = events[eid].replies;
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

    function existApplication(uint64 appid) public view returns (bool) {
        if (apps[appid].owner != address(0)) return true;
        return false;
    }

    function isApplicationActive(uint64 appid) public view returns (bool) {
        for (uint64 i = 0; i < activeApps.length; i++) {
            if (activeApps[i] == appid) return true;
        }

        return false;
    }

    function isApplicationUnregistered(uint64 appid)
    public
    view
    returns (bool)
    {
        if (apps[appid].unregisteredAt != 0) return true;
        return false;
    }

    function existContainer(uint64 rcid) public view returns (bool) {
        if (ctrs[rcid].appid != 0) return true;
        return false;
    }

    function isContainerAutodeployed(uint64 rcid) public view returns (bool) {
        DCR.ContainerInstance[] memory insts = ctrs[rcid].instances;
        if (
            ctrs[rcid].autodeployed &&
            insts.length == 1 &&
            insts[0].startedAt == 0
        ) return true;
        return false;
    }

    function isContainerActive(uint64 rcid) public view returns (bool) {
        for (uint64 i = 0; i < activeCtrs.length; i++) {
            if (activeCtrs[i] == rcid) return true;
        }

        return false;
    }

    function isContainerUnregistered(uint64 rcid) public view returns (bool) {
        if (ctrs[rcid].unregisteredAt != 0) return true;
        return false;
    }
}
