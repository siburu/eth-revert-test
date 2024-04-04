require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.24",
};

async function deploy(contractName) {
  const factory = await hre.ethers.getContractFactory(contractName);
  const contract = await factory.deploy();
  await contract.waitForDeployment();
  return contract;
}

async function saveAddress(contract) {
  const fs = require("fs");
  fs.writeFileSync("address.txt", contract.target);
}

task("deploy", "deploy contracts", async (taskArgs, hre) => {
  let contract, prev;
  for (let i = 0; i < 3; i++) {
    contract = await deploy("C");
    if (prev) {
      const tx = await contract.setCallee(prev.target);
      await tx.wait();
    }
    prev = contract;
  }

  await saveAddress(contract);
});
