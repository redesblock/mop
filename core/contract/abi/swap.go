package abi

const ERC20SimpleSwapABIv0_1_0 = `[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_issuer",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "_token",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "_defaultHardDepositTimeout",
				"type": "uint256"
			}
		],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"anonymous": false,
		"inputs": [],
		"name": "ChequeBounced",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "caller",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "totalPayout",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "cumulativePayout",
				"type": "uint256"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "callerPayout",
				"type": "uint256"
			}
		],
		"name": "ChequeCashed",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "HardDepositAmountChanged",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "decreaseAmount",
				"type": "uint256"
			}
		],
		"name": "HardDepositDecreasePrepared",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "timeout",
				"type": "uint256"
			}
		],
		"name": "HardDepositTimeoutChanged",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "Withdraw",
		"type": "event"
	},
	{
		"inputs": [],
		"name": "CASHOUT_TYPEHASH",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "CHEQUE_TYPEHASH",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "CUSTOMDECREASETIMEOUT_TYPEHASH",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "EIP712DOMAIN_TYPEHASH",
		"outputs": [
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "balance",
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
		"name": "bounced",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "cumulativePayout",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "beneficiarySig",
				"type": "bytes"
			},
			{
				"internalType": "uint256",
				"name": "callerPayout",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "issuerSig",
				"type": "bytes"
			}
		],
		"name": "cashCheque",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "recipient",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "cumulativePayout",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "issuerSig",
				"type": "bytes"
			}
		],
		"name": "cashChequeBeneficiary",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			}
		],
		"name": "decreaseHardDeposit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "defaultHardDepositTimeout",
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
		"inputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"name": "hardDeposits",
		"outputs": [
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "decreaseAmount",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "timeout",
				"type": "uint256"
			},
			{
				"internalType": "uint256",
				"name": "canBeDecreasedAt",
				"type": "uint256"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "increaseHardDeposit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "issuer",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "liquidBalance",
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
		"inputs": [
			{
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			}
		],
		"name": "liquidBalanceFor",
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
		"inputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"name": "paidOut",
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
		"inputs": [
			{
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "decreaseAmount",
				"type": "uint256"
			}
		],
		"name": "prepareDecreaseHardDeposit",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "beneficiary",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "hardDepositTimeout",
				"type": "uint256"
			},
			{
				"internalType": "bytes",
				"name": "beneficiarySig",
				"type": "bytes"
			}
		],
		"name": "setCustomHardDepositTimeout",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "token",
		"outputs": [
			{
				"internalType": "contract ERC20",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "totalHardDeposit",
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
		"name": "totalPaidOut",
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
		"inputs": [
			{
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "withdraw",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

const SimpleSwapFactoryABIv0_1_0 = `[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_ERC20Address",
				"type": "address"
			}
		],
		"stateMutability": "nonpayable",
		"type": "constructor"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": false,
				"internalType": "address",
				"name": "contractAddress",
				"type": "address"
			}
		],
		"name": "SimpleSwapDeployed",
		"type": "event"
	},
	{
		"inputs": [],
		"name": "ERC20Address",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "issuer",
				"type": "address"
			},
			{
				"internalType": "uint256",
				"name": "defaultHardDepositTimeoutDuration",
				"type": "uint256"
			},
			{
				"internalType": "bytes32",
				"name": "",
				"type": "bytes32"
			}
		],
		"name": "deploySimpleSwap",
		"outputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "",
				"type": "address"
			}
		],
		"name": "deployedContracts",
		"outputs": [
			{
				"internalType": "bool",
				"name": "",
				"type": "bool"
			}
		],
		"stateMutability": "view",
		"type": "function"
	}
]`

const SimpleSwapFactoryDeployedBinv0_1_0 = "0x608060405234801561001057600080fd5b50600436106100415760003560e01c806315efd8a714610046578063a6021ace14610076578063c70242ad14610094575b600080fd5b610060600480360381019061005b9190610277565b6100c4565b60405161006d91906102f3565b60405180910390f35b61007e6101bc565b60405161008b91906102f3565b60405180910390f35b6100ae60048036038101906100a9919061024e565b6101e2565b6040516100bb9190610345565b60405180910390f35b60008084600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16856040516100f890610202565b6101049392919061030e565b604051809103906000f080158015610120573d6000803e3d6000fd5b50905060016000808373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060006101000a81548160ff0219169083151502179055507fc0ffc525a1c7689549d7f79b49eca900e61ac49b43d977f680bcc3b36224c004816040516101a991906102f3565b60405180910390a1809150509392505050565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b60006020528060005260406000206000915054906101000a900460ff1681565b612f47806103f883390190565b60008135905061021e816103b2565b92915050565b600081359050610233816103c9565b92915050565b600081359050610248816103e0565b92915050565b60006020828403121561026057600080fd5b600061026e8482850161020f565b91505092915050565b60008060006060848603121561028c57600080fd5b600061029a8682870161020f565b93505060206102ab86828701610239565b92505060406102bc86828701610224565b9150509250925092565b6102cf81610360565b82525050565b6102de81610372565b82525050565b6102ed816103a8565b82525050565b600060208201905061030860008301846102c6565b92915050565b600060608201905061032360008301866102c6565b61033060208301856102c6565b61033d60408301846102e4565b949350505050565b600060208201905061035a60008301846102d5565b92915050565b600061036b82610388565b9050919050565b60008115159050919050565b6000819050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b6103bb81610360565b81146103c657600080fd5b50565b6103d28161037e565b81146103dd57600080fd5b50565b6103e9816103a8565b81146103f457600080fd5b5056fe60806040523480156200001157600080fd5b5060405162002f4738038062002f478339818101604052810190620000379190620000f7565b82600660006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555081600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555080600081905550505050620001bf565b600081519050620000da816200018b565b92915050565b600081519050620000f181620001a5565b92915050565b6000806000606084860312156200010d57600080fd5b60006200011d86828701620000c9565b93505060206200013086828701620000c9565b92505060406200014386828701620000e0565b9150509250925092565b60006200015a8262000161565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b62000196816200014d565b8114620001a257600080fd5b50565b620001b08162000181565b8114620001bc57600080fd5b50565b612d7880620001cf6000396000f3fe608060405234801561001057600080fd5b50600436106101425760003560e01c8063b6343b0d116100b8578063b7ec1a331161007c578063b7ec1a3314610342578063c49f91d314610360578063c76a4d311461037e578063d4c9a8e8146103ae578063e0bcf13a146103ca578063fc0c546a146103e857610142565b8063b6343b0d14610299578063b648b417146102cc578063b69ef8a8146102ea578063b777035014610308578063b79989071461032457610142565b80631d1438481161010a5780631d143848146101d95780632e1a7d4d146101f7578063338f3fed14610213578063488b017c1461022f57806381f03fcb1461024d578063946f46a21461027d57610142565b80630d5f26591461014757806312101021146101635780631357e1dc1461018157806315c3343f1461019f5780631633fb1d146101bd575b600080fd5b610161600480360381019061015c9190611fc1565b610406565b005b61016b610419565b604051610178919061269a565b60405180910390f35b61018961041f565b604051610196919061269a565b60405180910390f35b6101a7610425565b6040516101b49190612394565b60405180910390f35b6101d760048036038101906101d29190611ecc565b610449565b005b6101e16104e1565b6040516101ee9190612335565b60405180910390f35b610211600480360381019061020c9190612051565b610507565b005b61022d60048036038101906102289190611f85565b6106f4565b005b6102376108c0565b6040516102449190612394565b60405180910390f35b61026760048036038101906102629190611ea3565b6108e4565b604051610274919061269a565b60405180910390f35b61029760048036038101906102929190611ea3565b6108fc565b005b6102b360048036038101906102ae9190611ea3565b610a39565b6040516102c394939291906126ec565b60405180910390f35b6102d4610a69565b6040516102e19190612379565b60405180910390f35b6102f2610a7c565b6040516102ff919061269a565b60405180910390f35b610322600480360381019061031d9190611f85565b610b2e565b005b61032c610cd6565b6040516103399190612394565b60405180910390f35b61034a610cfa565b604051610357919061269a565b60405180910390f35b610368610d1d565b6040516103759190612394565b60405180910390f35b61039860048036038101906103939190611ea3565b610d41565b6040516103a5919061269a565b60405180910390f35b6103c860048036038101906103c39190611fc1565b610da6565b005b6103d2610f51565b6040516103df919061269a565b60405180910390f35b6103f0610f57565b6040516103fd91906124df565b60405180910390f35b610414338484600085610f7d565b505050565b60005481565b60035481565b7f48ebe6deff4a5ee645c01506a026031e2a945d6f41f1f4e5098ad65347492c1281565b61045f61045930338789876115b4565b84611612565b73ffffffffffffffffffffffffffffffffffffffff168673ffffffffffffffffffffffffffffffffffffffff16146104cc576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104c39061261a565b60405180910390fd5b6104d98686868585610f7d565b505050505050565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610597576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161058e9061257a565b60405180910390fd5b61059f610cfa565b8111156105e1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016105d89061267a565b60405180910390fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16836040518363ffffffff1660e01b8152600401610660929190612350565b602060405180830381600087803b15801561067a57600080fd5b505af115801561068e573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106b29190612028565b6106f1576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106e8906125fa565b60405180910390fd5b50565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610784576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161077b9061257a565b60405180910390fd5b61078c610a7c565b6107a18260055461166290919063ffffffff16565b11156107e2576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016107d99061253a565b60405180910390fd5b6000600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020905061083c82826000015461166290919063ffffffff16565b81600001819055506108598260055461166290919063ffffffff16565b600581905550600081600301819055508273ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad82600001546040516108b3919061269a565b60405180910390a2505050565b7f7d824962dd0f01520922ea1766c987b1db570cd5db90bdba5ccf5e320607950281565b60026020528060005260406000206000915090505481565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060030154421015801561095857506000816003015414155b610997576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161098e906125da565b60405180910390fd5b6109b28160010154826000015461167890919063ffffffff16565b8160000181905550600081600301819055506109dd816001015460055461167890919063ffffffff16565b6005819055508173ffffffffffffffffffffffffffffffffffffffff167f2506c43272ded05d095b91dbba876e66e46888157d3e078db5691496e96c5fad8260000154604051610a2d919061269a565b60405180910390a25050565b60046020528060005260406000206000915090508060000154908060010154908060020154908060030154905084565b600660149054906101000a900460ff1681565b6000600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b8152600401610ad99190612335565b60206040518083038186803b158015610af157600080fd5b505afa158015610b05573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610b29919061207a565b905090565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610bbe576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610bb59061257a565b60405180910390fd5b6000600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002090508060000154821115610c48576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610c3f9061263a565b60405180910390fd5b600080826002015414610c5f578160020154610c63565b6000545b90508042610c7191906127a3565b82600301819055508282600101819055508373ffffffffffffffffffffffffffffffffffffffff167fc8305077b495025ec4c1d977b176a762c350bb18cad4666ce1ee85c32b78698a84604051610cc8919061269a565b60405180910390a250505050565b7fe95f353750f192082df064ca5142d3a2d6f0bef0f3ffad66d80d8af86b7a749a81565b6000610d18600554610d0a610a7c565b61167890919063ffffffff16565b905090565b7fc2f8787176b8ac6bf7215b4adcc1e069bf4ab82d9ab1df05a57a91d425935b6e81565b6000610d9f600460008473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060000154610d91610cfa565b61166290919063ffffffff16565b9050919050565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610e36576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610e2d9061257a565b60405180910390fd5b610e4a610e4430858561168e565b82611612565b73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff1614610eb7576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610eae9061261a565b60405180910390fd5b81600460008573ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600201819055508273ffffffffffffffffffffffffffffffffffffffff167f7b816003a769eb718bd9c66bdbd2dd5827da3f92bc6645276876bd7957b08cf083604051610f44919061269a565b60405180910390a2505050565b60055481565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461107657610fe6610fe03087866116e6565b82611612565b73ffffffffffffffffffffffffffffffffffffffff16600660009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1614611075576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161106c9061265a565b60405180910390fd5b5b60006110ca600260008873ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020548561167890919063ffffffff16565b905060006110e0826110db89610d41565b61173e565b9050600061113082600460008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000015461173e565b905084821015611175576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161116c906125ba565b60405180910390fd5b60008114611234576111d281600460008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000015461167890919063ffffffff16565b600460008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000206000018190555061122d8160055461167890919063ffffffff16565b6005819055505b61128682600260008b73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020019081526020016000205461166290919063ffffffff16565b600260008a73ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020819055506112de8260035461166290919063ffffffff16565b600381905550600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb88611336888661167890919063ffffffff16565b6040518363ffffffff1660e01b8152600401611353929190612350565b602060405180830381600087803b15801561136d57600080fd5b505af1158015611381573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906113a59190612028565b6113e4576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016113db906125fa565b60405180910390fd5b600085146114db57600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1663a9059cbb33876040518363ffffffff1660e01b8152600401611449929190612350565b602060405180830381600087803b15801561146357600080fd5b505af1158015611477573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061149b9190612028565b6114da576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016114d1906125fa565b60405180910390fd5b5b3373ffffffffffffffffffffffffffffffffffffffff168773ffffffffffffffffffffffffffffffffffffffff168973ffffffffffffffffffffffffffffffffffffffff167f950494fc3642fae5221b6c32e0e45765c95ebb382a04a71b160db0843e74c99f858a8a604051611553939291906126b5565b60405180910390a48183146115aa576001600660146101000a81548160ff0219169083151502179055507f3f4449c047e11092ec54dc0751b6b4817a9162745de856c893a26e611d18ffc460405160405180910390a15b5050505050505050565b60007f7d824962dd0f01520922ea1766c987b1db570cd5db90bdba5ccf5e320607950286868686866040516020016115f1969594939291906123f4565b60405160208183030381529060405280519060200120905095945050505050565b600080611625611620611757565b6117ef565b846040516020016116379291906122fe565b604051602081830303815290604052805190602001209050611659818461185f565b91505092915050565b6000818361167091906127a3565b905092915050565b6000818361168691906127f9565b905092915050565b60007fe95f353750f192082df064ca5142d3a2d6f0bef0f3ffad66d80d8af86b7a749a8484846040516020016116c794939291906123af565b6040516020818303038152906040528051906020012090509392505050565b60007f48ebe6deff4a5ee645c01506a026031e2a945d6f41f1f4e5098ad65347492c1284848460405160200161171f94939291906123af565b6040516020818303038152906040528051906020012090509392505050565b600081831061174d578161174f565b825b905092915050565b61175f611dc6565b600046905060405180606001604052806040518060400160405280600a81526020017f436865717565626f6f6b0000000000000000000000000000000000000000000081525081526020016040518060400160405280600381526020017f312e30000000000000000000000000000000000000000000000000000000000081525081526020018281525091505090565b60007fc2f8787176b8ac6bf7215b4adcc1e069bf4ab82d9ab1df05a57a91d425935b6e82600001518051906020012083602001518051906020012084604001516040516020016118429493929190612455565b604051602081830303815290604052805190602001209050919050565b600080600061186e8585611886565b9150915061187b81611909565b819250505092915050565b6000806041835114156118c85760008060006020860151925060408601519150606086015160001a90506118bc87828585611c5a565b94509450505050611902565b6040835114156118f95760008060208501519150604085015190506118ee868383611d67565b935093505050611902565b60006002915091505b9250929050565b60006004811115611943577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b81600481111561197c577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b141561198757611c57565b600160048111156119c1577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b8160048111156119fa577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b1415611a3b576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611a32906124fa565b60405180910390fd5b60026004811115611a75577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b816004811115611aae577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b1415611aef576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611ae69061251a565b60405180910390fd5b60036004811115611b29577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b816004811115611b62577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b1415611ba3576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611b9a9061255a565b60405180910390fd5b600480811115611bdc577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b816004811115611c15577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b1415611c56576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401611c4d9061259a565b60405180910390fd5b5b50565b6000807f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08360001c1115611c95576000600391509150611d5e565b601b8560ff1614158015611cad5750601c8560ff1614155b15611cbf576000600491509150611d5e565b600060018787878760405160008152602001604052604051611ce4949392919061249a565b6020604051602081039080840390855afa158015611d06573d6000803e3d6000fd5b505050602060405103519050600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415611d5557600060019250925050611d5e565b80600092509250505b94509492505050565b60008060007f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff60001b841690506000601b60ff8660001c901c611daa91906127a3565b9050611db887828885611c5a565b935093505050935093915050565b60405180606001604052806060815260200160608152602001600081525090565b6000611dfa611df584612756565b612731565b905082815260208101848484011115611e1257600080fd5b611e1d8482856128b0565b509392505050565b600081359050611e3481612cfd565b92915050565b600081519050611e4981612d14565b92915050565b600082601f830112611e6057600080fd5b8135611e70848260208601611de7565b91505092915050565b600081359050611e8881612d2b565b92915050565b600081519050611e9d81612d2b565b92915050565b600060208284031215611eb557600080fd5b6000611ec384828501611e25565b91505092915050565b60008060008060008060c08789031215611ee557600080fd5b6000611ef389828a01611e25565b9650506020611f0489828a01611e25565b9550506040611f1589828a01611e79565b945050606087013567ffffffffffffffff811115611f3257600080fd5b611f3e89828a01611e4f565b9350506080611f4f89828a01611e79565b92505060a087013567ffffffffffffffff811115611f6c57600080fd5b611f7889828a01611e4f565b9150509295509295509295565b60008060408385031215611f9857600080fd5b6000611fa685828601611e25565b9250506020611fb785828601611e79565b9150509250929050565b600080600060608486031215611fd657600080fd5b6000611fe486828701611e25565b9350506020611ff586828701611e79565b925050604084013567ffffffffffffffff81111561201257600080fd5b61201e86828701611e4f565b9150509250925092565b60006020828403121561203a57600080fd5b600061204884828501611e3a565b91505092915050565b60006020828403121561206357600080fd5b600061207184828501611e79565b91505092915050565b60006020828403121561208c57600080fd5b600061209a84828501611e8e565b91505092915050565b6120ac8161282d565b82525050565b6120bb8161283f565b82525050565b6120ca8161284b565b82525050565b6120e16120dc8261284b565b6128f0565b82525050565b6120f08161288c565b82525050565b6000612103601883612787565b915061210e82612969565b602082019050919050565b6000612126601f83612787565b915061213182612992565b602082019050919050565b6000612149603483612787565b9150612154826129bb565b604082019050919050565b600061216c600283612798565b915061217782612a0a565b600282019050919050565b600061218f602283612787565b915061219a82612a33565b604082019050919050565b60006121b2601683612787565b91506121bd82612a82565b602082019050919050565b60006121d5602283612787565b91506121e082612aab565b604082019050919050565b60006121f8601d83612787565b915061220382612afa565b602082019050919050565b600061221b602583612787565b915061222682612b23565b604082019050919050565b600061223e602783612787565b915061224982612b72565b604082019050919050565b6000612261602983612787565b915061226c82612bc1565b604082019050919050565b6000612284602783612787565b915061228f82612c10565b604082019050919050565b60006122a7602483612787565b91506122b282612c5f565b604082019050919050565b60006122ca602883612787565b91506122d582612cae565b604082019050919050565b6122e981612875565b82525050565b6122f88161287f565b82525050565b60006123098261215f565b915061231582856120d0565b60208201915061232582846120d0565b6020820191508190509392505050565b600060208201905061234a60008301846120a3565b92915050565b600060408201905061236560008301856120a3565b61237260208301846122e0565b9392505050565b600060208201905061238e60008301846120b2565b92915050565b60006020820190506123a960008301846120c1565b92915050565b60006080820190506123c460008301876120c1565b6123d160208301866120a3565b6123de60408301856120a3565b6123eb60608301846122e0565b95945050505050565b600060c08201905061240960008301896120c1565b61241660208301886120a3565b61242360408301876120a3565b61243060608301866122e0565b61243d60808301856120a3565b61244a60a08301846122e0565b979650505050505050565b600060808201905061246a60008301876120c1565b61247760208301866120c1565b61248460408301856120c1565b61249160608301846122e0565b95945050505050565b60006080820190506124af60008301876120c1565b6124bc60208301866122ef565b6124c960408301856120c1565b6124d660608301846120c1565b95945050505050565b60006020820190506124f460008301846120e7565b92915050565b60006020820190508181036000830152612513816120f6565b9050919050565b6000602082019050818103600083015261253381612119565b9050919050565b600060208201905081810360008301526125538161213c565b9050919050565b6000602082019050818103600083015261257381612182565b9050919050565b60006020820190508181036000830152612593816121a5565b9050919050565b600060208201905081810360008301526125b3816121c8565b9050919050565b600060208201905081810360008301526125d3816121eb565b9050919050565b600060208201905081810360008301526125f38161220e565b9050919050565b6000602082019050818103600083015261261381612231565b9050919050565b6000602082019050818103600083015261263381612254565b9050919050565b6000602082019050818103600083015261265381612277565b9050919050565b600060208201905081810360008301526126738161229a565b9050919050565b60006020820190508181036000830152612693816122bd565b9050919050565b60006020820190506126af60008301846122e0565b92915050565b60006060820190506126ca60008301866122e0565b6126d760208301856122e0565b6126e460408301846122e0565b949350505050565b600060808201905061270160008301876122e0565b61270e60208301866122e0565b61271b60408301856122e0565b61272860608301846122e0565b95945050505050565b600061273b61274c565b905061274782826128bf565b919050565b6000604051905090565b600067ffffffffffffffff82111561277157612770612929565b5b61277a82612958565b9050602081019050919050565b600082825260208201905092915050565b600081905092915050565b60006127ae82612875565b91506127b983612875565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff038211156127ee576127ed6128fa565b5b828201905092915050565b600061280482612875565b915061280f83612875565b925082821015612822576128216128fa565b5b828203905092915050565b600061283882612855565b9050919050565b60008115159050919050565b6000819050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b600060ff82169050919050565b60006128978261289e565b9050919050565b60006128a982612855565b9050919050565b82818337600083830152505050565b6128c882612958565b810181811067ffffffffffffffff821117156128e7576128e6612929565b5b80604052505050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000601f19601f8301169050919050565b7f45434453413a20696e76616c6964207369676e61747572650000000000000000600082015250565b7f45434453413a20696e76616c6964207369676e6174757265206c656e67746800600082015250565b7f53696d706c65537761703a2068617264206465706f7369742063616e6e6f742060008201527f6265206d6f7265207468616e2062616c616e6365000000000000000000000000602082015250565b7f1901000000000000000000000000000000000000000000000000000000000000600082015250565b7f45434453413a20696e76616c6964207369676e6174757265202773272076616c60008201527f7565000000000000000000000000000000000000000000000000000000000000602082015250565b7f53696d706c65537761703a206e6f742069737375657200000000000000000000600082015250565b7f45434453413a20696e76616c6964207369676e6174757265202776272076616c60008201527f7565000000000000000000000000000000000000000000000000000000000000602082015250565b7f53696d706c65537761703a2063616e6e6f74207061792063616c6c6572000000600082015250565b7f53696d706c65537761703a206465706f736974206e6f74207965742074696d6560008201527f64206f7574000000000000000000000000000000000000000000000000000000602082015250565b7f53696d706c65537761703a2053696d706c65537761703a207472616e7366657260008201527f206661696c656400000000000000000000000000000000000000000000000000602082015250565b7f53696d706c65537761703a20696e76616c69642062656e65666963696172792060008201527f7369676e61747572650000000000000000000000000000000000000000000000602082015250565b7f53696d706c65537761703a2068617264206465706f736974206e6f742073756660008201527f66696369656e7400000000000000000000000000000000000000000000000000602082015250565b7f53696d706c65537761703a20696e76616c696420697373756572207369676e6160008201527f7475726500000000000000000000000000000000000000000000000000000000602082015250565b7f53696d706c65537761703a206c697175696442616c616e6365206e6f7420737560008201527f6666696369656e74000000000000000000000000000000000000000000000000602082015250565b612d068161282d565b8114612d1157600080fd5b50565b612d1d8161283f565b8114612d2857600080fd5b50565b612d3481612875565b8114612d3f57600080fd5b5056fea26469706673582212204355f428ef5fddd964a622cd7d2a475b6b3f5c13731629987a8e8068f02f7ad564736f6c63430008020033a2646970667358221220cefffad5bf1323c2dbb6ca8cb58176f95a58ca66ac31f0031c4a37ea19cdd77564736f6c63430008020033"
