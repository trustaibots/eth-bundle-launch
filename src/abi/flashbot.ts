export const FLASHBOT_ABI: any = [
	{
	  "inputs": [],
	  "stateMutability": "nonpayable",
	  "type": "constructor"
	},
	{
	  "inputs": [
		{
		  "internalType": "uint256",
		  "name": "_ethAmountToCoinbase",
		  "type": "uint256"
		},
		{
		  "internalType": "uint256[]",
		  "name": "_values",
		  "type": "uint256[]"
		},
		{
		  "internalType": "address[]",
		  "name": "_targets",
		  "type": "address[]"
		},
		{
		  "internalType": "bytes[]",
		  "name": "_payloads",
		  "type": "bytes[]"
		}
	  ],
	  "name": "execute",
	  "outputs": [],
	  "stateMutability": "payable",
	  "type": "function"
	},
	{
	  "stateMutability": "payable",
	  "type": "receive"
	}
  ]