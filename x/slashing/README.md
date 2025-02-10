# Slashing Module

## Overview
The slashing module wraps the SDK’s default slashing module to gatekeep the active validator set from validators who have not registered their public keys of activated proving schemes. Specifically, it prevents jailed validators from unjailing themselves to enter the active validator set without registering the required public keys first.
