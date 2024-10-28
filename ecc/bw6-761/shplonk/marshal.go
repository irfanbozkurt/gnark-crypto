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

package shplonk

import (
	"io"

	"github.com/consensys/gnark-crypto/ecc/bw6-761"
)

func (proof *OpeningProof) ReadFrom(r io.Reader) (int64, error) {

	dec := bw6761.NewDecoder(r)

	toDecode := []interface{}{
		&proof.W,
		&proof.WPrime,
		&proof.ClaimedValues,
	}

	for _, v := range toDecode {
		if err := dec.Decode(v); err != nil {
			return dec.BytesRead(), err
		}
	}

	return dec.BytesRead(), nil
}

// WriteTo writes binary encoding of a OpeningProof
func (proof *OpeningProof) WriteTo(w io.Writer) (int64, error) {

	enc := bw6761.NewEncoder(w)

	toEncode := []interface{}{
		&proof.W,
		&proof.WPrime,
		proof.ClaimedValues,
	}

	for _, v := range toEncode {
		if err := enc.Encode(v); err != nil {
			return enc.BytesWritten(), err
		}
	}

	return enc.BytesWritten(), nil
}
