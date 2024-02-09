package poseidon

import (
	"errors"
	"fmt"
	"math/big"
)

// Poseidon hash function based on https://github.com/iden3/go-iden3-crypto/tree/master/poseidon
// but with the ability to choose the elliptic curve field.

type Ptr[T any] interface {
	*T
}

type Element[E any] interface {
	SetUint64(uint64) E
	SetBigInt(*big.Int) E
	BigInt(*big.Int) *big.Int
	SetOne() E
	SetZero() E
	Inverse(E) E
	Set(E) E
	Square(E) E
	Mul(E, E) E
	Add(E, E) E
	Sub(E, E) E
}

// NROUNDSF constant from Poseidon paper
const NROUNDSF = 8

// NROUNDSP constant from Poseidon paper
var NROUNDSP = []int{56, 57, 56, 60, 60, 63, 64, 63, 60, 66, 60, 65, 70, 60, 64, 68}

const spongeChunkSize = 31
const spongeInputs = 16

func zero[E Element[E]](factory func() E) E {
	return factory().SetUint64(0)
}

var big5 = big.NewInt(5)

// exp5 performs x^5 mod p
// https://eprint.iacr.org/2019/458.pdf page 8
func exp5[E Element[E]](factory func() E, a E) {
	b := factory().Set(a)
	Exp[E](a, b, big5)
}

// exp5state perform exp5 for whole state
func exp5state[E Element[E]](factory func() E, state []E) {
	for i := 0; i < len(state); i++ {
		exp5(factory, state[i])
	}
}

// ark computes Add-Round Key, from the paper https://eprint.iacr.org/2019/458.pdf
func ark[E Element[E]](state, c []E, it int) {
	for i := 0; i < len(state); i++ {
		state[i].Add(state[i], c[it+i])
	}
}

// mix returns [[matrix]] * [vector]
func mix[E Element[E]](factory func() E, state []E, t int, m [][]E) []E {
	mul := zero[E](factory)
	newState := make([]E, t)
	for i := 0; i < t; i++ {
		newState[i] = zero[E](factory)
	}
	for i := 0; i < len(state); i++ {
		newState[i].SetUint64(0)
		for j := 0; j < len(state); j++ {
			mul.Mul(m[j][i], state[j])
			newState[i].Add(newState[i], mul)
		}
	}
	return newState
}

func HashMulti[E Element[E]](factory func() E, inpBI []*big.Int) (*big.Int, error) {
	groups := (len(inpBI) + 15) / 16
	state := big.NewInt(0)
	var err error
	for i := 0; i < groups-1; i++ {
		state, err = HashEx[E](factory, inpBI[i*16:(i+1)*16], state)
		if err != nil {
			return nil, err
		}
	}
	return HashEx[E](factory, inpBI[(groups-1)*16:], state)
}

// Hash computes the Poseidon hash for the given inputs
func Hash[E Element[E]](factory func() E, inpBI []*big.Int) (*big.Int, error) {
	return HashEx[E](factory, inpBI, big.NewInt(0))
}

func HashEx[E Element[E]](factory func() E, inpBI []*big.Int, initialState *big.Int) (*big.Int, error) {
	t := len(inpBI) + 1
	if len(inpBI) == 0 || len(inpBI) > len(NROUNDSP) {
		return nil, fmt.Errorf("invalid inputs length %d, max %d", len(inpBI), len(NROUNDSP))
	}
	if !CheckBigIntArrayInField(factory, inpBI) {
		return nil, errors.New("inputs values not inside Finite Field")
	}
	if !CheckBigIntInField(factory, initialState) {
		return nil, errors.New("inputs values not inside Finite Field")
	}
	inp := BigIntArrayToElementArray[E](factory, inpBI)

	nRoundsF := NROUNDSF
	nRoundsP := NROUNDSP[t-2]
	C := c[E](factory, t-2)
	S := s[E](factory, t-2)
	M := m[E](factory, t-2)
	P := p[E](factory, t-2)

	state := make([]E, t)
	state[0] = factory().SetBigInt(initialState)
	copy(state[1:], inp)

	ark(state, C, 0)

	for i := 0; i < nRoundsF/2-1; i++ {
		exp5state(factory, state)
		ark(state, C, (i+1)*t)
		state = mix(factory, state, t, M)
	}
	exp5state(factory, state)
	ark(state, C, (nRoundsF/2)*t)
	state = mix(factory, state, t, P)

	mul := zero[E](factory)
	for i := 0; i < nRoundsP; i++ {
		exp5(factory, state[0])
		state[0].Add(state[0], C[(nRoundsF/2+1)*t+i])

		mul.SetZero()
		newState0 := zero[E](factory)
		for j := 0; j < len(state); j++ {
			mul.Mul(S[(t*2-1)*i+j], state[j])
			newState0.Add(newState0, mul)
		}

		for k := 1; k < t; k++ {
			mul.SetZero()
			state[k] = state[k].Add(state[k], mul.Mul(state[0], S[(t*2-1)*i+t+k-1]))
		}
		state[0] = newState0
	}

	for i := 0; i < nRoundsF/2-1; i++ {
		exp5state(factory, state)
		ark(state, C, (nRoundsF/2+1)*t+nRoundsP+i*t)
		state = mix(factory, state, t, M)
	}
	exp5state(factory, state)
	state = mix(factory, state, t, M)

	rE := state[0]
	r := big.NewInt(0)
	rE.BigInt(r)
	return r, nil
}

