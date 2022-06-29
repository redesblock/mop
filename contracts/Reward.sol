// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import "./ERC20.sol";
import "./IERC20.sol";
import "./Ownable.sol";

contract Reward is Ownable {
    uint256 public systemTotalAmount;
    address public systemToken;
    mapping(address => uint256) _systemRecord;

    uint256 public totalAmount;
    address public token;
    mapping(address => uint256) _unCashRecord;
    mapping(address => uint256) _cashedRecord;

    constructor(address _systemToken, address _token) {
        systemToken = _systemToken;
        token = _token;
    }

    function doSystemToken(address[] memory _tos, uint256[] memory _values) external onlyOwner returns (bool)  {
        require(_tos.length > 0, "invalid tos");
        require(_values.length > 0, "invalid values");
        require(_values.length == _tos.length, "mismatch");

        for(uint i=0;i<_tos.length;i++){
            require(_tos[i] != address(0), "invalid address");
            require(_values[i] > 0);
            require(IERC20(systemToken).transferFrom(address(this), _tos[i], _values[i]), "failed to transfer");
            _systemRecord[_tos[i]] += _values[i];
            systemTotalAmount += _values[i];
        }
        return true;
    }

    function doToken(address[] memory _tos, uint256[] memory _values) external onlyOwner returns (bool) {
        require(_tos.length > 0, "invalid tos");
        require(_values.length > 0, "invalid values");
        require(_values.length == _tos.length, "mismatch");

        for(uint i=0;i<_tos.length;i++){
            require(_tos[i] != address(0), "invalid address");
            require(_values[i] > 0);
            _unCashRecord[_tos[i]] += _values[i];
            totalAmount += _values[i];
        }
        return true;
    }

    function cash(uint256 _values) external {
        require(_values <= _unCashRecord[msg.sender], "insufficient balance");
        require(IERC20(token).transferFrom(address(this), msg.sender, _values), "failed to transfer");
        _unCashRecord[msg.sender] -= _values;
        _cashedRecord[msg.sender] += _values;
    }

    function systemReward(address _addr) external view returns(uint256) {
        return _systemRecord[_addr];
    }

    function cashedReward(address _addr) external view returns(uint256) {
        return _cashedRecord[_addr];
    }

    function unCashReward(address _addr) external view returns(uint256) {
        return _unCashRecord[_addr];
    }
}