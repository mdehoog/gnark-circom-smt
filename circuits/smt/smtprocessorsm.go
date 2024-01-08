package smt

import "github.com/consensys/gnark/frontend"

func SMTProcessorSM(api frontend.API, xor, is0, levIns, fnc0, prev_top, prev_old0, prev_bot, prev_new1, prev_na, prev_upd frontend.Variable) (st_top, st_old0, st_bot, st_new1, st_na, st_upd frontend.Variable) {
	//  signal input xor;
	//  signal input is0;
	//  signal input levIns;
	//  signal input fnc[2];
	//
	//  signal input prev_top;
	//  signal input prev_old0;
	//  signal input prev_bot;
	//  signal input prev_new1;
	//  signal input prev_na;
	//  signal input prev_upd;
	//
	//  signal output st_top;
	//  signal output st_old0;
	//  signal output st_bot;
	//  signal output st_new1;
	//  signal output st_na;
	//  signal output st_upd;
	//
	//  signal aux1;
	//  signal aux2;
	//
	//  aux1 <==                  prev_top * levIns;
	aux1 := api.Mul(prev_top, levIns)
	//  aux2 <== aux1*fnc[0];  // prev_top * levIns * fnc[0]
	aux2 := api.Mul(aux1, fnc0)
	//
	//  // st_top = prev_top*(1-levIns)
	//  //    = + prev_top
	//  //      - prev_top * levIns            = aux1
	//
	//  st_top <== prev_top - aux1;
	st_top = api.Sub(prev_top, aux1)
	//
	//  // st_old0 = prev_top * levIns * is0 * fnc[0]
	//  //      = + prev_top * levIns * is0 * fnc[0]            = aux2 * is0
	//
	//  st_old0 <== aux2 * is0;  // prev_top * levIns * is0 * fnc[0]
	st_old0 = api.Mul(aux2, is0)
	//
	//  // st_new1 = prev_top * levIns * (1-is0)*fnc[0] * xor   +  prev_bot*xor =
	//  //    = + prev_top * levIns *       fnc[0] * xor     = aux2     * xor
	//  //      - prev_top * levIns * is0 * fnc[0] * xor     = st_old0  * xor
	//  //      + prev_bot *                         xor     = prev_bot * xor
	//
	//  st_new1 <== (aux2 - st_old0 + prev_bot)*xor;
	st_new1 = api.Mul(api.Add(api.Sub(aux2, st_old0), prev_bot), xor)
	//
	//
	//  // st_bot = prev_top * levIns * (1-is0)*fnc[0] * (1-xor) + prev_bot*(1-xor);
	//  //    = + prev_top * levIns *       fnc[0]
	//  //      - prev_top * levIns * is0 * fnc[0]
	//  //      - prev_top * levIns *       fnc[0] * xor
	//  //      + prev_top * levIns * is0 * fnc[0] * xor
	//  //      + prev_bot
	//  //      - prev_bot *                         xor
	//
	//  st_bot <== (1-xor) * (aux2 - st_old0 + prev_bot);
	st_bot = api.Mul(api.Sub(1, xor), api.Add(api.Sub(aux2, st_old0), prev_bot))
	//
	//
	//  // st_upd = prev_top * (1-fnc[0]) *levIns;
	//  //    = + prev_top * levIns
	//  //      - prev_top * levIns * fnc[0]
	//
	//  st_upd <== aux1 - aux2;
	st_upd = api.Sub(aux1, aux2)
	//
	//  // st_na = prev_new1 + prev_old0 + prev_na + prev_upd;
	//  //    = + prev_new1
	//  //      + prev_old0
	//  //      + prev_na
	//  //      + prev_upd
	//
	//  st_na <== prev_new1 + prev_old0 + prev_na + prev_upd;
	st_na = api.Add(api.Add(api.Add(prev_new1, prev_old0), prev_na), prev_upd)
	return
}
