// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

import "./DRS.sol";

contract Faucet {
    // Action list
    mapping(string => DRS.ReputableAction) private actions;

    constructor() {
        // Initialize system actions
        actions["sendEvent"] = DRS.ReputableAction(100, 1);
        actions["sendReply"] = DRS.ReputableAction(100, 1);
        actions["voteSolver"] = DRS.ReputableAction(100, 1);
        actions["solveEvent"] = DRS.ReputableAction(100, 1);
        actions["registerApp"] = DRS.ReputableAction(100, 1);
        actions["registerCtr"] = DRS.ReputableAction(100, 1);
        actions["activateCtr"] = DRS.ReputableAction(100, 1);
        actions["updateCtr"] = DRS.ReputableAction(100, 1);
        actions["unregisterApp"] = DRS.ReputableAction(100, 1);
        actions["unregisterCtr"] = DRS.ReputableAction(100, 1);
    }

    /////////////
    // Getters //
    /////////////
    function getActionLimit(string memory action) public view returns (int64) {
        return actions[action].limit;
    }

    function getActionVariation(
        string memory action
    ) public view returns (int64) {
        return actions[action].variation;
    }
}
