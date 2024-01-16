package smt

import (
	"github.com/consensys/gnark/frontend"

	"github.com/mdehoog/gnark-circom-smt/circuits"
)

func VerifierLevel(api frontend.API, stTop, stIOld, stINew, sibling, old1leaf, new1leaf, lrbit, child frontend.Variable) (root frontend.Variable) {
	proofHashL, proofHashR := circuits.Switcher(api, lrbit, child, sibling)
	proofHash := Hash2(api, proofHashL, proofHashR)
	root = api.Add(api.Add(api.Mul(proofHash, stTop), api.Mul(old1leaf, stIOld)), api.Mul(new1leaf, stINew))
	return
}
