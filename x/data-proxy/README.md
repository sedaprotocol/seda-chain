# Data Proxy Module

## Overview
The Data Proxy module is responsible for storing the configurations of data proxy providers.

## State
```
0x00 | data_proxy_pubkey -> data_proxy_config
0x01                     -> parameters
0x02 | expiration_height -> fee_updates_queue
```

### Data Proxy Configurations
Data proxy providers use their admin accounts to register and edit their configurations like payout address, public key, and fee in this module. Note the module imposes a minimum number of blocks before a fee change comes into effect to prevent abrupt fee changes.
