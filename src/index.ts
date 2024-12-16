
import * as afx from './global'

import dotenv from 'dotenv'
import Web3 from 'web3'
import { providers } from 'ethers'
import * as snipe from "./snipe"
dotenv.config()


export const web3 = new Web3(new Web3.providers.HttpProvider(afx.get_ethereum_rpc_http_url()))
export const provider = new providers.JsonRpcProvider(afx.get_ethereum_rpc_http_url())

afx.setWeb3(web3, provider)

const start = async () => {
    await afx.init()

    snipe.buildTx()
}

start()
