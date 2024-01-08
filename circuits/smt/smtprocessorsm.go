package smt

import "github.com/consensys/gnark/frontend"

func SMTProcessorSM(api frontend.API, xor, is0, levIns, fnc0, prev_top, prev_old0, prev_bot, prev_new1, prev_na, prev_upd frontend.Variable) (st_top, st_old0, st_bot, st_new1, st_na, st_upd frontend.Variable) {
	aux1 := api.Mul(prev_top, levIns)
	aux2 := api.Mul(aux1, fnc0)
	st_top = api.Sub(prev_top, aux1)
	st_old0 = api.Mul(aux2, is0)
	st_new1 = api.Mul(api.Add(api.Sub(aux2, st_old0), prev_bot), xor)
	st_bot = api.Mul(api.Sub(1, xor), api.Add(api.Sub(aux2, st_old0), prev_bot))
	st_upd = api.Sub(aux1, aux2)
	st_na = api.Add(api.Add(api.Add(prev_new1, prev_old0), prev_na), prev_upd)
	return
}
