// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "./Ownable.sol";
import "./IERC20.sol";

contract Pledge is Ownable {

    IERC20 public stakeToken;
    mapping(address => uint256) private _shares;
    uint256 public totalShare;
    mapping(address => uint256) private _slashes;
    uint256 public totalSlash;

    event Stake(address indexed from, uint256 amount);
    event Unstake(address indexed from, uint256 amount);
    event Slash(address indexed from, uint256 amount);

    constructor(address _stakeTokenAddr){
        stakeToken = IERC20(_stakeTokenAddr);
    }

    function stake(uint256 _amount) external
    {
        require(stakeToken.transferFrom(msg.sender, address(this), _amount), "failed to transfer");
        _shares[msg.sender] += _amount;
        totalShare += _amount;
        emit Stake(msg.sender, _amount);
    }

    function unStake(uint256 _amount) external
    {
        require(_amount <= _shares[msg.sender], "insufficient balance");
        stakeToken.transferFrom(address(this), msg.sender, _amount);
        _shares[msg.sender] -= _amount;
        totalShare -= _amount;
        emit Unstake(msg.sender, _amount);
    }

    function getShare(address _addr)
        external
        view
        returns(uint256)
    {
        return _shares[_addr];
    }

    function slash(uint256 _amount) external onlyOwner
    {
        require(_amount <= _slashes[msg.sender], "insufficient balance");
        _slashes[msg.sender] -= _amount;
        totalSlash += _amount;
        emit Slash(msg.sender, _amount);
    }

    function getSlash(address _addr)
        external
        view
        returns(uint256)
    {
        return _slashes[_addr];
    }
}