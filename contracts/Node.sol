pragma solidity ^0.6.6;
pragma experimental ABIEncoderV2;


contract Node {
    // Contract owner address (node)
    address private owner;

    // Controller smart contract address
    address private controller;

    // Node reputation
    uint64 private reputation;

    // Node specifications structure
    struct Specs {
        string arch;
        uint64 cores; // Logical cores number
        string mhz; // Physical cores frequency
        uint64 memTotal; // In bytes
        uint64 diskTotal; // In bytes
        string fileSystem;
        string OS;
        string hostname;
        uint64 bootTime; // Unix time
    }

    Specs private nodeSpecs;

    // Constructor
    constructor(address _owner) public {
        owner = _owner;
        controller = msg.sender;
        reputation = 100;
    }

    // Modifiers
    modifier onlyOwner() {
        require(msg.sender == owner, "Only the owner can call this function");
        _;
    }

    modifier onlyController() {
        require(
            msg.sender == controller,
            "Only the controller can call this function"
        );
        _;
    }

    // Functions
    function getReputation() public view returns (uint64) {
        return reputation;
    }

    function setSpecs(
        string memory _arch,
        uint64 _cores,
        string memory _mhz,
        uint64 _memTotal,
        uint64 _diskTotal,
        string memory _fileSystem,
        string memory _OS,
        string memory _hostname,
        uint64 _bootTime
    ) public onlyOwner {
        nodeSpecs.arch = _arch;
        nodeSpecs.cores = _cores;
        nodeSpecs.mhz = _mhz;
        nodeSpecs.memTotal = _memTotal;
        nodeSpecs.diskTotal = _diskTotal;
        nodeSpecs.fileSystem = _fileSystem;
        nodeSpecs.OS = _OS;
        nodeSpecs.hostname = _hostname;
        nodeSpecs.bootTime = _bootTime;
    }

    function getSpecs() public view returns (Specs memory) {
        return nodeSpecs;
    }
}
