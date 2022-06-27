// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./Ownable.sol";
import "./IERC20.sol";

contract Pledge is Ownable {

    IERC20 stakeToken;
    mapping(address => uint256) private shares;
    uint256 totalShares;
    mapping(address => uint256) private slashes;
    uint256 totalSlashes;

    event Stake(address indexed from, uint256 amount);
    event Unstake(address indexed from, uint256 amount);
    event Slash(address indexed from, uint256 amount);

    constructor(address _stakeTokenAddr){
        stakeToken = IERC20(_stakeTokenAddr);
    }

    function stake(uint256 _amount) external
    {
        stakeToken.transferFrom(msg.sender, address(this), _amount);
        shares[msg.sender] += _amount;
        totalShares += _amount;
        emit Stake(msg.sender, _amount);
    }

    function unStake(uint256 _amount) external
    {
        require(_amount <= shares[msg.sender], "UNSTAKE_AMOUNT_MUST_LESS_SHARES");
        stakeToken.transferFrom(address(this), msg.sender, _amount);
        shares[msg.sender] -= _amount;
        totalShares -= _amount;
        emit Unstake(msg.sender, _amount);
    }

    function getShare(address _addr)
        external
        view
        returns(uint256)
    {
        return shares[_addr];
    }

    function slash(uint256 _amount) external onlyOwner
    {
        require(_amount <= shares[msg.sender], "UNSLASH_AMOUNT_MUST_LESS_SHARES");
        shares[msg.sender] -= _amount;
        totalSlashes += _amount;
        emit Slash(msg.sender, _amount);
    }

    function getSlash(address _addr)
        external
        view
        returns(uint256)
    {
        return slashes[_addr];
    }

}