package poseidon

import (
	"math/big"
	"reflect"

	"github.com/consensys/gnark-crypto/field/pool"
)

// CheckBigIntInField checks if given *big.Int fits in a Field Q element
func CheckBigIntInField[E Element[E]](a *big.Int) bool {
	m := newElement[E]()
	modulus := m.Sub(m, newElement[E]().SetOne()).BigInt(big.NewInt(0))
	modulus.Add(modulus, big.NewInt(1))
	return a.Cmp(modulus) == -1
}

// CheckBigIntArrayInField checks if given *big.Int fits in a Field Q element
func CheckBigIntArrayInField[E Element[E]](arr []*big.Int) bool {
	for _, a := range arr {
		if !CheckBigIntInField[E](a) {
			return false
		}
	}
	return true
}

// BigIntArrayToElementArray converts an array of *big.Int into an array of *ff.Element
func BigIntArrayToElementArray[E Element[E]](bi []*big.Int) []E {
	o := make([]E, len(bi))
	for i := range bi {
		o[i] = newElement[E]().SetBigInt(bi[i])
	}
	return o
}

// Exp is a copy of gnark-crypto's implementation, but takes a pointer argument
func Exp[E Element[E]](z, x E, k *big.Int) {
	if k.IsUint64() && k.Uint64() == 0 {
		z.SetOne()
	}

	e := k
	if k.Sign() == -1 {
		// negative k, we invert
		// if k < 0: xᵏ (mod q) == (x⁻¹)ᵏ (mod q)
		x.Inverse(x)

		// we negate k in a temp big.Int since
		// Int.Bit(_) of k and -k is different
		e = pool.BigInt.Get()
		defer pool.BigInt.Put(e)
		e.Neg(k)
	}

	z.Set(x)

	for i := e.BitLen() - 2; i >= 0; i-- {
		z.Square(z)
		if e.Bit(i) == 1 {
			z.Mul(z, x)
		}
	}
}

func newElement[E Element[E]]() E {
	typ := reflect.TypeOf((*E)(nil)).Elem()
	val := reflect.New(typ.Elem())
	return val.Interface().(E)
}
