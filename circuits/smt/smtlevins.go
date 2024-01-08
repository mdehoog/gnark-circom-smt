package smt

import "github.com/consensys/gnark/frontend"

func SMTLevIns(api frontend.API, enabled frontend.Variable, siblings []frontend.Variable) (levIns []frontend.Variable) {
	levels := len(siblings)
	//    signal input enabled;
	//    signal input siblings[nLevels];
	//    signal output levIns[nLevels];
	levIns = make([]frontend.Variable, levels)
	//    signal done[nLevels-1];        // Indicates if the insLevel has aready been detected.
	done := make([]frontend.Variable, levels-1)
	//
	//    var i;
	//
	//    component isZero[nLevels];
	isZero := make([]frontend.Variable, levels)
	//
	//    for (i=0; i<nLevels; i++) {
	//        isZero[i] = IsZero();
	//        isZero[i].in <== siblings[i];
	//    }
	for i := 0; i < len(siblings); i++ {
		isZero[i] = api.IsZero(siblings[i])
	}
	//
	//    // The last level must always have a sibling of 0. If not, then it cannot be inserted.
	//    (isZero[nLevels-1].out - 1) * enabled === 0;
	api.AssertIsEqual(api.Mul(api.Sub(isZero[levels-1], 1), enabled), 0)
	//
	//    levIns[nLevels-1] <== (1-isZero[nLevels-2].out);
	levIns[levels-1] = api.Sub(1, isZero[levels-2])
	//    done[nLevels-2] <== levIns[nLevels-1];
	done[levels-2] = levIns[levels-1]
	//    for (i=nLevels-2; i>0; i--) {
	//        levIns[i] <== (1-done[i])*(1-isZero[i-1].out);
	//        done[i-1] <== levIns[i] + done[i];
	//    }
	for i := levels - 2; i > 0; i-- {
		levIns[i] = api.Mul(api.Sub(1, done[i]), api.Sub(1, isZero[i-1]))
		done[i-1] = api.Add(levIns[i], done[i])
	}
	//
	//    levIns[0] <== (1-done[0]);
	levIns[0] = api.Sub(1, done[0])
	return levIns
}
