package abi

const PledgepABI = `[
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_stakeTokenAddr",
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
				"indexed": true,
				"internalType": "address",
				"name": "previousOwner",
				"type": "address"
			},
			{
				"indexed": true,
				"internalType": "address",
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "OwnershipTransferred",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "Slash",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "Stake",
		"type": "event"
	},
	{
		"anonymous": false,
		"inputs": [
			{
				"indexed": true,
				"internalType": "address",
				"name": "from",
				"type": "address"
			},
			{
				"indexed": false,
				"internalType": "uint256",
				"name": "amount",
				"type": "uint256"
			}
		],
		"name": "Unstake",
		"type": "event"
	},
	{
		"inputs": [
			{
				"internalType": "address",
				"name": "_addr",
				"type": "address"
			}
		],
		"name": "getShare",
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
				"name": "_addr",
				"type": "address"
			}
		],
		"name": "getSlash",
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
		"name": "owner",
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
		"name": "renounceOwnership",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "_amount",
				"type": "uint256"
			}
		],
		"name": "slash",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "_amount",
				"type": "uint256"
			}
		],
		"name": "stake",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [],
		"name": "totalShare",
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
		"name": "totalSlash",
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
				"name": "newOwner",
				"type": "address"
			}
		],
		"name": "transferOwnership",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{
				"internalType": "uint256",
				"name": "_amount",
				"type": "uint256"
			}
		],
		"name": "unStake",
		"outputs": [],
		"stateMutability": "nonpayable",
		"type": "function"
	}
]`

