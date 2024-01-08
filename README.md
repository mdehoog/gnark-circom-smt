# gnark-circom-smt

Sparse Merkle Tree (SMT) implementation for [gnark](https://github.com/Consensys/gnark),
based on [circomlib](https://github.com/iden3/circomlib/tree/master/circuits/smt).

### Example

```go
package main

import (
	"math/big"
	"os"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"go.vocdoni.io/dvote/db"
	"go.vocdoni.io/dvote/db/pebbledb"
	"go.vocdoni.io/dvote/tree/arbo"

	"github.com/mdehoog/gnark-circom-smt/circuits/smt"
)

const levels = 20

type SMTCircuit struct {
	Fnc0     frontend.Variable         `gnark:",public"`
	Fnc1     frontend.Variable         `gnark:",public"`
	OldKey   frontend.Variable         `gnark:",public"`
	NewKey   frontend.Variable         `gnark:",public"`
	IsOld0   frontend.Variable         `gnark:",public"`
	OldValue frontend.Variable         `gnark:",public"`
	NewValue frontend.Variable         `gnark:",public"`
	OldRoot  frontend.Variable         `gnark:",public"`
	NewRoot  frontend.Variable         `gnark:",public"`
	Siblings [levels]frontend.Variable `gnark:",public"`
}

func (circuit *SMTCircuit) Define(api frontend.API) error {
	newRoot := smt.Processor(
		api,
		circuit.OldRoot,
		circuit.Siblings[:],
		circuit.OldKey,
		circuit.OldValue,
		circuit.IsOld0,
		circuit.NewKey,
		circuit.NewValue,
		circuit.Fnc0,
		circuit.Fnc1,
	)
	api.AssertIsEqual(newRoot, circuit.NewRoot)
	return nil
}

func main() {
	var circuit SMTCircuit
	ccs, _ := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)

	pk, vk, _ := groth16.Setup(ccs)

	temp, _ := os.MkdirTemp("", "*")
	defer os.RemoveAll(temp)
	database, _ := pebbledb.New(db.Options{Path: temp})
	tree, _ := arbo.NewTree(arbo.Config{
		Database:     database,
		MaxLevels:    32 * 8,
		HashFunction: arbo.HashFunctionPoseidon,
	})

	a := smt.NewWrapperArbo(tree, levels)
	input, _ := a.Set(big.NewInt(123), big.NewInt(456))

	var siblings [levels]frontend.Variable
	for i := 0; i < len(siblings); i++ {
		siblings[i] = input.Siblings[i]
	}
	assignment := SMTCircuit{
		Fnc0:     input.Fnc0,
		Fnc1:     input.Fnc1,
		IsOld0:   input.IsOld0,
		OldKey:   input.OldKey,
		OldValue: input.OldValue,
		NewKey:   input.NewKey,
		NewValue: input.NewValue,
		OldRoot:  input.OldRoot,
		NewRoot:  input.NewRoot,
		Siblings: siblings,
	}
	witness, _ := frontend.NewWitness(&assignment, ecc.BN254.ScalarField())
	publicWitness, _ := witness.Public()

	proof, _ := groth16.Prove(ccs, pk, witness)
	_ = groth16.Verify(proof, vk, publicWitness)
}
```
