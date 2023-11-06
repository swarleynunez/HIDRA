const Controller = artifacts.require("Controller");

contract("Controller", accounts => {

    let controllerInstance;

    // Mocha hook
    beforeEach("Init", async () => {
        controllerInstance = await Controller.new();
    });

    it("Test", async () => {
        await controllerInstance.registerNode("SPECS", { from: accounts[0] });
        await controllerInstance.registerNode("SPECS", { from: accounts[1] });
        await controllerInstance.registerNode("SPECS", { from: accounts[2] });

        await controllerInstance.registerApplication("APP_INFO", ["CTR_INFO"], false);
        await controllerInstance.sendEvent("EVENT", 1);

        let repScores = [
            ["0x539022a255e26C16A3F9C252aa5e50503AF554cc", "0"],
            ["0xaCD0d6aBF6D2Bef65c55b8AdFD3f3e9942b7C257", "1"]
        ];
        await controllerInstance.sendReply(1, repScores, { from: accounts[0] });
        repScores = [
            ["0xDb4BfC458422121bf3195BDAAf562f74aD44fd5F", "0"],
            ["0xaCD0d6aBF6D2Bef65c55b8AdFD3f3e9942b7C257", "1"]
        ];
        await controllerInstance.sendReply(1, repScores, { from: accounts[1] });
        repScores = [
            ["0xDb4BfC458422121bf3195BDAAf562f74aD44fd5F", "0"],
            ["0x539022a255e26C16A3F9C252aa5e50503AF554cc", "0"]
        ];
        await controllerInstance.sendReply(1, repScores, { from: accounts[2] });

        await controllerInstance.voteSolver(1, "0xaCD0d6aBF6D2Bef65c55b8AdFD3f3e9942b7C257", { from: accounts[0] });
        await controllerInstance.voteSolver(1, "0xaCD0d6aBF6D2Bef65c55b8AdFD3f3e9942b7C257", { from: accounts[1] });
        await controllerInstance.voteSolver(1, "0xaCD0d6aBF6D2Bef65c55b8AdFD3f3e9942b7C257", { from: accounts[2] });

        await controllerInstance.solveEvent(1, { from: accounts[2] });
    });
});
