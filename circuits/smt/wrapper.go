package smt

import (
	"math/big"
)

// Wrapper defines methods for wrapping existing SMT implementations, useful for
// generating circuit assignments for generating proof witnesses. See WrapperArbo
// for a concrete example that wrappers the arbo.Tree implementation.
type Wrapper interface {
	Get(key *big.Int) (*big.Int, error)
	Add(key, value *big.Int) (Assignment, error)
	Update(key, value *big.Int) (Assignment, error)
	Set(key, value *big.Int) (Assignment, error)
	Proof(key, value *big.Int) (Assignment, error)
}

type Assignment struct {
	Fnc0     uint8
	Fnc1     uint8
	OldKey   *big.Int
	NewKey   *big.Int
	IsOld0   uint8
	OldValue *big.Int
	NewValue *big.Int
	OldRoot  *big.Int
	NewRoot  *big.Int
	Siblings []*big.Int
}
