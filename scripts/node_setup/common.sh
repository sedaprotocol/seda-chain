
function auth_seda_chaind_command {
  local args=("$@")
  output=$(expect -c "
  spawn $BIN ${args[*]}
  expect {
    \"Enter keyring passphrase (attempt 1/3):\" {
      send \"$KEYRING_PASSWORD\r\"
    }
    timeout {
      send_user \"Timed out waiting for enter passphrase prompt\r\"
      exit 1
    }
  }
  expect eof
")
  echo "$output"
}