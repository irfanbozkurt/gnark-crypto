// Copyright 2020 Consensys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package pedersen

import (
	"crypto/rand"
	"errors"
	"github.com/consensys/gnark-crypto/ecc"
	curve "github.com/consensys/gnark-crypto/ecc/bls24-317"
	"github.com/consensys/gnark-crypto/ecc/bls24-317/fr"
	"io"
	"math/big"
)

// ProvingKey for committing and proofs of knowledge
type ProvingKey struct {
	Basis         []curve.G1Affine
	BasisExpSigma []curve.G1Affine // basisExpSigma[i] = Basis[i]^{σ}
}

type VerifyingKey struct {
	G         curve.G2Affine
	GSigmaNeg curve.G2Affine // GSigmaNeg = G^{-σ}
}

func randomFrSizedBytes() ([]byte, error) {
	res := make([]byte, fr.Bytes)
	_, err := rand.Read(res)
	return res, err
}

type setupConfig struct {
	g2Gen *curve.G2Affine
}

// SetupOption allows to customize Pedersen vector commitment setup.
type SetupOption func(cfg *setupConfig)

// WithG2Point allows to set the G2 generator for the Pedersen vector commitment
// setup. If this is not set, we sample a random G2 point.
func WithG2Point(g2 curve.G2Affine) SetupOption {
	return func(cfg *setupConfig) {
		cfg.g2Gen = &g2
	}
}

// Setup generates the proving keys for Pedersen commitments over the given
// bases allowing for batch proving. The common verifying key can be used to
// verify the batched proof of knowledge.
//
// By default the G2 generator is sampled randomly. This can be overridden by
// providing a custom G2 generator using [WithG2Point] option.
//
// The input bases do not have to be of the same length for individual
// committing and proving. The elements in bases[i] should be linearly
// independent of each other. Otherwise the prover may be able to construct
// multiple valid openings for a commitment.
//
// NB! This is a trusted setup process. The randomness during the setup must be discarded.
// Failing to do so allows to create proofs without knowing the committed values.
func Setup(bases [][]curve.G1Affine, options ...SetupOption) (pk []ProvingKey, vk VerifyingKey, err error) {
	var cfg setupConfig
	for _, o := range options {
		o(&cfg)
	}
	if cfg.g2Gen == nil {
		if vk.G, err = curve.RandomOnG2(); err != nil {
			return
		}
	} else {
		vk.G = *cfg.g2Gen
	}

	var modMinusOne big.Int
	modMinusOne.Sub(fr.Modulus(), big.NewInt(1))
	var sigma *big.Int
	if sigma, err = rand.Int(rand.Reader, &modMinusOne); err != nil {
		return
	}
	sigma.Add(sigma, big.NewInt(1))

	sigmaNeg := new(big.Int).Neg(sigma)
	vk.GSigmaNeg.ScalarMultiplication(&vk.G, sigmaNeg)

	pk = make([]ProvingKey, len(bases))
	for i := range bases {
		pk[i].BasisExpSigma = make([]curve.G1Affine, len(bases[i]))
		for j := range bases[i] {
			pk[i].BasisExpSigma[j].ScalarMultiplication(&bases[i][j], sigma)
		}
		pk[i].Basis = bases[i]
	}
	return
}

// ProveKnowledge generates a proof of knowledge of a commitment to the given
// values over proving key's basis.
func (pk *ProvingKey) ProveKnowledge(values []fr.Element) (pok curve.G1Affine, err error) {
	if len(values) != len(pk.Basis) {
		err = errors.New("must have as many values as basis elements")
		return
	}

	// TODO @gbotrel this will spawn more than one task, see
	// https://github.com/ConsenSys/gnark-crypto/issues/269
	config := ecc.MultiExpConfig{
		NbTasks: 1, // TODO Experiment
	}

	_, err = pok.MultiExp(pk.BasisExpSigma, values, config)
	return
}

