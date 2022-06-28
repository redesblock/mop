#!/usr/bin/env bash
set -e
set -o errtrace
set -o pipefail
set -u
#set -x

. common.sh

[ -z ${BACKEND+x} ] && BACKEND=http://localhost:8575
[ -z ${TOKEN_ADDRESS+x} ] && TOKEN_ADDRESS=
[ -z ${ACCOUNTS+x} ] && ACCOUNTS="bf4f9637c281ddfb1fbd3be5a1dae6531d408f11 c45d64d8f9642a604db93c59fd38492b262391ca"

PRIMARY_ACCOUNT=$(primary_account)
echo found primary account $PRIMARY_ACCOUNT >&2

for ACCOUNT in $ACCOUNTS
do
  transfer_erc20 $TOKEN_ADDRESS 0x$ACCOUNT 1000000000000000000 > /dev/null &
  echo "sending tokens for $ACCOUNT" >&2
done

wait