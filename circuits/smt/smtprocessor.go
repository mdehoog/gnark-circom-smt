package smt

import (
	"github.com/consensys/gnark/frontend"

	"github.com/liyue201/gnark-circomlib/circuits"
)

func SMTProcessor(api frontend.API, oldRoot frontend.Variable, siblings []frontend.Variable, oldKey, oldValue, isOld0, newKey, newValue, fnc0, fnc1 frontend.Variable) (newRoot frontend.Variable) {
	levels := len(siblings)
	enabled := api.Sub(api.Add(fnc0, fnc1), api.Mul(fnc0, fnc1))
	hash1Old := SMTHash1(api, oldKey, oldValue)
	hash1New := SMTHash1(api, newKey, newValue)
	n2bOld := circuits.Num2BitsStrict(api, oldKey, 254)
	n2bNew := circuits.Num2BitsStrict(api, newKey, 254)
	smtLevIns := SMTLevIns(api, enabled, siblings)

	xors := make([]frontend.Variable, levels)
	for i := 0; i < levels; i++ {
		xors[i] = circuits.Xor(api, n2bOld[i], n2bNew[i])
	}

	st_top := make([]frontend.Variable, levels)
	st_old0 := make([]frontend.Variable, levels)
	st_bot := make([]frontend.Variable, levels)
	st_new1 := make([]frontend.Variable, levels)
	st_na := make([]frontend.Variable, levels)
	st_upd := make([]frontend.Variable, levels)
	for i := 0; i < levels; i++ {
		if i == 0 {
			st_top[i], st_old0[i], st_bot[i], st_new1[i], st_na[i], st_upd[i] = SMTProcessorSM(api, xors[i], isOld0, smtLevIns[i], fnc0, enabled, 0, 0, 0, api.Sub(1, enabled), 0)
		} else {
			st_top[i], st_old0[i], st_bot[i], st_new1[i], st_na[i], st_upd[i] = SMTProcessorSM(api, xors[i], isOld0, smtLevIns[i], fnc0, st_top[i-1], st_old0[i-1], st_bot[i-1], st_new1[i-1], st_na[i-1], st_upd[i-1])
		}
	}

	api.AssertIsEqual(api.Add(api.Add(st_na[levels-1], st_new1[levels-1]), api.Add(st_old0[levels-1], st_upd[levels-1])), 1)

	levelsOldRoot := make([]frontend.Variable, levels)
	levelsNewRoot := make([]frontend.Variable, levels)
	for i := levels - 1; i != -1; i-- {
		if i == levels-1 {
			levelsOldRoot[i], levelsNewRoot[i] = SMTProcessorLevel(api, st_top[i], st_old0[i], st_bot[i], st_new1[i], st_upd[i], siblings[i], hash1Old, hash1New, n2bNew[i], 0, 0)
		} else {
			levelsOldRoot[i], levelsNewRoot[i] = SMTProcessorLevel(api, st_top[i], st_old0[i], st_bot[i], st_new1[i], st_upd[i], siblings[i], hash1Old, hash1New, n2bNew[i], levelsOldRoot[i+1], levelsNewRoot[i+1])
		}
	}

	topSwitcherL, topSwitcherR := circuits.Switcher(api, api.Mul(fnc0, fnc1), levelsOldRoot[0], levelsNewRoot[0])
	circuits.ForceEqualIfEnabled(api, oldRoot, topSwitcherL, enabled)

	newRoot = api.Add(api.Mul(enabled, api.Sub(topSwitcherR, oldRoot)), oldRoot)

	areKeyEquals := circuits.IsEqual(api, oldKey, newKey)
	in := []frontend.Variable{
		api.Sub(1, fnc0),
		fnc1,
		api.Sub(1, areKeyEquals),
	}
	keysOk := circuits.MultiAnd(api, in)
	api.AssertIsEqual(keysOk, 0)
	return
}
