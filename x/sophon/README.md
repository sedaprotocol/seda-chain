# Sophon Module

Sophon is SEDA's trust-based solution for high-speed execution of Oracle Programs. A user pre-pays `aseda` and receives credits which can be spent on Oracle Program execution and data proxy access.

The Sophon module is responsible for storing the configurations of Sophons, allowing admins to manage Sophons and the users of the Sophon, keeping track of credits, verifying Sophon eligibility, and distributing Sophon data proxy payouts.

## Overview

> [!INFO]
> The module supports multiple Sophons, but initially only a single Sophon operated by the SEDA core team will be used.

The Sophon module provides the following key functionality:

1. **Sophon Registration & Management**
   - Registering new Sophons with their public keys and admin addresses
   - Allowing owners to edit Sophon details and transfer ownership
   - Allowing admins to manage Sophon users and credits

1. **Report Submission**
   - Sophons periodically submit execution reports containing Oracle Program results
   - Valid reports trigger credit deductions and data proxy payouts

1. **Credit System**
   - Users can top up credits by depositing `aseda` tokens
   - Credits are spent when users request Oracle Program executions
   - Admins can settle credits through withdrawing or burning tokens

The module allows for implementing a permissioned system where trusted Sophons can provide fast Oracle Program execution services while data proxies can still receive payouts for providing access.

## Request Flow

The typical flow for requesting Oracle Program execution through a Sophon is:

1. **User Setup**
   - User gets added to a Sophon in the module by the Sophon's admin
   - User tops up their credits by depositing `aseda` tokens

2. **Execution Request**
   - User sends a request directly to the Sophon endpoint
   - Sophon verifies the user has sufficient credits
   - Sophon executes the Oracle Program and returns result to user
   - User receives fast response without waiting for consensus

3. **Report Submission**
   - Sophon periodically submits execution reports on-chain
   - Reports contain details of Oracle Program executions
   - Module deducts user credits
   - Data proxies receive payouts for their program access

4. **Credit Settlement**
   - Admin can settle accumulated credits by:
     - Withdrawing tokens to specified address
     - Burning tokens to reduce supply
   - Expired credits can be reclaimed by the admin

This flow enables fast off-chain execution while maintaining economic incentives and accountability through periodic on-chain settlement.

## State

```
0x00                       -> parameters
0x01                       -> sophon_id_sequence
0x02 | pub_key             -> sophon_info
0x03 | sophon_id | user_id -> sophon_user
0x04 | sophon_id | address -> transfer (set)
```

We use the `sophon_id` for the SophonUser key to make public key rotation for the Sophon easier. Now only the SophonInfo has to be moved to a new store entry and we don't need to iterate over all the users to update them. Since most queries and transactions that use the SophonUser also require the SophonInfo this does not introduce unnecessary store access.
