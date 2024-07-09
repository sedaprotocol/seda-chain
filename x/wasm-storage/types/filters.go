package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"unsafe"

	"github.com/tidwall/gjson"
)

type Filter interface {
	DecodeFilterInput(input []byte) error
	ApplyFilter(reveals []RevealBody) ([]int, bool, error)
}

const (
	b             byte   = 0
	u             uint64 = 0
	ByteLenFilter        = int(unsafe.Sizeof(b))
	ByteLenUint64        = int(unsafe.Sizeof(u))
)

type FilterMode struct {
	dataPath string // JSON path to reveal data
}

// Mode filter input looks as follows:
// 0             1                  9       9+data_path_length
// | filter_type | data_path_length |   data_path   |
func (f *FilterMode) DecodeFilterInput(input []byte) error {
	if len(input) < 9 {
		return ErrInvalidFilterInputLen.Wrapf("%d < %d", len(input), 9)
	}

	var pathLen uint64
	err := binary.Read(bytes.NewReader(input[1:9]), binary.BigEndian, &pathLen)
	if err != nil {
		return err
	}

	path := input[9:]
	if len(path) != int(pathLen) {
		return ErrInvalidPathLen.Wrapf("expected: %d got: %d", int(pathLen), len(path))
	}
	f.dataPath = string(path)
	return nil
}

// ApplyFilter takes in a list of reveals. It returns an outlier list,
// a consensus boolean, and an error if applicable.
// Value of 1 at index i in the outlier list indicates that the i-th
// reveal is an outlier. Value of 0 indicates a non-outlier reveal.
// A reveal is an outlier if it has less than or equal to 2/3 frequency.
// The consensus boolean is true if one of the following criteria is met:
// 1. More than 1/3 of the reveals are corrupted (non-zero exit code,
// invalid data path, etc.)
// 2. More than 2/3 of the reveals are identical.
func (f FilterMode) ApplyFilter(reveals []RevealBody) ([]int, bool, error) {
	var maxFreq, corruptCount int
	freq := make(map[string]int, len(reveals))
	dataList := make([]string, len(reveals))
	for i, r := range reveals {
		if r.ExitCode != 0 {
			corruptCount++
			continue
		}

		// Extract the reveal data to track frequency.
		revealBytes, err := base64.StdEncoding.DecodeString(r.Reveal)
		if err != nil {
			corruptCount++
			continue
		}
		res := gjson.GetBytes(revealBytes, f.dataPath)
		if !res.Exists() {
			corruptCount++
			continue
		}

		data := res.String()
		freq[data]++
		maxFreq = max(freq[data], maxFreq)
		dataList[i] = data
	}

	outliers := make([]int, len(reveals))
	for i := 0; i < len(outliers); i++ {
		outliers[i] = 1
	}

	// If more than 1/3 of the reveals are corrupted,
	// we reach consensus that the reveals are unusable.
	if corruptCount*3 > len(reveals) {
		return outliers, true, nil
	}

	// If less than 2/3 of the reveals match the max frequency,
	// there is no consensus.
	if maxFreq*3 < len(reveals)*2 {
		return outliers, false, nil
	}

	for i, r := range dataList {
		if freq[r] == maxFreq {
			outliers[i] = 0
		}
	}
	return outliers, true, nil
}

type FilterStdDev struct {
	maxSigma   uint64
	numberType byte
	dataPath   string // JSON path to reveal data
}

// Standard deviation filter input looks as follows:
// 0             1           9             10                 18    18+json_path_length
// | filter_type | max_sigma | number_type | json_path_length | json_path |
func (f *FilterStdDev) DecodeFilterInput(input []byte) error {
	if len(input) < 18 {
		return ErrInvalidFilterInputLen.Wrapf("%d < %d", len(input), 18)
	}

	var maxSigma uint64
	err := binary.Read(bytes.NewReader(input[1:9]), binary.BigEndian, &maxSigma)
	if err != nil {
		return err
	}
	f.maxSigma = maxSigma

	f.numberType = input[9]

	var pathLen uint64
	err = binary.Read(bytes.NewReader(input[10:18]), binary.BigEndian, &pathLen)
	if err != nil {
		return err
	}

	path := input[18:]
	if len(path) != int(pathLen) {
		return ErrInvalidPathLen.Wrapf("expected: %d got: %d", int(pathLen), len(path))
	}
	f.dataPath = string(path)
	return nil
}

// TODO
func (f FilterStdDev) ApplyFilter(reveals []RevealBody) ([]int, bool, error) {
	outliers := make([]int, len(reveals))
	for i := 0; i < len(outliers); i++ {
		outliers[i] = 1
	}

	return outliers, true, nil
}