const PledgepBin = "0x60806040523480156200001157600080fd5b506040516200109638038062001096833981810160405281019062000037919062000182565b620000576200004b6200009f60201b60201c565b620000a760201b60201c565b80600160006101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff16021790555050620001fc565b600033905090565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b6000815190506200017c81620001e2565b92915050565b6000602082840312156200019557600080fd5b6000620001a5848285016200016b565b91505092915050565b6000620001bb82620001c2565b9050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b620001ed81620001ae565b8114620001f957600080fd5b50565b610e8a806200020c6000396000f3fe608060405234801561001057600080fd5b506004361061009e5760003560e01c80635d3eea91116100665780635d3eea911461015b578063715018a6146101775780638da5cb5b14610181578063a694fc3a1461019f578063f2fde38b146101bb5761009e565b8063026c4207146100a35780630adfc683146100c15780631ab2ae7c146100f157806345bc4d101461010f5780634b3ab9c51461012b575b600080fd5b6100ab6101d7565b6040516100b89190610bef565b60405180910390f35b6100db60048036038101906100d691906109b5565b6101e1565b6040516100e89190610bef565b60405180910390f35b6100f961022a565b6040516101069190610bef565b60405180910390f35b61012960048036038101906101249190610a07565b610234565b005b610145600480360381019061014091906109b5565b61037e565b6040516101529190610bef565b60405180910390f35b61017560048036038101906101709190610a07565b6103c7565b005b61017f6105bb565b005b6101896105cf565b6040516101969190610afd565b60405180910390f35b6101b960048036038101906101b49190610a07565b6105f8565b005b6101d560048036038101906101d091906109b5565b6107a8565b005b6000600354905090565b6000600460008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b6000600554905090565b61023c61082c565b600460003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020548111156102be576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016102b590610bcf565b60405180910390fd5b80600460003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825461030d9190610c71565b9250508190555080600560008282546103269190610c1b565b925050819055503373ffffffffffffffffffffffffffffffffffffffff167fa69f22d963cb7981f842db8c1aafcc93d915ba2a95dcf26dcc333a9c2a09be26826040516103739190610bef565b60405180910390a250565b6000600260008373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020549050919050565b600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002054811115610449576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161044090610b4f565b60405180910390fd5b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd3033846040518463ffffffff1660e01b81526004016104a893929190610b18565b602060405180830381600087803b1580156104c257600080fd5b505af11580156104d6573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906104fa91906109de565b5080600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff168152602001908152602001600020600082825461054a9190610c71565b9250508190555080600360008282546105639190610c71565b925050819055503373ffffffffffffffffffffffffffffffffffffffff167f85082129d87b2fe11527cb1b3b7a520aeb5aa6913f88a3d8757fe40d1db02fdd826040516105b09190610bef565b60405180910390a250565b6105c361082c565b6105cd60006108aa565b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff16905090565b600160009054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff166323b872dd3330846040518463ffffffff1660e01b815260040161065793929190610b18565b602060405180830381600087803b15801561067157600080fd5b505af1158015610685573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906106a991906109de565b6106e8576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016106df90610baf565b60405180910390fd5b80600260003373ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff16815260200190815260200160002060008282546107379190610c1b565b9250508190555080600360008282546107509190610c1b565b925050819055503373ffffffffffffffffffffffffffffffffffffffff167febedb8b3c678666e7f36970bc8f57abf6d8fa2e828c0da91ea5b75bf68ed101a8260405161079d9190610bef565b60405180910390a250565b6107b061082c565b600073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161415610820576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161081790610b6f565b60405180910390fd5b610829816108aa565b50565b61083461096e565b73ffffffffffffffffffffffffffffffffffffffff166108526105cf565b73ffffffffffffffffffffffffffffffffffffffff16146108a8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161089f90610b8f565b60405180910390fd5b565b60008060009054906101000a900473ffffffffffffffffffffffffffffffffffffffff169050816000806101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508173ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff167f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e060405160405180910390a35050565b600033905090565b60008135905061098581610e0f565b92915050565b60008151905061099a81610e26565b92915050565b6000813590506109af81610e3d565b92915050565b6000602082840312156109c757600080fd5b60006109d584828501610976565b91505092915050565b6000602082840312156109f057600080fd5b60006109fe8482850161098b565b91505092915050565b600060208284031215610a1957600080fd5b6000610a27848285016109a0565b91505092915050565b610a3981610ca5565b82525050565b6000610a4c601f83610c0a565b9150610a5782610d1c565b602082019050919050565b6000610a6f602683610c0a565b9150610a7a82610d45565b604082019050919050565b6000610a92602083610c0a565b9150610a9d82610d94565b602082019050919050565b6000610ab5601583610c0a565b9150610ac082610dbd565b602082019050919050565b6000610ad8601f83610c0a565b9150610ae382610de6565b602082019050919050565b610af781610ce3565b82525050565b6000602082019050610b126000830184610a30565b92915050565b6000606082019050610b2d6000830186610a30565b610b3a6020830185610a30565b610b476040830184610aee565b949350505050565b60006020820190508181036000830152610b6881610a3f565b9050919050565b60006020820190508181036000830152610b8881610a62565b9050919050565b60006020820190508181036000830152610ba881610a85565b9050919050565b60006020820190508181036000830152610bc881610aa8565b9050919050565b60006020820190508181036000830152610be881610acb565b9050919050565b6000602082019050610c046000830184610aee565b92915050565b600082825260208201905092915050565b6000610c2682610ce3565b9150610c3183610ce3565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03821115610c6657610c65610ced565b5b828201905092915050565b6000610c7c82610ce3565b9150610c8783610ce3565b925082821015610c9a57610c99610ced565b5b828203905092915050565b6000610cb082610cc3565b9050919050565b60008115159050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f554e5354414b455f414d4f554e545f4d5553545f4c4553535f53484152455300600082015250565b7f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160008201527f6464726573730000000000000000000000000000000000000000000000000000602082015250565b7f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572600082015250565b7f5354414b455f414d4f554e545f4d5553545f4841440000000000000000000000600082015250565b7f554e534c4153485f414d4f554e545f4d5553545f4c4553535f53484152455300600082015250565b610e1881610ca5565b8114610e2357600080fd5b50565b610e2f81610cb7565b8114610e3a57600080fd5b50565b610e4681610ce3565b8114610e5157600080fd5b5056fea2646970667358221220181b48618772ba1c8db4a9045a717dac266b26bcaa15ed4eeaadfc5fa9b770e664736f6c63430008020033"