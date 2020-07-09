pragma solidity ^0.6.6;

contract Faucet {
    // Limit, reward and penalty reputations for system actions
    struct ActionReps {
        int64 limit;
        int64 variation;
    }

    // System actions list
    mapping(string => ActionReps) private actions;

    // Constructor
    constructor() public {
        // Initialize system actions
        actions["sendEvent"] = ActionReps(100, 1);
        actions["sendReply"] = ActionReps(100, 1);
        actions["voteSolver"] = ActionReps(100, 1);
        actions["solveEvent"] = ActionReps(100, 1);
        actions["recordContainer"] = ActionReps(100, 1);
        actions["removeContainer"] = ActionReps(100, 1);
    }

    // Functions
    function getActionLimit(string memory action) public view returns (int64) {
        return actions[action].limit;
    }

    function getActionVariation(string memory action)
        public
        view
        returns (int64)
    {
        return actions[action].variation;
    }
}
