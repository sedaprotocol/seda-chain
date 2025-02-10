# Vesting Module

## Overview
The purpose of the vesting module is to support a custom vesting account type `ClawbackContinuousVestingAccount` that wraps the SDK’s `ContinuousVestingAccount` to provide an additional feature of “clawing back” funds. That is, the funder of the vesting account can at any point during vesting send a clawback transaction to terminate the vesting stream and retrieve the remaining vesting funds, which are taken from the following sources in order:
1. Vesting funds that have not been used towards delegation
2. Delegations
3. Unbonding delegations