// Commit computes a commitment to the values over proving key's basis
func (pk *ProvingKey) Commit(values []fr.Element) (commitment curve.G1Affine, err error) {

	if len(values) != len(pk.Basis) {
		err = errors.New("must have as many values as basis elements")
		return
	}

	// TODO @gbotrel this will spawn more than one task, see
	// https://github.com/ConsenSys/gnark-crypto/issues/269
	config := ecc.MultiExpConfig{
		NbTasks: 1,
	}
	_, err = commitment.MultiExp(pk.Basis, values, config)

	return
}

// BatchProve computes a single proof of knowledge for multiple commitments. The
// single PoK can be verified with a single call to [VerifyingKey.Verify] with
// folded commitments. The commitments can be folded into one using [curve.G1Affine.Fold].
//
// The argument combinationCoeff is used as a linear combination coefficient to
// fold separate proofs into one. It must be the same for batch proving and when
// folding commitments. This means that in an interactive setting, it must be
// randomly generated by the verifier and sent to the prover. Otherwise, it must
// be generated via Fiat-Shamir.
func BatchProve(pk []ProvingKey, values [][]fr.Element, combinationCoeff fr.Element) (pok curve.G1Affine, err error) {
	if len(pk) != len(values) {
		err = errors.New("must have as many value vectors as bases")
		return
	}

	if len(pk) == 1 { // no need to fold
		pok, err = pk[0].ProveKnowledge(values[0])
		return
	} else if len(pk) == 0 { // nothing to do at all
		return
	}

	offset := 0
	for i := range pk {
		if len(values[i]) != len(pk[i].Basis) {
			err = errors.New("must have as many values as basis elements")
			return
		}
		offset += len(values[i])
	}

	// prepare one amalgamated MSM
	scaledValues := make([]fr.Element, offset)
	basis := make([]curve.G1Affine, offset)

	copy(basis, pk[0].BasisExpSigma) // #nosec G602 false positive
	copy(scaledValues, values[0])    // #nosec G602 false positive

	offset = len(values[0]) // #nosec G602 false positive
	rI := combinationCoeff
	for i := 1; i < len(pk); i++ {
		copy(basis[offset:], pk[i].BasisExpSigma)
		for j := range pk[i].Basis {
			scaledValues[offset].Mul(&values[i][j], &rI)
			offset++
		}
		if i+1 < len(pk) {
			rI.Mul(&rI, &combinationCoeff)
		}
	}

	// TODO @gbotrel this will spawn more than one task, see
	// https://github.com/ConsenSys/gnark-crypto/issues/269
	config := ecc.MultiExpConfig{
		NbTasks: 1,
	}

	_, err = pok.MultiExp(basis, scaledValues, config)
	return
}

// Verify checks if the proof of knowledge is valid for a given commitment.
func (vk *VerifyingKey) Verify(commitment curve.G1Affine, knowledgeProof curve.G1Affine) error {

	if !commitment.IsInSubGroup() || !knowledgeProof.IsInSubGroup() {
		return errors.New("subgroup check failed")
	}

	if isOne, err := curve.PairingCheck([]curve.G1Affine{commitment, knowledgeProof}, []curve.G2Affine{vk.GSigmaNeg, vk.G}); err != nil {
		return err
	} else if !isOne {
		return errors.New("proof rejected")
	}
	return nil
}

