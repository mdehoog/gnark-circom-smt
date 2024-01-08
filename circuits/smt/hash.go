package smt

import (
	"github.com/consensys/gnark/frontend"

	"github.com/liyue201/gnark-circomlib/circuits"
)

func Hash1(api frontend.API, key, value frontend.Variable) frontend.Variable {
	inputs := []frontend.Variable{key, value, 1}
	return circuits.Poseidon(api, inputs)
}

func Hash2(api frontend.API, l, r frontend.Variable) frontend.Variable {
	inputs := []frontend.Variable{l, r}
	return circuits.Poseidon(api, inputs)
}
