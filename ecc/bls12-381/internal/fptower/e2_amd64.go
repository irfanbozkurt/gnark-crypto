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

package fptower

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-381/fp"
)

// q + r'.r = 1, i.e., qInvNeg = - q⁻¹ mod r
// used for Montgomery reduction
const qInvNeg uint64 = 9940570264628428797

// Field modulus q (Fp)
const (
	q0 uint64 = 13402431016077863595
	q1 uint64 = 2210141511517208575
	q2 uint64 = 7435674573564081700
	q3 uint64 = 7239337960414712511
	q4 uint64 = 5412103778470702295
	q5 uint64 = 1873798617647539866
)

var qElement = fp.Element{
	q0,
	q1,
	q2,
	q3,
	q4,
	q5,
}

//go:noescape
func addE2(res, x, y *E2)

//go:noescape
func subE2(res, x, y *E2)

//go:noescape
func doubleE2(res, x *E2)

//go:noescape
func negE2(res, x *E2)

//go:noescape
func mulNonResE2(res, x *E2)

//go:noescape
func squareAdxE2(res, x *E2)

//go:noescape
func mulAdxE2(z, x, y *E2)

// Mul sets z to the E2-product of x,y, returns z
func (z *E2) Mul(x, y *E2) *E2 {
	mulAdxE2(z, x, y)
	return z
}

// MulByNonResidue multiplies a E2 by (1,1)
func (z *E2) MulByNonResidue(x *E2) *E2 {
	mulNonResE2(z, x)
	return z
}

// Square sets z to the E2-product of x,x, returns z
func (z *E2) Square(x *E2) *E2 {
	squareAdxE2(z, x)
	return z
}
