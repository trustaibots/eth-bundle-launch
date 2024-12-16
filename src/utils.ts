import assert from 'assert';
import * as afx from './global'

import * as uniconst from './uniconst'

import dotenv from 'dotenv'
dotenv.config()

export const sleep = (ms: number) => {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export const getWalletETHBalance = async(wallet: string): Promise<number> => {

    assert(afx.web3)

    try {
        const ethValue = Number(await afx.web3.eth.getBalance(wallet)) / (10 ** uniconst.WETH_DECIMALS)
        return ethValue
    } catch (error) {
        console.log(error)
    }
     
    return 0
}

export const roundDecimal = (number: number, digits: number = 5) => {
    return number.toLocaleString('en-US', { maximumFractionDigits: digits });
}

export const getTokenInfo = async (tokenAddress: string) : Promise<any | null> => {

    assert(afx.web3)

    return new Promise(async (resolve, reject) => {

        let tokenContract: any
        try {
            tokenContract = new afx.web3.eth.Contract(afx.get_ERC20_abi(), tokenAddress);
            var tokenPromises: any[] = [];
            tokenPromises.push(tokenContract.methods.name().call());
            tokenPromises.push(tokenContract.methods.symbol().call());
            tokenPromises.push(tokenContract.methods.decimals().call());
            tokenPromises.push(tokenContract.methods.totalSupply().call());
            Promise.all(tokenPromises).then(tokenInfo => {
                console.log(tokenInfo)
                const decimals = parseInt(tokenInfo[2])
                const totalSupply = Number(tokenInfo[3]) / 10 ** decimals
                const result = { address: tokenAddress, name: tokenInfo[0], symbol: tokenInfo[1], decimals: decimals, totalSupply }

                resolve(result)

            }).catch(err => {

                resolve(null)
            })
        } catch (error) {
            console.log("getTokenInfo", error)
            resolve(null)
            return
        }
    })
}
