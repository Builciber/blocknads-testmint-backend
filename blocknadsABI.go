package main

const abi string = `[
	{
		"inputs": [
		  {
			"internalType": "uint256",
			"name": "numTickets",
			"type": "uint256"
		  }
		],
		"name": "purchaseRaffleTickets",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	  },
	  {
		"inputs": [
		  {
			"internalType": "uint256",
			"name": "nonce",
			"type": "uint256"
		  },
		  {
			"internalType": "bytes",
			"name": "signature",
			"type": "bytes"
		  }
		],
		"name": "raffleMint",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	  },
	  {
		"inputs": [
		  {
			"internalType": "uint256",
			"name": "nonce",
			"type": "uint256"
		  },
		  {
			"internalType": "uint64",
			"name": "id",
			"type": "uint64"
		  },
		  {
			"internalType": "bytes",
			"name": "signature",
			"type": "bytes"
		  }
		],
		"name": "whitelistMint",
		"outputs": [],
		"stateMutability": "payable",
		"type": "function"
	  },
	  {
		"inputs": [
		  {
			"internalType": "uint256",
			"name": "",
			"type": "uint256"
		  }
		],
		"name": "slots",
		"outputs": [
		  {
			"internalType": "uint256",
			"name": "",
			"type": "uint256"
		  }
		],
		"stateMutability": "view",
		"type": "function"
	  },
	  {
		"inputs": [],
		"name": "MAX_NFTS",
		"outputs": [
		  {
			"internalType": "uint16",
			"name": "",
			"type": "uint16"
		  }
		],
		"stateMutability": "view",
		"type": "function"
	  }
]`
