#!/usr/bin/env bash

# Grab the directory where this script is located
SCRIPTS_DIR="$(dirname "$0")"
# Source common functions
source "${SCRIPTS_DIR}/seda-scripts/common.sh"

# Does add-genesis-account for each entry in a CSV file.
# Usage: genesis_accounts_from_csv <CSV_FILE> <GENESIS_FILE> <ERROR_LOG>
# Requires: sedad, jq

usage "$0" 3 "$#" "CSV_FILE" "GENESIS_FILE" "ERROR_LOG"

CSV_FILE=$1
GENESIS_FILE=$2
ERROR_LOG=$3

# Clear the error log file at the beginning of the script run
> "$ERROR_LOG"

# Check if the CSV file exists
if [ ! -f "$CSV_FILE" ]; then
	echo "CSV file not found: $CSV_FILE"
	exit 1
fi

# Check if the genesis file exists
if [ ! -f "$GENESIS_FILE" ]; then
	echo "Genesis file not found: $GENESIS_FILE"
	exit 1
fi

echo "Processing genesis accounts from $CSV_FILE"
# Skip the header line; read the rest of the lines
tail -n +2 "$CSV_FILE" | while IFS=, read -r address amount vesting_amount vesting_start_time vesting_end_time funder_addr || [ -n "$address" ]; do
  echo "Processing address: $address"

	# check address and amount are not empty
	if [ -z "$address" ] || [ -z "$amount" ]; then
		echo "Failed to process address: $address"
		echo "$address,$amount,$vesting_amount,$vesting_start_time,$vesting_end_time,$funder_addr" >> "$ERROR_LOG"
		echo "Address and amount are required fields" >> "$ERROR_LOG"
		echo "--------------------------------" >> "$ERROR_LOG"
		continue
	fi

  # Initialize command with mandatory parameters
  cmd="${SEDA_BINARY} add-genesis-account \"$address\" \"${amount}\""

	# if the vesting start is empty read the genesis file and set the vesting start time to the genesis time
	# only if the vesting amount is not empty
	if [ -n "$vesting_amount" ] && [ -z "$vesting_start_time" ]; then
		echo "No vesting start time provided, reading genesis time from the genesis file"
		# Get the genesis time from the genesis file
		genesis_time=$(jq -r '.genesis_time' "$GENESIS_FILE")
		# check if genesis time was found
		if [ -z "$genesis_time" ]; then
			echo "Failed to process address: $address"
			echo "$address,$amount,$vesting_amount,$vesting_start_time,$vesting_end_time,$funder_addr" >> "$ERROR_LOG"
			echo "Genesis time not found in the genesis file" >> "$ERROR_LOG"
			echo "--------------------------------" >> "$ERROR_LOG"
			continue
		fi
		# Set the vesting start time to the genesis time by converting it to a unix timestamp
		vesting_start_time=$(date -d "$genesis_time" +%s)
		echo "Genesis time: $vesting_start_time"
	fi

  # Add conditional parameters
	# If the vesting amount is not empty:
	# neither should the vesting start time, end time, and funder address
	if [ -z "$vesting_amount" ] && ([ -n "$vesting_start_time" ] || [ -n "$vesting_end_time" ] || [ -n "$funder_addr" ]); then
		echo "Failed to process address: $address"
		echo "$address,$amount,$vesting_amount,$vesting_start_time,$vesting_end_time,$funder_addr" >> "$ERROR_LOG"
		echo "Vesting amount requires both vesting start time, vesting end time and funder_addr" >> "$ERROR_LOG"
		echo "--------------------------------" >> "$ERROR_LOG"
		continue
	fi
  [ -n "$vesting_amount" ] && cmd+=" --vesting-amount \"${vesting_amount}\""
  [ -n "$vesting_start_time" ] && cmd+=" --vesting-start-time \"$vesting_start_time\""
  [ -n "$vesting_end_time" ] && cmd+=" --vesting-end-time \"$vesting_end_time\""
  [ -n "$funder_addr" ] && cmd+=" --funder \"$funder_addr\""

  # Execute the command and capture all output
	error_output=$(eval $cmd 2>&1)
	
  # Check for any error, if eval command fails
  if [ $? -ne 0 ]; then
    echo "Failed to process address: $address"
		echo "Command Executed: $cmd" >> "$ERROR_LOG"
    # Write the failing entry and error message to the error log
    echo "$address,$amount,$vesting_amount,$vesting_start_time,$vesting_end_time,$funder_addr" >> "$ERROR_LOG"
    # Filter for lines with "Error:" or "ERR" and remove ANSI color codes, then write to the log
		echo "$error_output" | grep -E "Error:|ERR" | sed 's/\x1b\[[0-9;]*m//g' >> "$ERROR_LOG"
    echo "--------------------------------" >> "$ERROR_LOG"
  fi

done

echo "All addresses processed."