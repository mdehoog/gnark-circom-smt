package poseidon

import (
	"math/big"

	"github.com/mdehoog/poseidon/poseidon"
	"go.vocdoni.io/dvote/tree/arbo"
)

// HashPoseidon implements the HashFunction interface for the Poseidon hash
type HashPoseidon[E poseidon.Element[E]] struct {
}

// Type returns the type of HashFunction for the HashPoseidon
func (HashPoseidon[E]) Type() []byte {
	return arbo.TypeHashPoseidon
}

// Len returns the length of the Hash output
func (HashPoseidon[E]) Len() int {
	return len(poseidon.NewElement[E]().Marshal())
}

// Hash implements the hash method for the HashFunction HashPoseidon. It
// expects the byte arrays to be little-endian representations of big.Int
// values.
func (f HashPoseidon[E]) Hash(b ...[]byte) ([]byte, error) {
	var toHash []*big.Int
	for i := 0; i < len(b); i++ {
		bi := arbo.BytesToBigInt(b[i])
		toHash = append(toHash, bi)
	}
	h, err := poseidon.Hash[E](toHash)
	if err != nil {
		return nil, err
	}
	hB := arbo.BigIntToBytes(f.Len(), h)
	return hB, nil
}
