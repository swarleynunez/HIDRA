pragma solidity ^0.6.6;


contract Faucet {
    // Limit, reward and penalty reputations for system actions
    struct ActionReps {
        uint64 limit;
        uint64 reward;
        uint64 penalty;
    }

    // System actions list
    mapping(string => ActionReps) public actions;

    // Constructor
    constructor() public {
        // Initialize system actions
        actions["sendEvent"] = ActionReps(100, 0, 0);
    }

    // Functions
    function getActionLimit(string memory action) public view returns (uint64) {
        return actions[action].limit;
    }

    function getActionReward(string memory action)
        public
        view
        returns (uint64)
    {
        return actions[action].reward;
    }

    function getActionPenalty(string memory action)
        public
        view
        returns (uint64)
    {
        return actions[action].penalty;
    }
}
