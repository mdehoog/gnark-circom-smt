package smt

import (
	"github.com/consensys/gnark/frontend"

	"github.com/liyue201/gnark-circomlib/circuits"
)

func SMTProcessorLevel(api frontend.API, st_top, st_old0, st_bot, st_new1, st_upd, sibling, old1leaf, new1leaf, newlrbit, oldChild, newChild frontend.Variable) (oldRoot, newRoot frontend.Variable) {
	//    signal input st_top;
	//    signal input st_old0;
	//    signal input st_bot;
	//    signal input st_new1;
	//    signal input st_na;
	//    signal input st_upd;
	//
	//    signal output oldRoot;
	//    signal output newRoot;
	//    signal input sibling;
	//    signal input old1leaf;
	//    signal input new1leaf;
	//    signal input newlrbit;
	//    signal input oldChild;
	//    signal input newChild;
	//
	//    signal aux[4];
	//
	//    component oldProofHash = SMTHash2();
	//    component newProofHash = SMTHash2();
	//
	//    component oldSwitcher = Switcher();
	//    component newSwitcher = Switcher();
	//
	//    // Old side
	//
	//    oldSwitcher.L <== oldChild;
	//    oldSwitcher.R <== sibling;
	//
	//    oldSwitcher.sel <== newlrbit;
	//    oldProofHash.L <== oldSwitcher.outL;
	//    oldProofHash.R <== oldSwitcher.outR;
	oldProofHashL, oldProofHashR := circuits.Switcher(api, newlrbit, oldChild, sibling)
	oldProofHash := SMTHash2(api, oldProofHashL, oldProofHashR)
	//
	//    aux[0] <== old1leaf * (st_bot + st_new1 + st_upd);
	//    oldRoot <== aux[0] +  oldProofHash.out * st_top;
	oldRoot = api.Add(api.Mul(old1leaf, api.Add(api.Add(st_bot, st_new1), st_upd)), api.Mul(oldProofHash, st_top))
	//
	//    // New side
	//
	//    aux[1] <== newChild * ( st_top + st_bot);
	//    newSwitcher.L <== aux[1] + new1leaf*st_new1;
	//
	//    aux[2] <== sibling*st_top;
	//    newSwitcher.R <== aux[2] + old1leaf*st_new1;
	//
	//    newSwitcher.sel <== newlrbit;
	//    newProofHash.L <== newSwitcher.outL;
	//    newProofHash.R <== newSwitcher.outR;
	newProofHashL, newProofHashR := circuits.Switcher(api, newlrbit, api.Add(api.Mul(newChild, api.Add(st_top, st_bot)), api.Mul(new1leaf, st_new1)), api.Add(api.Mul(sibling, st_top), api.Mul(old1leaf, st_new1)))
	newProofHash := SMTHash2(api, newProofHashL, newProofHashR)
	//
	//    aux[3] <== newProofHash.out * (st_top + st_bot + st_new1);
	//    newRoot <==  aux[3] + new1leaf * (st_old0 + st_upd);
	newRoot = api.Add(api.Mul(newProofHash, api.Add(api.Add(st_top, st_bot), st_new1)), api.Mul(new1leaf, api.Add(st_old0, st_upd)))
	return
}
