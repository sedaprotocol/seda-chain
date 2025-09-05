# x/fast Module

`x/fast` is SEDA's trust-based solution for high-speed execution of Oracle Programs. A user pre-pays `aseda` and receives credits which can be spent on Oracle Program execution and data proxy access.

The module is responsible for storing the configurations of Fast clients, allowing admins to manage Fast clients and the users of the Fast client, keeping track of credits, verifying Fast client eligibility, and distributing Fast client data proxy payouts.

> [!CAUTION]
> TODO: 
> - [ ] Deletion queue in keeper based on timestamp
> - [ ] Deletion processing in endblock
> - [ ] Delete TX
> - [ ] Cancel deletion
> - [ ] Disallow TopUp when deletion is pending
> - [ ] Update eligibility query to return 0 credits when deletion is pending
> - [ ] Update this README with new state layout and functionality
> - [ ] Update validation logic (move to msg handler, deduplicate logic and error messages)
> - [ ] Update balances to be separate mappings so it's more efficient when updating credits

## Overview

> [!INFO]
> The module supports multiple Fast clients, but initially only a single Fast client operated by the SEDA core team will be used.

The Fast module provides the following key functionality:

1. **Fast Client Registration & Management**
   - Registering new Fast clients with their public keys and admin addresses
   - Allowing owners to edit Fast client details and transfer ownership
   - Allowing admins to manage Fast client users and credits

1. **Report Submission**
   - Fast clients periodically submit usage reports
   - Valid reports trigger credit deductions and data proxy payouts

1. **Credit System**
   - Users can top up credits by depositing `aseda` tokens
   - Credits are spent when users request Oracle Program executions
   - Admins can settle credits through withdrawing or burning tokens

The module allows for implementing a permissioned system where trusted Fast clients can provide fast Oracle Program execution services while data proxies can still receive payouts for providing access.

## Request Flow

The typical flow for requesting Oracle Program execution through a Fast client is:

1. **User Setup**
   - User gets added to a Fast client in the module by the Fast client's admin
   - User or admin tops up their credits by depositing `aseda` tokens

2. **Execution Request**
   - User sends a request directly to the Fast client endpoint, which is an off-chain service
   - The Fast client verifies the user has sufficient credits
   - The Fast client executes the Oracle Program and returns result to user
   - User receives fast response without waiting for consensus

3. **Report Submission**
   - Fast clients periodically submits execution reports on-chain
   - Reports contain details of Oracle Program executions
   - Module deducts user credits
   - Data proxies receive payouts when they provided access to data

4. **Credit Settlement**
   - The admin can settle accumulated credits by:
     - Withdrawing tokens
     - Burning tokens to reduce supply
   - The admin can expire a Fast client users credits

This flow enables fast off-chain execution while maintaining economic incentives and accountability through periodic on-chain settlement.

## State

```
0x00                            -> parameters
0x01                            -> fast_client_id_sequence
0x02 | pub_key                  -> fast_client
0x03 | fast_client_id | user_id -> fast_client_user
0x04 | fast_client_id           -> pending_owner_address
```

We use the `fast_client_id` for the FastUser key to make public key rotation for the Fast client easier. Now only the FastClient has to be moved to a new store entry and we don't need to iterate over all the users to update them. Since most queries and transactions that use the FastUser also require the Fast client this does not introduce unnecessary store access.
