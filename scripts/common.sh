#!/usr/bin/env bash

# -u: Treat unset variables as an error when substituting.
# -o pipefail: The return value of a pipeline is the status of the last command to exit with a non-zero status, or zero if no command exited with a non-zero status.
set -uo pipefail
trap 's=$?; error "$0: Error on line "$LINENO": $BASH_COMMAND"; exit $s' ERR

SEDA_CHAIN_ID=${SEDA_CHAIN_ID:-"seda-1-local"}
SEDA_CHAIN_RPC=${SEDA_CHAIN_RPC:-"http://127.0.0.1:26657"}
SEDA_BINARY=${SEDA_BINARY:-$(git rev-parse --show-toplevel)/build/sedad}
TXN_GAS_FLAGS=${TXN_GAS_FLAGS:-"--gas-prices 10000000000aseda --gas auto --gas-adjustment 1.5"}
KEYRING_BACKEND=${KEYRING_BACKEND:-"test"}
# Prints an error message to stderr
# Usage: error "Error message"
error() {
	usage "${FUNCNAME[0]}" 1 "$#" "ERROR_MESSAGE"
	printf "\033[31mERROR: \033[0m%s\n" "$1" >&2
}

# Checks if the script/function was given the correct number of arguments
# Usage: usage CALLER EXPECTED_NUM_OF_ARGS NUM_OF_ARGS ARG_NAME_1 ARG_NAME_2 ... ARG_NAME_N
usage() {
	local caller=$1
	local expected_num_args=$2
	local actual_num_args=$3 # This will be the count of arguments actually passed to the calling function.
	shift 3
	if [ "${actual_num_args}" -ne "${expected_num_args}" ]; then
		echo -n "Usage: $caller"
		while (("$#")); do
			echo -n " <$1>"
			shift
		done
		echo
		exit 1
	fi
}

# Checks if the script/function was given the correct number of arguments variadic
# Usage: usage CALLER MINUM_ARGS NUM_OF_ARGS ARG_NAME_1 ARG_NAME_2 ... ARG_NAME_N
usage_variadic() {
	caller=$1
	min_args=$2
	actual_args=$3
	shift 3
	if [ "${actual_args}" -lt "${min_args}" ]; then
		echo -n "Usage: $caller"
		while (("$#")); do
			echo -n " <$1>"
			shift
		done
		echo
		exit 1
	fi
}


# Checks if command(s) exists on the sytem
# Usage: check_commands COMMAND_NAME_1 COMMAND_NAME_2 ... COMMAND_NAME_N
check_commands() {
	usage_variadic "${FUNCNAME[0]}" 1 "$#" "COMMAND_NAMES"
	local command_names=("$@")
	local command_unset=false
	for command_name in "${command_names[@]}"; do
		if ! command -v ${command_name} >/dev/null 2>&1; then
			error "Command \`${command_name}\` not found."
			command_unset=true
		fi
	done
	[ "$command_unset" = "true" ] && exit 1

	return 0
}

seda_wasm_query() {
	usage "${FUNCNAME[0]}" 1 "$#" "QUERY_MSG"
	check_commands $SEDA_BINARY
	local OUTPUT="$(${SEDA_BINARY} query wasm contract-state smart ${SEDA_CORE_CONTRACT_ADDRESS} "$1" --node ${SEDA_CHAIN_RPC} --output json)"
	echo $OUTPUT
}

