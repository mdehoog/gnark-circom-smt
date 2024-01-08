package smt

import (
	"github.com/consensys/gnark/frontend"

	"github.com/liyue201/gnark-circomlib/circuits"
)

func SMTProcessorLevel(api frontend.API, st_top, st_old0, st_bot, st_new1, st_upd, sibling, old1leaf, new1leaf, newlrbit, oldChild, newChild frontend.Variable) (oldRoot, newRoot frontend.Variable) {
	oldProofHashL, oldProofHashR := circuits.Switcher(api, newlrbit, oldChild, sibling)
	oldProofHash := SMTHash2(api, oldProofHashL, oldProofHashR)

	oldRoot = api.Add(api.Mul(old1leaf, api.Add(api.Add(st_bot, st_new1), st_upd)), api.Mul(oldProofHash, st_top))

	newProofHashL, newProofHashR := circuits.Switcher(api, newlrbit, api.Add(api.Mul(newChild, api.Add(st_top, st_bot)), api.Mul(new1leaf, st_new1)), api.Add(api.Mul(sibling, st_top), api.Mul(old1leaf, st_new1)))
	newProofHash := SMTHash2(api, newProofHashL, newProofHashR)

	newRoot = api.Add(api.Mul(newProofHash, api.Add(api.Add(st_top, st_bot), st_new1)), api.Mul(new1leaf, api.Add(st_old0, st_upd)))
	return
}
