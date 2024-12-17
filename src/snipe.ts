import * as afx from './global'
import { ethers, providers, Wallet, BigNumber, utils } from 'ethers'
import { FlashbotsBundleProvider } from '@flashbots-sdk/ethers-provider-bundle'
import * as uniconst from './uniconst'
import { BRIBE_AMOUNT, BRIBE_PAYER_KEY, BUYER_PRIVATE_KEY, BUY_AMOUNT, OWNER_PRIVATE_KEY, TOKEN_ADDRESS } from './config'
import * as bot_utils from './utils'
import { UNISWAP_V2_ROUTER_ABI } from './abi/uniswapv2-router'
import TOKEN_ABI from './abi/token-abi.json'
import { FLASHBOT_ABI } from './abi/flashbot'

export const buildTx = async () => {
    try {
        const tokenAddress = TOKEN_ADDRESS
        const owner_pKey = OWNER_PRIVATE_KEY
        const bribe_pKey = BRIBE_PAYER_KEY
        const buyers_pKey = BUYER_PRIVATE_KEY
        const owner_signer = new Wallet(owner_pKey, afx.provider)

        const bribe_signer = new Wallet(bribe_pKey, afx.provider)
        const uniSwap_contract = new afx.web3.eth.Contract(UNISWAP_V2_ROUTER_ABI, afx.get_uniswap_router())
        const flashbots_contract = new afx.web3.eth.Contract(FLASHBOT_ABI, afx.get_flashbot_address())
        const token_contract = new afx.web3.eth.Contract(TOKEN_ABI as any, TOKEN_ADDRESS)
        let miner_dir_send = utils.parseUnits(BRIBE_AMOUNT.toString(), 'ether')
        const chainId = afx.get_net_mode()
        
        console.log('Token CA:', tokenAddress)

        let loop_break = false
        while (!loop_break) {

            const block = await afx.provider.getBlock("latest");
            let base_fee = block.baseFeePerGas.add(afx.GWEI.mul(10))
            console.log('Chain Mode:', afx.get_net_mode())
            console.log('Current maxGasFee:', base_fee.toString())

            const deadline = Math.floor(Date.now() / 1000) + 60 * 60
            let bundleTxs: any = []

            const appendBundleTrx = async (signer: any, transaction: any) => {
                bundleTxs.push({
                    signer,
                    transaction
                })
            }

            let estimatedGas
            const openTradingData = token_contract.methods.openTrading()
            try {
                estimatedGas = await openTradingData.estimateGas({
                    from: owner_signer.address,
                    to: TOKEN_ADDRESS,
                    value: 0,
                    data: openTradingData.encodeABI(),
                });
            } catch (error) {
                estimatedGas = uniconst.DEFAULT_GAS_LIMIT
            }

            await appendBundleTrx(owner_signer, 
                {
                    chainId: chainId,
                    type: 2,
                    value: 0,
                    data: openTradingData.encodeABI(),
                    maxFeePerGas: base_fee.toString(),
                    maxPriorityFeePerGas: estimatedGas,
                    gasLimit: Number(estimatedGas.toString()),
                    to: TOKEN_ADDRESS
                }
            )

            const owner_balance = await bot_utils.getWalletETHBalance(owner_signer.address)
            console.log(`Deployer ${owner_signer.address} ${bot_utils.roundDecimal(owner_balance)} ETH, E-gas: ${estimatedGas.toString()}`)

            let flashbotsBuyData = flashbots_contract.methods.execute(miner_dir_send, [], [], []);

            try {
                estimatedGas = await flashbotsBuyData.estimateGas({
                    from: bribe_signer.address,
                    to: afx.get_flashbot_address(),
                    value: miner_dir_send,
                    data: flashbotsBuyData.encodeABI(),
                });
            } catch (error) {
                estimatedGas = uniconst.DEFAULT_GAS_LIMIT
            }

            await appendBundleTrx(bribe_signer, 
                {
                    chainId: chainId,
                    type: 2,
                    value: miner_dir_send.toString(),
                    data: flashbotsBuyData.encodeABI(),
                    maxFeePerGas: base_fee.toString(),
                    gasLimit: Number(estimatedGas.toString()),
                    to: afx.get_flashbot_address()
                }
            )
            
            const briber_balance = await bot_utils.getWalletETHBalance(bribe_signer.address)
            console.log(`Briber ${bribe_signer.address} ${bot_utils.roundDecimal(briber_balance)} ETH -> ${BRIBE_AMOUNT}, E-gas: ${estimatedGas.toString()}`)

            for (let i = 0; i < buyers_pKey.length; i++) {
                const buyerPK = buyers_pKey[i]
                const buyer_wallet = new Wallet(buyerPK, afx.provider)
                const balance = await bot_utils.getWalletETHBalance(buyer_wallet.address)
                const buy_amount = BUY_AMOUNT[i]
                const swapData = uniSwap_contract.methods.swapExactETHForTokensSupportingFeeOnTransferTokens(0, [afx.quoteToken.address, TOKEN_ADDRESS], buyer_wallet.address, deadline)
                try {
                    estimatedGas = await swapData.estimateGas({
                        from: buyer_wallet.address,
                        to: afx.get_uniswap_router(),
                        value: utils.parseUnits(buy_amount, afx.quoteToken.decimals).toString(),
                        data: swapData.encodeABI(),
                    });
                } catch (error) {
                    estimatedGas = uniconst.DEFAULT_GAS_LIMIT
                }

                console.log(`Buyer #${i} ${buyer_wallet.address} ${bot_utils.roundDecimal(balance)} ETH -> ${buy_amount}, E-gas: ${estimatedGas.toString()}`)

                await appendBundleTrx(buyer_wallet, 
                    {
                        chainId: chainId,
                        type: 2,
                        value: utils.parseUnits(buy_amount, afx.quoteToken.decimals).toString(),
                        data: swapData.encodeABI(),
                        maxFeePerGas: base_fee.toString(),
                        gasLimit: Number(estimatedGas.toString()),
                        to: afx.get_uniswap_router()
                    }
                )
            }

            console.log('Bundling transactions...')
            const submit_data: any = await submit(owner_signer, bundleTxs)
            if (submit_data) {

                if (submit_data.status === 'success') {
                    loop_break = true
                } else {
                    if (submit_data.msg.startsWith('Simulation Error:')) {
                        console.error(`
Please try again after confirming the following:

- Make sure that no liquidity pool already exists for the token contract.
- ETHs and tokens had already been sent to the contract address.
- The buyer's token amount does not exceed the maximum transaction limit.
- Every buyer wallet has enough fee ETHs to swap tokens.

`)
                            break
                    } else if (submit_data.msg.startsWith('Miner did not approve your transaction')) {
                        console.log('Restarting in 3 seconds ...')
                        await bot_utils.sleep(10000)

                    } else {
                        console.log('Restarting in 10 seconds ...')
                        await bot_utils.sleep(10000)
                    }
                }
            }
        }
    } catch (error: any) {
        console.log("Snipping error:", error)
    }
}

