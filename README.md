# Eth-Bundle-Launch v1.1
**Ethereum Token Bundle Launch Project**

This project will perform an bundle token launch of ERC-20. This will be done by 3 steps
1. Token Contract deployment
2. Estimate the required ETHs for each buyer wallets
3. Configure and launch script

## 1. Token Contract deployment

The project has two main components, one is the script and another is the smart contract. 

### Steps to Customize the Smart Contract:

Navigate to either contracts/contract-mainnet.sol or contracts/contract-sepolia.sol, depending on your target network:

- contract-mainnet.sol: Contract for Ethereum Mainnet
- contract-sepolia.sol: Contract for Ethereum Sepolia Testnet

The only difference between these contracts is the Uniswap V2 router address specified in the code.
Customize the contract as needed to fit your requirements.

### Deploy the Smart Contract

- Use [Remix](https://remix.ethereum.org/) to deploy the contract

Once deployed, ensure that:
- You have transferred ETH to the contract. This ETH will serve as the amount for initial liquidity.
- You have transferred tokens to the contract. These tokens will also be used for initial liquidity.

### Updating the ABI

If any function or variable names in the contract are modified:

- Download the updated ABI JSON file using Etherscan or Remix.
- Replace the file located at src/abi/token-abi.json with the new ABI JSON file.

## 2. Estimate the required ETHs for each buyer wallets

Use python script ./calculate_eth_for_wallets.py on Ubuntu

- Install python3

apt install python3

- Modify parameters

// Parameters in calculate_eth_for_wallets.py
...
initial_eth = 2.0	// ETHs in initial liquidity 
initial_tokens = 100    // Tokens in initial liquidity (100 is fine, no need to change this parameter)
num_wallets = 50 	// Wallet count to be used as buyers
target_percentage = 80  // Token hold percentage by sniping
...

- Execute the python script

python3 calculate_eth_for_wallets.py

Execution results are
```bash
Wallet 1: Needs 0.0325 ETH
Wallet 2: Needs 0.0336 ETH
Wallet 3: Needs 0.0347 ETH

...

Wallet 49: Needs 0.6386 ETH
Wallet 50: Needs 0.7407 ETH
Total ETH required by all wallets: 8.0000 ETH
```

## 3. Configure and launch script

### Configure 

Navigate src/config.ts and configure the script

```ts
export const MAINNET_RPC= "https://boldest-bold-uranium.quiknode.pro/a5e9ce66d6648e49889274a783acd07aebcc02bc/"
// Replace this with your mainnet rpc
export const SEPOLIA_RPC="https://old-green-glade.ethereum-sepolia.quiknode.pro/0523b575936957f0e7eae638096d19465aae8f8c/"
// Replace this with your sepolia testnet rpc

export const FLASHBOTS_ENDPOINT="https://relay.flashbots.net"
export const SEPOLIA_FLASHBOTS_ENDPOINT="https://relay-sepolia.flashbots.net"
export const VERSION = 1.0

// Ethereum = 1,
// Sepolia = 11155111
export const NET_MODE = 11155111             // Set the chain ID to specify whether to launch on Mainnet or Testnet
export const OWNER_PRIVATE_KEY = "af264be3f6a97b5ef1f19b675e2fe84ed15fd726ba38e59c5468d95f53f6de71"; 
// Replace this with your own private key.

export const BRIBE_PAYER_KEY = "a36d14d380505993394deef92e13d079e1e0053b0f294939d4679c5d3d80671c"; 
// Replace this with the private key of the bribe payer.

export const BUYER_PRIVATE_KEY = [
    "2cc095269dc37126b5df0307534ae78a5c4287459041d3fc3d83225def084b28", // Replace with buyer wallet private keys.
    "c50f78fcc2f162e2c02ae24372d52e11c79b0fe4f7ed00ffd7d6ed623b14f641", // ...
    "c50f78fcc2f162e2c02ae24372d52e11c79b0fe4f7ed00ffd7d6ed623b14f641", // ...
    // Add more keys as needed...
    "c50f78fcc2f162e2c02ae24372d52e11c79b0fe4f7ed00ffd7d6ed623b14f641", // ...
    "6acd8aa799d644271af6ce326648ba4c4d6da1dc1ba905cd70d3525a4d2a9537", // ...
];

export const BUY_AMOUNT = [
    "0.0325", // Replace with the calculated ETH amount per wallet from result of calculate_eth_for_wallets.py.
    "0.0336", // ...
    "0.0347", // ...
    // Add more amounts as needed...
    "0.6386", // ...
    "0.7407", // ...
];

export const TOKEN_ADDRESS = "0xcf1e7Df33a0Cb3046D56B17a5B7b30EA25c2fd44"; 
// Replace with the address of your deployed token contract.

export const BRIBE_AMOUNT = 0.1; 
// Set the bribe amount to be used with the flashbots provider.

```

### Launch preparation

- Ensure that all the wallets (Deployer, Briber, and Buyer) have enough balance to cover the transaction fees for sniping the tokens.

Deployer Wallet: Must have enough balance to cover the transaction fee for executing the ```openTrading``` function in the token contract.
Briber Wallet: Must have enough ETH to transfer the bribe and cover the associated transaction fee.
Buyer Wallets: Each buyer wallet must have enough ETH to purchase tokens, along with the transaction fee required for the token swap.

- Confirm that no liquidity pool already exists for the token contract.
- Make sure that ETH and tokens have already been transferred to the token contract address.
- Verify that the amount of tokens each buyer intends to purchase does not exceed the max-transaction limit. If necessary, adjust the ETH balance in the buyer wallets to accommodate the swap.

### Launch the script 

#### Run the following commands
```
> npm install
```

#### To start the script:
```
npm start
```

### Note: If you have renamed the 'openTrading' function in the contract, make sure to update the name in the snipe.ts file as well.

```ts
let estimatedGas
const openTradingData = token_contract.methods.openTrading() // Update the function name here to match the changes in your contract. Ensure the ABI is also updated as mentioned earlier.

try {
    estimatedGas = await openTradingData.estimateGas({
```

Always double-check that the ABI (```abi/token-abi.json```) is updated to reflect any changes in the contract before proceeding.

# Donate for further updates

Your support helps us continue improving this project and developing innovative solutions for the community. If you’d like to contribute, donations are warmly welcomed at the following wallet addresses:

ERC-20 / BEP-20 (EVM Compatible Wallets):
Wallet Address: ```0x782819b6564B13e4d8b70165CA2f3E212a6a0C2b```

TRC-20 (Tron Wallet):
Wallet Address: ```TG1YC57m8G5VGHRgmMmfJ96N7tUp9DU8pS```

Solana Wallet:
Wallet Address: ```HM2Yw3Zb1diPkMVhEttCHhEn9W75gzbZUyMaV6t8aWVr```

Your generosity will play a crucial role in supporting ongoing development, introducing new features, and enhancing the documentation for this project. Thank you for considering a donation and contributing to our progress!

If you make a donation, please let us know by emailing us at [devteam@trustaibots.com](mailto:devteam@trustaibots.com)

# Support

Feel free to reach out to our team on [Telegram](https://t.me/trustaibotsdevteam) if you have any questions or encounter any issues during the bundle launch

# Author

[Dev Team](https://github.com/trustaibots)
Telegram: https://t.me/trustaibotsdevteam
Website: https://trustaibots.com

