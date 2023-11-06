// SPDX-License-Identifier: MIT
pragma solidity ^0.8.21;

import "./DDR.sol";

contract Node {
    // Contract owner address
    address private addr;

    // Object containing node data
    DDR.NodeData private data;

    //constructor(address _addr, string memory _specs, int64 _reputation) {
    constructor(address _addr, string memory _specs) {
        addr = _addr;
        data.controller = msg.sender;
        data.specs = _specs; // Node specifications
        //data.reputation = _reputation; // Reputation by default
        data.registeredAt = block.timestamp;
    }

    modifier onlyOwner() {
        require(
            msg.sender == addr,
            "Only the contract owner can call this function"
        );
        _;
    }

    modifier onlyController() {
        require(
            msg.sender == data.controller,
            "Only the controller contract can call this function"
        );
        _;
    }

    /////////////
    // Setters //
    /////////////
    function updateSpecs(string memory _specs) public onlyOwner {
        data.specs = _specs;
    }

    /*function setVariation(int64 variation) public onlyController {
        data.reputation += variation;
    }*/

    /////////////
    // Getters //
    /////////////
    function getSpecs() public view returns (string memory) {
        return data.specs;
    }

    /*function getReputation() public view returns (int64) {
        return data.reputation;
    }*/
}