// BatchVerifyMultiVk verifies multiple separate proofs of knowledge using n+1
// pairings instead of 2n pairings.
//
// The verifying keys may be from different setup ceremonies, but the G2 point
// must be the same. This can be enforced using [WithG2Point] option during
// setup.
//
// The argument combinationCoeff is used as a linear combination coefficient to
// fold separate proofs into one. This means that in an interactive setting, it
// must be randomly generated by the verifier and sent to the prover. Otherwise,
// it must be generated via Fiat-Shamir.
//
// The prover can fold the proofs using [curve.G1Affine.Fold] itself using the
// random challenge, providing the verifier only the folded proof. In this case
// the argument pok should contain only the single folded proof.
func BatchVerifyMultiVk(vk []VerifyingKey, commitments []curve.G1Affine, pok []curve.G1Affine, combinationCoeff fr.Element) error {
	if len(commitments) != len(vk) {
		return errors.New("commitments length mismatch")
	}
	// we use folded POK if provided
	if len(vk) != len(pok) && len(pok) != 1 {
		return errors.New("pok length mismatch")
	}
	for i := range commitments {
		if !commitments[i].IsInSubGroup() {
			return errors.New("commitment subgroup check failed")
		}
		if i != 0 && vk[i].G != vk[0].G {
			return errors.New("parameter mismatch: G2 element")
		}
	}
	for i := range pok {
		if !pok[i].IsInSubGroup() {
			return errors.New("pok subgroup check failed")
		}
	}

	pairingG1 := make([]curve.G1Affine, len(vk)+1)
	pairingG2 := make([]curve.G2Affine, len(vk)+1)
	r := combinationCoeff
	pairingG1[0] = commitments[0]
	var rI big.Int
	for i := range vk {
		pairingG2[i] = vk[i].GSigmaNeg
		if i != 0 {
			r.BigInt(&rI)
			pairingG1[i].ScalarMultiplication(&commitments[i], &rI)
			if i+1 != len(vk) {
				r.Mul(&r, &combinationCoeff)
			}
		}
	}
	if foldedPok, err := new(curve.G1Affine).Fold(pok, combinationCoeff, ecc.MultiExpConfig{NbTasks: 1}); err != nil {
		return err
	} else {
		pairingG1[len(vk)] = *foldedPok
	}
	pairingG2[len(vk)] = vk[0].G

	if isOne, err := curve.PairingCheck(pairingG1, pairingG2); err != nil {
		return err
	} else if !isOne {
		return errors.New("proof rejected")
	}
	return nil
}

// Marshal

func (pk *ProvingKey) writeTo(enc *curve.Encoder) (int64, error) {
	if err := enc.Encode(pk.Basis); err != nil {
		return enc.BytesWritten(), err
	}

	err := enc.Encode(pk.BasisExpSigma)

	return enc.BytesWritten(), err
}

func (pk *ProvingKey) WriteTo(w io.Writer) (int64, error) {
	return pk.writeTo(curve.NewEncoder(w))
}

func (pk *ProvingKey) WriteRawTo(w io.Writer) (int64, error) {
	return pk.writeTo(curve.NewEncoder(w, curve.RawEncoding()))
}

func (pk *ProvingKey) ReadFrom(r io.Reader) (int64, error) {
	dec := curve.NewDecoder(r)

	if err := dec.Decode(&pk.Basis); err != nil {
		return dec.BytesRead(), err
	}
	if err := dec.Decode(&pk.BasisExpSigma); err != nil {
		return dec.BytesRead(), err
	}

	if len(pk.Basis) != len(pk.BasisExpSigma) {
		return dec.BytesRead(), errors.New("commitment/proof length mismatch")
	}

	return dec.BytesRead(), nil
}

func (vk *VerifyingKey) WriteTo(w io.Writer) (int64, error) {
	return vk.writeTo(curve.NewEncoder(w))
}

func (vk *VerifyingKey) WriteRawTo(w io.Writer) (int64, error) {
	return vk.writeTo(curve.NewEncoder(w, curve.RawEncoding()))
}

func (vk *VerifyingKey) writeTo(enc *curve.Encoder) (int64, error) {
	var err error

	if err = enc.Encode(&vk.G); err != nil {
		return enc.BytesWritten(), err
	}
	err = enc.Encode(&vk.GSigmaNeg)
	return enc.BytesWritten(), err
}

func (vk *VerifyingKey) ReadFrom(r io.Reader) (int64, error) {
	return vk.readFrom(r)
}

func (vk *VerifyingKey) UnsafeReadFrom(r io.Reader) (int64, error) {
	return vk.readFrom(r, curve.NoSubgroupChecks())
}

func (vk *VerifyingKey) readFrom(r io.Reader, decOptions ...func(*curve.Decoder)) (int64, error) {
	dec := curve.NewDecoder(r, decOptions...)
	var err error

	if err = dec.Decode(&vk.G); err != nil {
		return dec.BytesRead(), err
	}
	err = dec.Decode(&vk.GSigmaNeg)
	return dec.BytesRead(), err
}