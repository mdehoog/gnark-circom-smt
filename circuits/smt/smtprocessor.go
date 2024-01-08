package smt

import (
	"github.com/consensys/gnark/frontend"

	"github.com/liyue201/gnark-circomlib/circuits"
)

func SMTProcessor(api frontend.API, oldRoot frontend.Variable, siblings []frontend.Variable, oldKey, oldValue, isOld0, newKey, newValue, fnc0, fnc1 frontend.Variable) (newRoot frontend.Variable) {
	levels := len(siblings)
	//    signal input oldRoot;
	//    signal output newRoot;
	//    signal input siblings[nLevels];
	//    signal input oldKey;
	//    signal input oldValue;
	//    signal input isOld0;
	//    signal input newKey;
	//    signal input newValue;
	//    signal input fnc[2];
	//
	//    signal enabled;
	//
	//    var i;
	//
	//    enabled <== fnc[0] + fnc[1] - fnc[0]*fnc[1];
	enabled := api.Sub(api.Add(fnc0, fnc1), api.Mul(fnc0, fnc1))
	//
	//    component hash1Old = SMTHash1();
	//    hash1Old.key <== oldKey;
	//    hash1Old.value <== oldValue;
	hash1Old := SMTHash1(api, oldKey, oldValue)
	//
	//    component hash1New = SMTHash1();
	//    hash1New.key <== newKey;
	//    hash1New.value <== newValue;
	hash1New := SMTHash1(api, newKey, newValue)
	//
	//    component n2bOld = Num2Bits_strict();
	//    component n2bNew = Num2Bits_strict();
	//
	//    n2bOld.in <== oldKey;
	//    n2bNew.in <== newKey;
	n2bOld := circuits.Num2BitsStrict(api, oldKey, 254)
	n2bNew := circuits.Num2BitsStrict(api, newKey, 254)
	//
	//    component smtLevIns = SMTLevIns(nLevels);
	//    for (i=0; i<nLevels; i++) smtLevIns.siblings[i] <== siblings[i];
	//    smtLevIns.enabled <== enabled;
	smtLevIns := SMTLevIns(api, enabled, siblings)
	//
	//    component xors[nLevels];
	xors := make([]frontend.Variable, levels)
	//    for (i=0; i<nLevels; i++) {
	//        xors[i] = XOR();
	//        xors[i].a <== n2bOld.out[i];
	//        xors[i].b <== n2bNew.out[i];
	//    }
	for i := 0; i < levels; i++ {
		xors[i] = circuits.Xor(api, n2bOld[i], n2bNew[i])
	}
	//
	//    component sm[nLevels];
	st_top := make([]frontend.Variable, levels)
	st_old0 := make([]frontend.Variable, levels)
	st_bot := make([]frontend.Variable, levels)
	st_new1 := make([]frontend.Variable, levels)
	st_na := make([]frontend.Variable, levels)
	st_upd := make([]frontend.Variable, levels)
	//    for (i=0; i<nLevels; i++) {
	//        sm[i] = SMTProcessorSM();
	//        if (i==0) {
	//            sm[i].prev_top <== enabled;
	//            sm[i].prev_old0 <== 0;
	//            sm[i].prev_bot <== 0;
	//            sm[i].prev_new1 <== 0;
	//            sm[i].prev_na <== 1-enabled;
	//            sm[i].prev_upd <== 0;
	//        } else {
	//            sm[i].prev_top <== sm[i-1].st_top;
	//            sm[i].prev_old0 <== sm[i-1].st_old0;
	//            sm[i].prev_bot <== sm[i-1].st_bot;
	//            sm[i].prev_new1 <== sm[i-1].st_new1;
	//            sm[i].prev_na <== sm[i-1].st_na;
	//            sm[i].prev_upd <== sm[i-1].st_upd;
	//        }
	//        sm[i].is0 <== isOld0;
	//        sm[i].xor <== xors[i].out;
	//        sm[i].fnc[0] <== fnc[0];
	//        sm[i].fnc[1] <== fnc[1];
	//        sm[i].levIns <== smtLevIns.levIns[i];
	//    }
	for i := 0; i < levels; i++ {
		if i == 0 {
			st_top[i], st_old0[i], st_bot[i], st_new1[i], st_na[i], st_upd[i] = SMTProcessorSM(api, xors[i], isOld0, smtLevIns[i], fnc0, enabled, 0, 0, 0, api.Sub(1, enabled), 0)
		} else {
			st_top[i], st_old0[i], st_bot[i], st_new1[i], st_na[i], st_upd[i] = SMTProcessorSM(api, xors[i], isOld0, smtLevIns[i], fnc0, st_top[i-1], st_old0[i-1], st_bot[i-1], st_new1[i-1], st_na[i-1], st_upd[i-1])
		}
	}

	//    sm[nLevels-1].st_na + sm[nLevels-1].st_new1 + sm[nLevels-1].st_old0 +sm[nLevels-1].st_upd === 1;
	api.AssertIsEqual(api.Add(api.Add(st_na[levels-1], st_new1[levels-1]), api.Add(st_old0[levels-1], st_upd[levels-1])), 1)
	//
	//    component levels[nLevels];
	levelsOldRoot := make([]frontend.Variable, levels)
	levelsNewRoot := make([]frontend.Variable, levels)
	//    for (i=nLevels-1; i != -1; i--) {
	//        levels[i] = SMTProcessorLevel();
	//
	//        levels[i].st_top <== sm[i].st_top;
	//        levels[i].st_old0 <== sm[i].st_old0;
	//        levels[i].st_bot <== sm[i].st_bot;
	//        levels[i].st_new1 <== sm[i].st_new1;
	//        levels[i].st_na <== sm[i].st_na;
	//        levels[i].st_upd <== sm[i].st_upd;
	//
	//        levels[i].sibling <== siblings[i];
	//        levels[i].old1leaf <== hash1Old.out;
	//        levels[i].new1leaf <== hash1New.out;
	//
	//        levels[i].newlrbit <== n2bNew.out[i];
	//        if (i==nLevels-1) {
	//            levels[i].oldChild <== 0;
	//            levels[i].newChild <== 0;
	//        } else {
	//            levels[i].oldChild <== levels[i+1].oldRoot;
	//            levels[i].newChild <== levels[i+1].newRoot;
	//        }
	//    }
	for i := levels - 1; i != -1; i-- {
		if i == levels-1 {
			levelsOldRoot[i], levelsNewRoot[i] = SMTProcessorLevel(api, st_top[i], st_old0[i], st_bot[i], st_new1[i], st_upd[i], siblings[i], hash1Old, hash1New, n2bNew[i], 0, 0)
		} else {
			levelsOldRoot[i], levelsNewRoot[i] = SMTProcessorLevel(api, st_top[i], st_old0[i], st_bot[i], st_new1[i], st_upd[i], siblings[i], hash1Old, hash1New, n2bNew[i], levelsOldRoot[i+1], levelsNewRoot[i+1])
		}

	}
	//
	//    component topSwitcher = Switcher();
	//
	//    topSwitcher.sel <== fnc[0]*fnc[1];
	//    topSwitcher.L <== levels[0].oldRoot;
	//    topSwitcher.R <== levels[0].newRoot;
	topSwitcherL, topSwitcherR := circuits.Switcher(api, api.Mul(fnc0, fnc1), levelsOldRoot[0], levelsNewRoot[0])
	//
	//    component checkOldInput = ForceEqualIfEnabled();
	//    checkOldInput.enabled <== enabled;
	//    checkOldInput.in[0] <== oldRoot;
	//    checkOldInput.in[1] <== topSwitcher.outL;
	circuits.ForceEqualIfEnabled(api, oldRoot, topSwitcherL, enabled)
	//
	//    newRoot <== enabled * (topSwitcher.outR - oldRoot) + oldRoot;
	newRoot = api.Add(api.Mul(enabled, api.Sub(topSwitcherR, oldRoot)), oldRoot)
	//
	////    topSwitcher.outL === oldRoot*enabled;
	////    topSwitcher.outR === newRoot*enabled;
	//
	//    // Ckeck keys are equal if updating
	//    component areKeyEquals = IsEqual();
	//    areKeyEquals.in[0] <== oldKey;
	//    areKeyEquals.in[1] <== newKey;
	areKeyEquals := circuits.IsEqual(api, oldKey, newKey)
	//
	//    component keysOk = MultiAND(3);
	//    keysOk.in[0] <== 1-fnc[0];
	//    keysOk.in[1] <== fnc[1];
	//    keysOk.in[2] <== 1-areKeyEquals.out;
	in := []frontend.Variable{
		api.Sub(1, fnc0),
		fnc1,
		api.Sub(1, areKeyEquals),
	}
	keysOk := circuits.MultiAnd(api, in)
	//
	//    keysOk.out === 0;
	api.AssertIsEqual(keysOk, 0)
	return
}
