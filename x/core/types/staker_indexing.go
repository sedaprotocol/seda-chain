package types

import (
	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// StakerIndexing is used to store staker objects while maintaining their indices.
type StakerIndexing struct {
	collections.Map[string, Staker]
	indexToKey collections.Map[uint32, string]
	keyToIndex collections.Map[string, uint32]
	count      collections.Item[uint32]
}

func NewStakerIndexing(sb *collections.SchemaBuilder, cdc codec.BinaryCodec) StakerIndexing {
	return StakerIndexing{
		Map:        collections.NewMap(sb, StakersKeyPrefix, "stakers", collections.StringKey, codec.CollValue[Staker](cdc)),
		indexToKey: collections.NewMap(sb, StakerIndexToKeyPrefix, "staker_index_to_key", collections.Uint32Key, collections.StringValue),
		keyToIndex: collections.NewMap(sb, StakerKeyToIndexPrefix, "staker_key_to_index", collections.StringKey, collections.Uint32Value),
		count:      collections.NewItem(sb, StakerCountKey, "staker_count", collections.Uint32Value),
	}
}

func (s StakerIndexing) Set(ctx sdk.Context, pubKey string, staker Staker) error {
	// Simple update if staker already exists.
	exists, err := s.Has(ctx, pubKey)
	if err != nil {
		return err
	}
	if exists {
		return s.Map.Set(ctx, pubKey, staker)
	}

	// In case of a new staker, assign the current count as the index
	// and increment the count.
	count, err := s.count.Get(ctx)
	if err != nil {
		return err
	}

	err = s.indexToKey.Set(ctx, count, pubKey)
	if err != nil {
		return err
	}
	err = s.keyToIndex.Set(ctx, pubKey, count)
	if err != nil {
		return err
	}
	err = s.Map.Set(ctx, pubKey, staker)
	if err != nil {
		return err
	}
	return s.count.Set(ctx, count+1)
}

func (s StakerIndexing) Get(ctx sdk.Context, pubKey string) (Staker, error) {
	return s.Map.Get(ctx, pubKey)
}

func (s StakerIndexing) Remove(ctx sdk.Context, pubKey string) error {
	count, err := s.count.Get(ctx)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrUnexpectedZeroCount
	}

	// Identify and verify the index of the given public key.
	index, err := s.keyToIndex.Get(ctx, pubKey)
	if err != nil {
		return err
	}
	foundPubKey, err := s.indexToKey.Get(ctx, index)
	if err != nil {
		return err
	}
	if foundPubKey != pubKey {
		return ErrUnexpectedStakerKey.Wrapf("index: %d, public key: %s", index, pubKey)
	}

	if count == 1 || index == count-1 {
		// Handle special case of removing the only item or item at last index.
		err = s.indexToKey.Remove(ctx, index)
		if err != nil {
			return err
		}
	} else {
		// Swap last index with the given index and remove the last index.
		lastIndex := count - 1
		lastKey, err := s.indexToKey.Get(ctx, lastIndex)
		if err != nil {
			return err
		}

		err = s.indexToKey.Set(ctx, index, lastKey)
		if err != nil {
			return err
		}
		err = s.keyToIndex.Set(ctx, lastKey, index)
		if err != nil {
			return err
		}

		err = s.indexToKey.Remove(ctx, lastIndex)
		if err != nil {
			return err
		}
	}

	// Remove the given staker and decrement the count.
	err = s.keyToIndex.Remove(ctx, pubKey)
	if err != nil {
		return err
	}
	err = s.Map.Remove(ctx, pubKey)
	if err != nil {
		return err
	}
	return s.count.Set(ctx, count-1)
}

func (s StakerIndexing) GetStakerCount(ctx sdk.Context) (uint32, error) {
	return s.count.Get(ctx)
}

func (s StakerIndexing) SetStakerCount(ctx sdk.Context, count uint32) error {
	return s.count.Set(ctx, count)
}

func (s StakerIndexing) GetStakerIndex(ctx sdk.Context, pubKey string) (uint32, error) {
	return s.keyToIndex.Get(ctx, pubKey)
}

func (s StakerIndexing) GetStakerKey(ctx sdk.Context, index uint32) (string, error) {
	return s.indexToKey.Get(ctx, index)
}

func (s StakerIndexing) GetExecutors(ctx sdk.Context, offset, limit uint32) ([]Staker, error) {
	rng := &collections.Range[uint32]{}
	rng.StartInclusive(offset).EndExclusive(offset + limit)

	itr, err := s.indexToKey.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	defer itr.Close()

	executors := make([]Staker, 0, limit)
	for ; itr.Valid(); itr.Next() {
		stakerPubKey, err := itr.Value()
		if err != nil {
			return nil, err
		}
		staker, err := s.Get(ctx, stakerPubKey)
		if err != nil {
			return nil, err
		}
		executors = append(executors, staker)
	}
	return executors, nil
}
