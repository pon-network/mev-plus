#!/bin/sh

# Assign proper value to _DAPPNODE_GLOBAL_EXECUTION_CLIENT.
case "$_DAPPNODE_GLOBAL_EXECUTION_CLIENT_MAINNET" in
"geth.dnp.dappnode.eth") EXECUTION_LAYER="http://geth.dappnode:8545" ;;
"nethermind.dnp.dappnode.eth") EXECUTION_LAYER="http://nethermind.dappnode:8545" ;;
"besu.dnp.dappnode.eth") EXECUTION_LAYER="http://besu.dappnode:8545" ;;
"erigon.dnp.dappnode.eth") EXECUTION_LAYER="http://erigon.dappnode:8545" ;;
*)
  echo "Unknown value for _DAPPNODE_GLOBAL_EXECUTION_CLIENT. Please confirm that the value is correct."
  exit 1
  ;;
esac

# Assign proper value to _DAPPNODE_GLOBAL_CONSENSUS_CLIENT.
case "$_DAPPNODE_GLOBAL_CONSENSUS_CLIENT_MAINNET" in
"prysm.dnp.dappnode.eth") BEACON_NODE_API="http://beacon-chain.prysm.dappnode:3500" ;;
"teku.dnp.dappnode.eth") BEACON_NODE_API="http://beacon-chain.teku.dappnode:3500" ;;
"lighthouse.dnp.dappnode.eth") BEACON_NODE_API="http://beacon-chain.lighthouse.dappnode:3500" ;;
"nimbus.dnp.dappnode.eth") BEACON_NODE_API="http://beacon-validator.nimbus.dappnode:4500" ;;
"lodestar.dnp.dappnode.eth") BEACON_NODE_API="http://beacon-chain.lodestar.dappnode:3500" ;;
*)
  echo "_DAPPNODE_GLOBAL_CONSENSUS_CLIENT env is not set properly."
  exit 1
  ;;
esac

# MEVBOOST is set up as an external proxy to MEV Plus for users that require external block building - this allows the builder API to be shared
if [ -n "$_DAPPNODE_GLOBAL_MEVBOOST_MAINNET" ] && [ "$_DAPPNODE_GLOBAL_MEVBOOST_MAINNET" == "false" ]; then
    EXTRA_OPTS="${EXTRA_OPTS} -externalValidatorProxy.address http://mev-boost.mev-boost.dappnode:18550"
fi

./mevPlus \
   -builderApi.listen-address http://0.0.0.0:18551 \
   -k2.eth1-private-key $ETH1_PRIVATE_KEY \
   -k2.beacon-node-url $BEACON_NODE_API \
   -k2.execution-node-url $EXECUTION_LAYER \
   -k2.logger-level $LOGGER_LEVEL \
   ${EXTRA_OPTS}
