#!/usr/bin/env bash

function to_hex() {
  printf '0x%x\n' $1
}

function to_abi_hex() {
  printf "%0$2x" "$1" | tr ' ' '0'
}

function to_abi_address() {
  left_pad $(echo "$1" | cut -c 3-) 64
}

function left_pad() {
  printf "%0$2s" "$1" | tr ' ' '0'
}

function jsonrpc_raw() {
  curl -s -X POST -H "Content-Type: application/json" --data "{\"jsonrpc\":\"2.0\",\"method\":\"$1\",\"params\":$2, \"id\":\"1\" }" "$BACKEND"
}

function jsonrpc() {
  local ret
  ret=$(jsonrpc_raw "$1" "$2")
  if test $? -eq 7
  then
    echo could not connect to backend >&2
    return 2
  fi

  if echo "$ret" | jq -e .error > /dev/null
  then
    echo "$ret" | jq -e .error >&2
    return 1
  else
    echo "$ret" | jq -e .result
  fi
}

function eth_accounts() {
  jsonrpc eth_accounts "[]"
}

function primary_account() {
  eth_accounts | jq -r '.[0]'
}

function eth_sendTransaction() {
  local args=''
  [ ! -z ${FROM+x} ] && args="$args --arg from $FROM"
  [ ! -z ${TO+x} ] && args="$args --arg to $TO"
  [ ! -z ${VALUE+x} ] && args="$args --arg value $(to_hex $VALUE)"
  [ ! -z ${DATA+x} ] && args="$args --arg data $DATA"
  [ ! -z ${GAS+x} ] && args="$args --arg gas $(to_hex $GAS)"
  jsonrpc eth_sendTransaction "$(jq -n $args '[. + $ARGS.named'])" | jq -r
}

function wait_for_tx() {
  local txhash="$1"
  while
    local receipt=$(jsonrpc eth_getTransactionReceipt "[\"$txhash\"]")
    [ "$receipt" == "null" ]
  do
    sleep 1
  done
  echo "$receipt"
}

function wait_for_deploy() {
  wait_for_tx "$@" | jq -r .contractAddress
}

TRANSFER_TOKEN_DATA=0xa9059cbb__RECIPIENT____AMOUNT__
function transfer_erc20() {
  local data=$(echo $TRANSFER_TOKEN_DATA | sed -e "s.__RECIPIENT__.$(to_abi_address $2)." -e "s.__AMOUNT__.$(to_abi_hex $3 64).")
  wait_for_tx $(FROM=$(primary_account) TO="$1" DATA="$data" eth_sendTransaction)
}

TOKEN_MINT_DATA=0x40c10f19__RECIPIENT____AMOUNT__
function mint_erc20() {
  local data=$(echo $TOKEN_MINT_DATA | sed -e "s.__RECIPIENT__.$(to_abi_address $2)." -e "s.__AMOUNT__.$(to_abi_hex $3 64).")
  wait_for_tx $(FROM=$(primary_account) TO="$1" DATA="$data" eth_sendTransaction)
}

function grantPriceOracleRole() {
  wait_for_tx $(FROM=$(primary_account) TO="$1" DATA="0x2f2ff15ddd24a0f121e5ab7c3e97c63eaaf859e0b46792c3e0edfd86e2b3ad50f63011d8$(to_abi_address $2)" eth_sendTransaction)
}

SET_PRICE_DATA=0x91b7f5ed__AMOUNT__
function setPrice() {
    local data=$(echo $SET_PRICE_DATA | sed -e "s.__AMOUNT__.$(to_abi_hex $2 64).")
    wait_for_tx $(FROM=$(primary_account) TO="$1" DATA="$data" eth_sendTransaction)
}