# Executes an arbitrary message on a contract and outputs the transaction hash
# Usage: wasm_execute TARGET_CONTRACT EXECUTE_MSG FROM AMOUNT
# Requires: SEDA_BINARY, jq, SEDA_CHAIN_RPC, SEDA_CHAIN_ID
seda_wasm_execute() {
	usage_variadic "${FUNCNAME[0]}" 1 "$#" "EXECUTE_MSG" ["FROM"] ["AMOUNT"]
	check_commands $SEDA_BINARY jq

	local EXECUTE_MSG=$1
	local FROM=${2:-}
	local AMOUNT=${3:-}

	# Build args conditionally
	local extra_args=()
	[ -n "$FROM" ] && extra_args+=(--from "$FROM")
	[ -n "$AMOUNT" ] && extra_args+=(--amount "${AMOUNT}seda")

	local EXECUTE
	if ! EXECUTE="$(${SEDA_BINARY} tx wasm execute "${SEDA_CORE_CONTRACT_ADDRESS}" "$EXECUTE_MSG" \
  --node "${SEDA_CHAIN_RPC}" ${TXN_GAS_FLAGS} --keyring-backend "${KEYRING_BACKEND}" -y \
  --chain-id "${SEDA_CHAIN_ID}" "${extra_args[@]}")"; then
		error "Execute failed"; return 1
	fi

	local OUTPUT
	OUTPUT="$(printf '%s\n' "$EXECUTE" | ${SEDA_BINARY} query wait-tx --node "$SEDA_CHAIN_RPC" --output json)"
	echo "$OUTPUT"
}

# A general query to the seda chain not wasm specific
seda_query() {
	usage "${FUNCNAME[0]}" 2 "$#" "QUERY_TYPE" "QUERY_MSG"
	check_commands $SEDA_BINARY
	local OUTPUT="$(${SEDA_BINARY} query $1 $2 --node ${SEDA_CHAIN_RPC} --output json)"
	echo $OUTPUT
}

# a general execute function for the seda chain not wasm specific
seda_execute() {
	usage_variadic "${FUNCNAME[0]}" 2 "$#" "EXECUTE_TYPE" "EXECUTE_MSG" ["FROM"] ["AMOUNT"]
	check_commands $SEDA_BINARY

	local EXECUTE_TYPE_RAW=$1
	local EXECUTE_MSG=$2
	local FROM=${3:-}
	local AMOUNT=${4:-}

	# split into an array
	read -r -a EXECUTE_TYPE <<<"$EXECUTE_TYPE_RAW"

	# Build args conditionally
	local extra_args=()
	[ -n "$FROM" ] && extra_args+=(--from "$FROM")
	[ -n "$AMOUNT" ] && extra_args+=(--amount "${AMOUNT}seda")

	local UPLOAD
	if ! UPLOAD="$(${SEDA_BINARY} tx "${EXECUTE_TYPE[@]}" "$EXECUTE_MSG" \
		--node "$SEDA_CHAIN_RPC" ${TXN_GAS_FLAGS} --keyring-backend "$KEYRING_BACKEND" -y \
		--chain-id "$SEDA_CHAIN_ID" "${extra_args[@]}")"; then
		error "Upload failed"; return 1
	fi

	local OUTPUT
	OUTPUT="$(printf '%s\n' "$UPLOAD" | ${SEDA_BINARY} query wait-tx --node "$SEDA_CHAIN_RPC" --output json)"
	echo "$OUTPUT"
}

fetch_core_contract_address() {
	usage "${FUNCNAME[0]}" 0 "$#"
	check_commands jq
	seda_query wasm-storage core-contract-registry | jq -r '.address'
}

SEDA_CORE_CONTRACT_ADDRESS=${SEDA_CORE_CONTRACT_ADDRESS:-$(fetch_core_contract_address)}

seda_store_oracle_program() {
	usage "${FUNCNAME[0]}" 2 "$#" "PATH_TO_ORACLE_PROGRAM" "ACCOUNT"
	check_commands $SEDA_BINARY
	local OUTPUT=$(seda_execute "wasm-storage store-oracle-program" "$1" "$2")
	local PROGRAM_HASH=$(echo $OUTPUT | jq -r '.events[] | select(.type=="store_oracle_program") | .attributes[] | select(.key=="oracle_program_hash") | .value')
	echo $PROGRAM_HASH
}

seda_list_oracle_programs() {
	usage "${FUNCNAME[0]}" 0 "$#"
	check_commands jq
	seda_query wasm-storage list-oracle-programs | jq -r '.list[]'
}