// HashBytes returns a sponge hash of a msg byte slice split into blocks of 31 bytes
func HashBytes[E Element[E]](factory func() E, msg []byte) (*big.Int, error) {
	return HashBytesX[E](factory, msg, spongeInputs)
}

// HashBytesX returns a sponge hash of a msg byte slice split into blocks of 31 bytes
func HashBytesX[E Element[E]](factory func() E, msg []byte, frameSize int) (*big.Int, error) {
	if frameSize < 2 || frameSize > 16 {
		return nil, errors.New("incorrect frame size")
	}

	// not used inputs default to zero
	inputs := make([]*big.Int, frameSize)
	for j := 0; j < frameSize; j++ {
		inputs[j] = new(big.Int)
	}
	dirty := false
	var hash *big.Int
	var err error

	k := 0
	for i := 0; i < len(msg)/spongeChunkSize; i++ {
		dirty = true
		inputs[k].SetBytes(msg[spongeChunkSize*i : spongeChunkSize*(i+1)])
		if k == frameSize-1 {
			hash, err = Hash[E](factory, inputs)
			dirty = false
			if err != nil {
				return nil, err
			}
			inputs = make([]*big.Int, frameSize)
			inputs[0] = hash
			for j := 1; j < frameSize; j++ {
				inputs[j] = new(big.Int)
			}
			k = 1
		} else {
			k++
		}
	}

	if len(msg)%spongeChunkSize != 0 {
		// the last chunk of the message is less than 31 bytes
		// zero padding it, so that 0xdeadbeaf becomes
		// 0xdeadbeaf000000000000000000000000000000000000000000000000000000
		var buf [spongeChunkSize]byte
		copy(buf[:], msg[(len(msg)/spongeChunkSize)*spongeChunkSize:])
		inputs[k] = new(big.Int).SetBytes(buf[:])
		dirty = true
	}

	if dirty {
		// we haven't hashed something in the main sponge loop and need to do hash here
		hash, err = Hash[E](factory, inputs)
		if err != nil {
			return nil, err
		}
	}

	return hash, nil
}

// SpongeHash returns a sponge hash of inputs (using Poseidon with frame size of 16 inputs)
func SpongeHash[E Element[E]](factory func() E, inputs []*big.Int) (*big.Int, error) {
	return SpongeHashX[E](factory, inputs, spongeInputs)
}

// SpongeHashX returns a sponge hash of inputs using Poseidon with configurable frame size
func SpongeHashX[E Element[E]](factory func() E, inputs []*big.Int, frameSize int) (*big.Int, error) {
	if frameSize < 2 || frameSize > 16 {
		return nil, errors.New("incorrect frame size")
	}

	// not used frame default to zero
	frame := make([]*big.Int, frameSize)
	for j := 0; j < frameSize; j++ {
		frame[j] = new(big.Int)
	}
	dirty := false
	var hash *big.Int
	var err error

	k := 0
	for i := 0; i < len(inputs); i++ {
		dirty = true
		frame[k] = inputs[i]
		if k == frameSize-1 {
			hash, err = Hash[E](factory, frame)
			dirty = false
			if err != nil {
				return nil, err
			}
			frame = make([]*big.Int, frameSize)
			frame[0] = hash
			for j := 1; j < frameSize; j++ {
				frame[j] = new(big.Int)
			}
			k = 1
		} else {
			k++
		}
	}

	if dirty {
		// we haven't hashed something in the main sponge loop and need to do hash here
		hash, err = Hash[E](factory, frame)
		if err != nil {
			return nil, err
		}
	}

	return hash, nil
}