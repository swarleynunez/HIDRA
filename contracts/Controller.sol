// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

import "./Faucet.sol";
import "./Cluster.sol";
import "./Node.sol";
import "./DEL.sol";
import "./DCR.sol";

contract Controller {
    // Faucet smart contract instance
    //Faucet public faucet;

    // Cluster configuration
    Cluster.State public state;
    Cluster.Config public config;

    // Node list
    mapping(address => address) public nodes;
    address[] private registeredNodes;

    // Event list
    mapping(uint64 => DEL.Event) public events;
    uint64[] private currentEvents;

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

    constructor() {
        // Faucet instance
        //faucet = new Faucet();

        // Initialize cluster
        state = Cluster.State(0, 1, 1, 1, block.timestamp);
        config = Cluster.Config(66, 66, 4);
    }

    /////////////
    // Setters //
    /////////////
    function registerNode(string memory specs) public {
        require(
            !isNodeRegistered(msg.sender),
            "The node is already registered"
        );

        nodes[msg.sender] = address(new Node(msg.sender, specs));

        // Update node list
        registeredNodes.push(msg.sender);
        state.nodeCount++;
    }

    /////////////////////////
    // Reputable functions //
    /////////////////////////
    function sendEvent(string memory _eType, uint64 _rcid) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        /*require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("sendEvent") // Required reputation to send events
            ),
            "The node has not enough reputation"
        );*/

        // Update event list
        currentEvents.push(state.nextEventId);

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

        emit NewEvent(state.nextEventId);
        state.nextEventId++;

        // Update the sender reputation
        //updateReputation("sendEvent");
    }

    function sendReply(
        uint64 eid,
        DEL.ReputationScore[] memory _repScores
    ) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        /*require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("sendReply") // Required reputation to send replies
            ),
            "The node has not enough reputation"
        );*/
        require(existEvent(eid), "The event does not exist");
        require(!isEventSolved(eid), "The event is solved");
        require(
            !events[eid].hasRequiredReplies &&
                !hasRequiredCount(
                    config.nodesTh,
                    uint64(events[eid].replies.length)
                ),
            "The event already has the required replies"
        );
        require(
            !hasAlreadyReplied(eid, msg.sender),
            "The node has already replied the event"
        );
        events[eid].repliers[msg.sender] = true;

        // Create a new reply
        DEL.EventReply storage reply = events[eid].replies.push();
        reply.replier = msg.sender;
        reply.repliedAt = block.timestamp;
        for (uint64 i = 0; i < _repScores.length; i++) {
            require(
                _repScores[i].node != msg.sender,
                "Self-reputations are not allowed"
            );
            require(
                isNodeRegistered(_repScores[i].node),
                "Invalid reputation score (the node is not registered)"
            );
            require(
                !reply.reputedNodes[_repScores[i].node],
                "Invalid reputation score (repeated node)"
            );
            reply.reputedNodes[_repScores[i].node] = true;

            // Store reputation score
            reply.repScores.push(
                DEL.ReputationScore(_repScores[i].node, _repScores[i].score)
            );
        }

        // Required replies?
        if (
            hasRequiredCount(config.nodesTh, uint64(events[eid].replies.length))
        ) {
            events[eid].hasRequiredReplies = true;
            emit RequiredReplies(eid);
        }

        // Update the sender reputation
        //updateReputation("sendReply");
    }

    function voteSolver(uint64 eid, address candidateAddr) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        /*require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("voteSolver") // Required reputation to vote solvers
            ),
            "The node has not enough reputation"
        );*/
        require(existEvent(eid), "The event does not exist");
        require(!isEventSolved(eid), "The event is solved");
        require(
            events[eid].hasRequiredReplies,
            "The event does not have the required replies"
        );
        require(
            !events[eid].hasRequiredVotes,
            "The event already has the required votes"
        );
        require(
            isNodeRegistered(candidateAddr),
            "The candidate node is not registered"
        );
        require(
            !hasAlreadyVoted(eid, msg.sender),
            "The node has already voted a solver"
        );
        events[eid].voters[msg.sender] = true;

        // Vote candidate
        events[eid].votes[candidateAddr]++;

        // Required votes?
        if (
            hasRequiredCount(config.votesTh, events[eid].votes[candidateAddr])
        ) {
            events[eid].solver = candidateAddr; // Set event solver
            events[eid].hasRequiredVotes = true;
            emit RequiredVotes(eid);
        }

        // Update the sender reputation
        //updateReputation("voteSolver");
    }

    function solveEvent(uint64 eid) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        /*require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("solveEvent") // Required reputation to solve events
            ),
            "The node has not enough reputation"
        );*/
        require(existEvent(eid), "The event does not exist");
        require(!isEventSolved(eid), "The event is already solved");
        require(
            canSolveEvent(eid, msg.sender),
            "The node cannot solve the event"
        );

        // Solve event
        events[eid].solvedAt = block.timestamp;

        uint64 rcid = events[eid].rcid;
        if (rcid > 0) {
            // Add a new container instance (if applicable)
            if (!isContainerHost(rcid, events[eid].solver)) {
                DCR.ContainerInstance memory ci;
                ci.host = msg.sender;
                ci.startedAt = block.timestamp;
                ctrs[rcid].instances.push(ci);
            }

            // Activate application (if applicable)
            if (!isApplicationActive(ctrs[rcid].appid))
                activeApps.push(ctrs[rcid].appid);

            // Activate container (if applicable)
            if (!isContainerActive(rcid)) activeCtrs.push(rcid);
        }

        // Update event list
        popArrayItemById(currentEvents, eid);

        emit EventSolved(eid);

        // Update the sender reputation
        //updateReputation("solveEvent");
    }

    function registerApplication(
        string memory appInfo,
        string[] memory ctrInfos,
        bool autodeploy
    ) public {
        require(isNodeRegistered(msg.sender), "The node is not registered");
        /*require(
            hasNodeReputation(
                msg.sender,
                faucet.getActionLimit("registerApp") // Required reputation to register applications
            ),
            "The node has not enough reputation"
        );*/

        // Register a new application
        apps[state.nextAppId].owner = msg.sender;
        apps[state.nextAppId].info = appInfo;
        apps[state.nextAppId].registeredAt = block.timestamp;

        emit ApplicationRegistered(state.nextCtrId);

        // Register all application containers
        for (uint64 i = 0; i < ctrInfos.length; i++) {
            /*require(
                hasNodeReputation(
                    msg.sender,
                    faucet.getActionLimit("registerCtr") // Required reputation to register containers
                ),
                "The node has not enough reputation"
            );*/

            recordContainer(state.nextAppId, ctrInfos[i], autodeploy);
        }

        state.nextAppId++;

        // Update the sender reputation
        //updateReputation("registerApp");
    }

    /*function registerContainer(
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
        //updateReputation("activateCtr");
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
            !isContainerInCurrentEvent(rcid),
            "The container is being managed by the cluster"
        );
        require(
            !isContainerUnregistered(rcid),
            "The container was unregistered"
        );

        ctrs[rcid].info = _info;

        emit ContainerUpdated(rcid);

        // Update the sender reputation
        //updateReputation("updateCtr");
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

        // Unregister all application containers
        for (uint64 i = 0; i < apps[appid].rcids.length; i++) {
            require(
                hasNodeReputation(
                    msg.sender,
                    faucet.getActionLimit("unregisterCtr") // Required reputation to unregister containers
                ),
                "The node has not enough reputation"
            );
            require(
                !isContainerInCurrentEvent(apps[appid].rcids[i]),
                "An application's container is being managed by the cluster"
            );

            unrecordContainer(apps[appid].rcids[i]);
        }

        // Update the sender reputation
        //updateReputation("unregisterApp");
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
            !isContainerInCurrentEvent(rcid),
            "The container is being managed by the cluster"
        );
        require(
            !isContainerUnregistered(rcid),
            "The container has already been unregistered"
        );

        unrecordContainer(rcid);
    }*/

    ///////////////////////
    // Private functions //
    ///////////////////////
    /*function updateReputation(string memory action) private {
        Node node = Node(nodes[msg.sender]);
        // Node instance
        node.setVariation(faucet.getActionVariation(action));
    }*/

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
        //updateReputation("registerCtr");
    }

    /*function unrecordContainer(uint64 rcid) private {
        // Set container removal time
        ctrs[rcid].unregisteredAt = block.timestamp;

        // Unlink container from its application
        popArrayItemById(apps[ctrs[rcid].appid].rcids, rcid);

        // Deactivate container
        popArrayItemById(activeCtrs, rcid);

        emit ContainerUnregistered(rcid);

        // Update the sender reputation
        //updateReputation("unregisterCtr");
    }*/

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
    function getRegisteredNodes() public view returns (address[] memory) {
        return registeredNodes;
    }

    /*function getCurrentEvents() public view returns (uint64[] memory) {
        return currentEvents;
    }*/

    function getEventReplyCount(uint64 eid) public view returns (uint) {
        return events[eid].replies.length;
    }

    function getEventReply(
        uint64 eid,
        uint64 index
    ) public view returns (address, DEL.ReputationScore[] memory, uint) {
        return (
            events[eid].replies[index].replier,
            events[eid].replies[index].repScores,
            events[eid].replies[index].repliedAt
        );
    }

    function getApplicationContainers(
        uint64 appid
    ) public view returns (uint64[] memory) {
        return apps[appid].rcids;
    }

    function getContainerInstances(
        uint64 rcid
    ) public view returns (DCR.ContainerInstance[] memory) {
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

    /*function hasNodeReputation(
        address nodeAddr,
        int64 reputation
    ) public view returns (bool) {
        Node node = Node(nodes[nodeAddr]);
        // Node instance
        if (node.getReputation() >= reputation) return true;
        return false;
    }*/

    function hasRequiredCount(
        uint8 th,
        uint64 count
    ) public view returns (bool) {
        uint64 required = (state.nodeCount * th) / 100;
        if (required == 0 || required == count) return true;
        return false;
    }

    function isApplicationOwner(
        uint64 appid,
        address nodeAddr
    ) public view returns (bool) {
        if (apps[appid].owner == nodeAddr) return true;
        return false;
    }

    function isContainerHost(
        uint64 rcid,
        address nodeAddr
    ) public view returns (bool) {
        uint length = ctrs[rcid].instances.length;
        if (length > 0 && ctrs[rcid].instances[length - 1].host == nodeAddr)
            return true;
        return false;
    }

    function hasAlreadyReplied(
        uint64 eid,
        address nodeAddr
    ) public view returns (bool) {
        return events[eid].repliers[nodeAddr];
    }

    function hasAlreadyVoted(
        uint64 eid,
        address nodeAddr
    ) public view returns (bool) {
        return events[eid].voters[nodeAddr];
    }

    function existEvent(uint64 eid) public view returns (bool) {
        if (events[eid].sender != address(0)) return true;
        return false;
    }

    function isEventSolved(uint64 eid) public view returns (bool) {
        if (events[eid].solvedAt != 0) return true;
        return false;
    }

    function canSolveEvent(
        uint64 eid,
        address nodeAddr
    ) public view returns (bool) {
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

    function isApplicationUnregistered(
        uint64 appid
    ) public view returns (bool) {
        if (apps[appid].unregisteredAt != 0) return true;
        return false;
    }

    function existContainer(uint64 rcid) public view returns (bool) {
        if (ctrs[rcid].appid != 0) return true;
        return false;
    }

    function isContainerAutodeployed(uint64 rcid) public view returns (bool) {
        if (
            ctrs[rcid].autodeployed &&
            ctrs[rcid].instances.length == 1 &&
            ctrs[rcid].instances[0].startedAt == 0
        ) return true;
        return false;
    }

    function isContainerActive(uint64 rcid) public view returns (bool) {
        for (uint64 i = 0; i < activeCtrs.length; i++) {
            if (activeCtrs[i] == rcid) return true;
        }

        return false;
    }

    /*function isContainerInCurrentEvent(uint64 rcid) public view returns (bool) {
        for (uint64 i = 0; i < currentEvents.length; i++) {
            if (events[currentEvents[i]].rcid == rcid) return true;
        }

        return false;
    }*/

    function isContainerUnregistered(uint64 rcid) public view returns (bool) {
        if (ctrs[rcid].unregisteredAt != 0) return true;
        return false;
    }
}
