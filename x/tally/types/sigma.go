package types

import (
	"bytes"
	"encoding/binary"

	"golang.org/x/exp/constraints"
)

// Sigma is a 10^6 precision fixed-point unsigned number represented
// by a uint64, whose last 6 digits represent the fractional part.
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
// For example, the whole number part of Sigma(1_000_000) is 1.
func (s Sigma) WholeNumber() uint64 {
	return uint64(s) / 1e6
}

// FractionalPart returns Sigma's fractional part as a uint64.
// For example, the fractional part of Sigma(1_500_000) is 500_000,
// representing the fraction 500_000 / 1_000_000.
func (s Sigma) FractionalPart() uint64 {
	return uint64(s) % 1e6
}

// HalfStepInt is an integer type with half-step increments (0.5).
type HalfStepInt[T constraints.Integer] struct {
	integer  T
	neg      bool // if true, the integer is negative (to represent -0.5)
	halfStep bool // if true, the number contains fractional part (0.5)
}

// Mid sets h to the middle point between the two integers x and y
// and returns h.
func (h *HalfStepInt[T]) Mid(x, y T) *HalfStepInt[T] {
	h.integer = (x + y) / 2
	h.neg = (x + y) < 0
	h.halfStep = false
	// Set h's halfStep to true if x+y is odd, or equivalently,
	// if only one of x and y is odd.
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
	// absDiff represents the integer part of the absolute difference
	// between h and x. The true absolute difference may contain a
	// half-step fractional part (0.5), which can be inferred by h.halfStep.
	var absDiff uint64
	switch {
	case h.integer > x:
		absDiff = uint64(h.integer - x)
		// If h's halfStep direction pushes absDiff to the left,
		// absDiff must be decremented by one.
		if h.neg && h.halfStep {
			absDiff--
		}
	case h.integer < x:
		absDiff = uint64(x - h.integer)
		// If h's halfStep direction pushes absDiff to the left,
		// absDiff must be decremented by one.
		if !h.neg && h.halfStep {
			absDiff--
		}
	case h.integer == x:
		absDiff = 0
		// It is never necessary to adjust absDiff since |0.5| = |-0.5|.
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
