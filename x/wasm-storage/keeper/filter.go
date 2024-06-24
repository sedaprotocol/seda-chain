package keeper

const (
	filterNone   byte = 0x00
	filterMode   byte = 0x01
	filterStdDev byte = 0x02
)

// ApplyFilter processes filter of the type specified in the first byte of
// consensus filter. It returns an outlier list, which is a boolean list where
// true at index i means that the reveal at index i is an outlier, consensus
// boolean, and error.
func ApplyFilter(filter []byte, reveals []RevealBody) ([]int, bool, error) {
	if len(filter) < 1 {
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}

	switch filter[0] {
	case filterNone:
		return make([]int, len(reveals)), true, nil

	// TODO: Reactivate mode filter
	// case filterMode:
	// 	return nil, false, errors.New("filter type mode is not implemented")

	// TODO: Reactivate standard deviation filter
	// case filterStdDev:
	// 	return nil, false, errors.New("filter type standard deviation is not implemented")

	default:
		outliers := make([]int, len(reveals))
		for i := range outliers {
			outliers[i] = 1
		}
		return outliers, false, nil
	}
}