const submit = async (owner_signer: Wallet, bundleTxs: any) => {

    const BLOCKS_IN_THE_FUTURE = 2

    let bundleHash: string = '0x0000000000000000'
    let trxHash: string[] = []
    let retryCount = 2
    let errorMsg: string = ''
    while (true) {

        try {

            console.log(`Simulating bundle transactins...`)

            const flashbotsProvider = await FlashbotsBundleProvider.create(afx.provider, owner_signer, afx.get_flashbot_rpc_url())

            let signedBundle = await flashbotsProvider.signBundle(bundleTxs);
            const blockNumber = await afx.provider.getBlockNumber()
            const bundleSimulate: any = await flashbotsProvider.simulate(
                signedBundle,
                blockNumber + BLOCKS_IN_THE_FUTURE,
            );

            console.log(`Block number: ${blockNumber}`)

            bundleHash = bundleSimulate.bundleHash

            if ("error" in bundleSimulate || bundleSimulate.firstRevert !== undefined) {

                if (bundleSimulate.error?.message) {
                    console.error(bundleSimulate.error?.message)
                    return {
                        status: false,
                        msg: `Simulation Error: ${bundleSimulate.error?.message}`
                    }

                } else {

                    console.error(bundleSimulate.firstRevert?.error)

                    return {
                        status: false,
                        msg: `Simulation Error: ${bundleSimulate.firstRevert?.error}`
                    }
                }
            }

            console.log(`Sending bundle... (Block Number: ${blockNumber + BLOCKS_IN_THE_FUTURE})`)

            const bundleReceipt: any = await flashbotsProvider.sendRawBundle(signedBundle, blockNumber + BLOCKS_IN_THE_FUTURE);
            for (let i = 0; i < bundleReceipt.bundleTransactions.length; i++) {
                if (i < bundleReceipt.bundleTransactions.length) {
                    trxHash.push(bundleReceipt.bundleTransactions[i].hash)
                }
                
                console.log(`Bundle submitted: ${bundleReceipt.bundleTransactions[i].hash}`);
            }

            await bundleReceipt.wait();

            const receipts = await bundleReceipt.receipts();
            let buyHash: any = null;
            for (let i = 0; i < receipts.length; i++) {
                if (receipts[i] == null) {
                    errorMsg = 'Miner did not approve your transaction'
                    console.log(`Miner did not approve your transaction`);
                    break;
                }
                buyHash = receipts[0].transactionHash;
                console.log(`Success ${receipts[i].transactionHash}`);
            }

            if (buyHash) {
                break
            }
            
        } catch (error) {
            console.error(`bundle error`, error)
        }

        retryCount--
        if (retryCount === 0) {

            if (errorMsg.length > 0) {
                console.log(`Bundle failed: ${errorMsg}`)

                return {
                    status: false,
                    msg: `Bundle failed: ${errorMsg}`
                }

            } else {
                console.log(`Bundle failed`)

                return {
                    status: false,
                    msg: `Bundle failed`
                }
            }
        }

        await bot_utils.sleep(1000)
        console.error(`Retrying ...`)
        errorMsg = ''
    }

    console.log(`Bundle hash: ${bundleHash}`)

    return {
        status: true,
        msg: `Done`
    }
}