seda_query_committing_data_requests() {
	usage "${FUNCNAME[0]}" 0 "$#"
	check_commands jq
	seda_wasm_query '{"get_data_requests_by_status":{"status": "committing", "limit": 10}}'
}

seda_get_account_address() {
	usage "${FUNCNAME[0]}" 1 "$#" "ACCOUNT"
	check_commands $SEDA_BINARY jq
	local OUTPUT="$(${SEDA_BINARY} keys show "$1" --output json --keyring-backend ${KEYRING_BACKEND} | jq -r '.address')"
	echo "$OUTPUT"
}

seda_add_to_allowlist() {
	usage "${FUNCNAME[0]}" 2 "$#" "FROM" "PUBLIC_KEY"
	check_commands $SEDA_BINARY

	local ADD
	if ! ADD="$(seda_wasm_execute '{"add_to_allowlist":{"public_key":"'"$2"'"}}' "$1" 2>&1)"; then
		printf "%s\n" "$ADD" >&2
		error "Failed to add public_key to allowlist"
		return 1
	fi

	echo "$ADD"
}

seda_executor_stake() {
	usage "${FUNCNAME[0]}" 2 "$#" "FROM" "AMOUNT"
	check_commands $SEDA_BINARY

	local FROM=$1
	local AMOUNT=$2

	local STAKE
	if ! STAKE="$(seda_wasm_execute '{"stake":{"amount":"'"$AMOUNT"'","from":"'"$FROM"'"}}' "$FROM" 2>&1)"; then
		printf "%s\n" "$STAKE" >&2
		error "Failed to stake"
		return 1
	fi

	echo "$STAKE"
}

seda_post_data_request() {
  usage "${FUNCNAME[0]}" 4 "$#" "FROM" "PROGRAM_ID" "INPUTS_STR" "MEMO"
  check_commands $SEDA_BINARY jq base64
  set -o noglob
  local FROM=$1 PID=$2 INPUTS_STR=$3 MEMO=$4

  # encode inputs string to base64
  local INPUTS_B64
  INPUTS_B64="$(printf '%s' "$INPUTS_STR" | base64 -w0)"

	local MEMO_B64; MEMO_B64="$(printf '%s' "$MEMO" | base64 -w0)"

  local MSG
  MSG="$(jq -n --arg pid "$PID" --arg in "$INPUTS_B64" --arg memo "$MEMO_B64" '{
    post_data_request:{
      posted_dr:{
        version:"0.1.0",
        exec_program_id:$pid,
        exec_inputs:"",
        exec_gas_limit:300000000000000,
        tally_program_id:$pid,
        tally_inputs:$in,
        tally_gas_limit:50000000000000,
        replication_factor:1,
        consensus_filter:"",
        gas_price:"2000",
        memo:$memo,
      },
      seda_payload:"",
      payback_address:""
    }
  }')"

  local OUTPUT="$(seda_wasm_execute "$MSG" "$FROM" "1000" | jq -r '.events[] | select(.type=="wasm-seda-data-request") | .attributes[] | select(.key=="dr_id") | .value')"
	echo "$OUTPUT"
}

# post 15 data requests
# sad can't do in parallel unless we have 15 accounts prepared :/
for i in {1..15}; do
  seda_post_data_request "satoshi" "5f3b31bff28c64a143119ee6389d62e38767672daace9c36db54fa2d18e9f391" "" "Data request $i"
done
wait

# for 30 seconds query committing data requests
curl -s 'http://127.0.0.1:6060/debug/pprof/profile?seconds=30' > cpu.pb.gz &

curl_pid=$!
end=$((SECONDS+30))
while [ $SECONDS -lt $end ]; do
  seda_query_committing_data_requests
done
wait $curl_pid

# helpful pprof commands
# GMC=$(go env GOMODCACHE)
# go tool pprof -source_path="$GMC" cpu.pb.gz
# (pprof) focus=github.com/CosmWasm/wasmvm/v2
# (pprof) top -cum
