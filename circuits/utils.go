package circuits

import (
	"fmt"
	"math/big"

	"github.com/consensys/gnark/frontend"
)

func IsEqual(api frontend.API, a frontend.Variable, b frontend.Variable) frontend.Variable {
	return api.IsZero(api.Sub(a, b))
}

func ForceEqualIfEnabled(api frontend.API, a, b, enabled frontend.Variable) {
	c := api.IsZero(api.Sub(a, b))
	api.AssertIsEqual(api.Mul(api.Sub(1, c), enabled), 0)
}

func MultiAnd(api frontend.API, in []frontend.Variable) frontend.Variable {
	out := frontend.Variable(1)
	for i := 0; i < len(in); i++ {
		out = api.And(out, in[i])
	}
	return out
}

func Lsh(k int64, n uint) *big.Int {
	z := big.NewInt(k)
	return z.Lsh(z, n)
}

func BigRsh(k *big.Int, n uint) *big.Int {
	return big.NewInt(0).Rsh(k, n)
}

func BigAnd(x *big.Int, y *big.Int) *big.Int {
	return big.NewInt(0).And(x, y)
}

func BigSub(x *big.Int, y *big.Int) *big.Int {
	return big.NewInt(0).Sub(x, y)
}

func Switcher(api frontend.API, sel, l, r frontend.Variable) (frontend.Variable, frontend.Variable) {
	aux := api.Mul(api.Sub(r, l), sel)

	outL := api.Add(aux, l)
	outR := api.Sub(r, aux)

	return outL, outR
}

func Num2BitsStrict(api frontend.API, in frontend.Variable, n int) []frontend.Variable {
	bits := api.ToBinary(in, n)
	AliasCheck(api, bits)
	return bits
}

func AliasCheck(api frontend.API, in []frontend.Variable) {
	q := api.Compiler().Field()
	q.Sub(q, big.NewInt(1))
	api.AssertIsEqual(CompConstant(api, in, q), 0)
}

// CompConstant returns 1 if in (in binary) > ct
func CompConstant(api frontend.API, in []frontend.Variable, ct *big.Int) frontend.Variable {
	if len(in) != 254 {
		panic(fmt.Sprintf("CompConstant: invalid len: %v", len(in)))
	}
	var (
		parts [127]frontend.Variable
		clsb  *big.Int
		cmsb  *big.Int
	)

	sum := frontend.Variable(0)
	b := frontend.Variable(BigSub(Lsh(1, 128), big.NewInt(1)))
	a := frontend.Variable(1)
	e := frontend.Variable(1)

	for i := 0; i < 127; i++ {
		clsb = BigAnd(BigRsh(ct, uint(i<<1)), big.NewInt(1))
		cmsb = BigAnd(BigRsh(ct, uint((i<<1)+1)), big.NewInt(1))
		slsb := in[i<<1]
		smsb := in[(i<<1)+1]
		if cmsb.Int64() == 0 && clsb.Int64() == 0 {
			parts[i] = api.Add(api.Mul(-1, b, smsb, slsb), api.Mul(b, smsb), api.Mul(b, slsb))
		} else if cmsb.Int64() == 0 && clsb.Int64() == 1 {
			parts[i] = api.Add(api.Mul(a, smsb, slsb), api.Mul(-1, a, slsb), api.Mul(b, smsb), api.Mul(-1, a, smsb), a)
		} else if cmsb.Int64() == 1 && clsb.Int64() == 0 {
			parts[i] = api.Add(api.Mul(b, smsb, slsb), api.Mul(-1, a, smsb), a)
		} else {
			parts[i] = api.Add(api.Mul(-1, a, smsb, slsb), a)
		}
		sum = api.Add(sum, parts[i])
		b = api.Sub(b, e)
		a = api.Add(a, e)
		e = api.Mul(e, 2)
	}
	bits := api.ToBinary(sum, 135)

	return bits[127]
}
