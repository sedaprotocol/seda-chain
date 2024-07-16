package types

import (
	"bytes"
	"encoding/binary"

	"golang.org/x/exp/constraints"
)

// Sigma is a 10^6 precision fixed-point unsigned number represented
//
//	by a uint64, whose last 6 digits represent the fractional part.
type Sigma uint64

func NewSigma(data []byte) (Sigma, error) {
	var s Sigma
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &s)
	if err != nil {
		return s, err
	}
	return s, nil
}

// WholeNumber returns Sigma's whole number part.
func (s Sigma) WholeNumber() uint64 {
	return uint64(s) / 1e6
}

// FractionalPart returns Sigma's fractional part as a uint64.
func (s Sigma) FractionalPart() uint64 {
	return uint64(s) % 1e6
}

// HalfStepInt is an integer type with half-step increments (0.5).
type HalfStepInt[T constraints.Integer] struct {
	integer  T
	halfStep bool // if true, the number contains fractional part (0.5)
}

func NewHalfStepInt[T constraints.Integer](integer T, halfStep bool) HalfStepInt[T] {
	return HalfStepInt[T]{
		integer:  integer,
		halfStep: halfStep,
	}
}

// Mid sets h to the middle point between the two integers x and y
// and returns h.
func (h *HalfStepInt[T]) Mid(x, y T) *HalfStepInt[T] {
	h.integer = (x + y) / 2
	h.halfStep = false
	// Set h's halfStep to true if the earlier division operation has
	// truncted the result.
	if (x%2 == 0 && y%2 != 0) ||
		(x%2 != 0 && y%2 == 0) {
		h.halfStep = true
	}
	return h
}

// IsWithinSigma returns true if and only if the integer x is within
// the maxSigma range from the halfStepInt h. That is, IsWithinSigma
// returns true if and only if the absolute difference between x and h
// is less than or equal to maxSigma.
func (h HalfStepInt[T]) IsWithinSigma(x T, maxSigma Sigma) bool {
	var absDiff uint64 // truncated absolute difference between h and x
	switch {
	case h.integer >= x:
		absDiff = uint64(h.integer - x)
		// If h's halfStep direction and truncation direction match,
		// absDiff must be decremented by one.
		// Note if absDiff is zero, we do not adjust absDiff because
		// the truncation direction changes.
		if h.integer < 0 && h.halfStep && absDiff > 0 {
			absDiff--
		}
	case h.integer < x:
		absDiff = uint64(x - h.integer)
		// Note absDiff > 0 check is unnecessary.
		if h.integer >= 0 && h.halfStep {
			absDiff--
		}
	}

	if absDiff > maxSigma.WholeNumber() {
		return false
	}
	if absDiff == maxSigma.WholeNumber() && h.halfStep {
		// If we reach here, it means that absDiff = int(maxSigma) + 0.5.
		// Therefore, we check that maxSigma's fractional part is
		// greater than or equal to 0.5.
		return maxSigma.FractionalPart() >= 5e5
	}
	return true
}
