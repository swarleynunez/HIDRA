pragma solidity ^0.6.6;

contract Node {
    // Contract owner address (node)
    address private owner;

    // Controller smart contract address
    address private controller;

    // Node reputation
    int64 private reputation;

    // Node specifications
    string public nodeSpecs;

    // Constructor
    constructor(address nodeAddr, string memory specs) public {
        owner = nodeAddr;
        controller = msg.sender;

        // Reputation by default
        reputation = 100;

        // First specifications update
        nodeSpecs = specs;
    }

    // Modifiers
    modifier onlyOwner() {
        require(
            msg.sender == owner,
            "Only the contract owner can call this function"
        );
        _;
    }

    modifier onlyController() {
        require(
            msg.sender == controller,
            "Only the controller contract can call this function"
        );
        _;
    }

    // Functions
    function getReputation() public view returns (int64) {
        return reputation;
    }

    function updateSpecs(string memory specs) public onlyOwner {
        nodeSpecs = specs;
    }

    function setVariation(int64 variation) public onlyController {
        reputation += variation;
    }
}
