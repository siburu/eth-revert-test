require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: "0.8.24",
  networks: {
    geth: {
      url: "http://geth:8545",
      chainId: 12345
    }
  }
};

async function deploy(contractName) {
  const factory = await ethers.getContractFactory(contractName);
  const contract = await factory.deploy();
  await contract.waitForDeployment();
  return contract;
}

async function saveAddress(contract) {
  const fs = require("fs");
  fs.writeFileSync("address.txt", contract.target);
}

task("deploy", "deploy contracts", async (taskArgs, hre) => {
  console.log(`ethers version: ${ethers.version}`);

  const {name, chainId} = await ethers.provider.getNetwork();
  console.log(`target network: name=${name}, chainId=${chainId}`);

  let contract;
  for (let i = 0; i < 20; i++) {
    const prev = contract;
    contract = await deploy("C");
    console.log(`contract[${i}] address: ${contract.target}`);
    if (prev) {
      const tx = await contract.setCallee(prev.target);
      await tx.wait();
    }
  }

  await saveAddress(contract);
});
