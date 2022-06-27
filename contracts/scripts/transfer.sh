package abi

#!/usr/bin/env bash
set -e
set -o errtrace
set -o pipefail
set -u
#set -x

. common.sh

[ -z ${BACKEND+x} ] && BACKEND=http://localhost:8545
[ -z ${TOKEN_ADDRESS+x} ] && TOKEN_ADDRESS=
[ -z ${ACCOUNTS+x} ] && ACCOUNTS="bf4f9637c281ddfb1fbd3be5a1dae6531d408f11 c45d64d8f9642a604db93c59fd38492b262391ca"

PRIMARY_ACCOUNT=$(primary_account)
echo found primary account $PRIMARY_ACCOUNT >&2

for ACCOUNT in ACCOUNTS
do
  echo "sending tokens for $ACCOUNT" >&2
  transfer_erc20 $TOKEN_ADDRESS 0x$NODEACCOUNT 1000000000000000000 > /dev/null &
  echo "sending bnb to $ACCOUNT" >&2
  wait_for_tx $(FROM=$PRIMARY_ACCOUNT TO=0x$NODEACCOUNT VALUE=100000000000000000 eth_sendTransaction) > /dev/null &
done

wait