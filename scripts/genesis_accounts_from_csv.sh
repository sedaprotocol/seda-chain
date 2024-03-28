#!/usr/bin/env bash

# Grab the directory where this script is located
SCRIPTS_DIR="$(dirname "$0")"
# Source common functions
source "${SCRIPTS_DIR}/seda-scripts/common.sh"

# Does add-genesis-account for each entry in a CSV file.
# Usage: genesis_accounts_from_csv <CSV_FILE> <ERROR_LOG>
# Requires: sedad

usage "$0" 2 "$#" "CSV_FILE" "ERROR_LOG"

CSV_FILE=$1
ERROR_LOG=$2

# Clear the error log file at the beginning of the script run
> "$ERROR_LOG"

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

  # Add conditional parameters
	# If the vesting amount is not empty, neither should the vesting start time and end time
	if [ -n "$vesting_amount" ] && ([ -z "$vesting_start_time" ] || [ -z "$vesting_end_time" ]); then
		echo "Failed to process address: $address"
		echo "$address,$amount,$vesting_amount,$vesting_start_time,$vesting_end_time,$funder_addr" >> "$ERROR_LOG"
		echo "Vesting amount requires both vesting start time and vesting end time" >> "$ERROR_LOG"
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