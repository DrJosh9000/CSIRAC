/*
   Copyright 2021 Josh Deprez

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package csirac

import (
	"math/bits"
	"strconv"
)

const (
	allBits = 0x000FFFFF
	lo10    = 0x000003FF
	hi10    = 0x000FFC00
	signBit = 0x00080000

	sourceMask = 0x000003E0
	destMask   = 0x0000001F
)

// Word represents the basic numeric type used by CSIRAC, a 20-bit value. This
type Word uint32

// String formats the word as a signed decimal integer.
func (w Word) String() string {
	if w&signBit != 0 {
		return strconv.Itoa(-int(w - signBit))
	}
	return strconv.Itoa(int(w))
}

// Bit returns the nth bit from the right as 0 or 1.
func (w Word) Bit(n int) Word { return (w & (1 << n)) >> n }

// Sign returns the sign bit (MSB) as 0 or 1.
func (w Word) Sign() Word { return (w & signBit) >> 19 }

// Ones returns the number of 1 bits in the word.
func (w Word) Ones() Word { return Word(bits.OnesCount32(uint32(w))) }

// Lo returns the lower 10 bits.
func (w Word) Lo() Word { return w & lo10 }

// Hi returns the upper 10 bits.
func (w Word) Hi() Word { return (w & hi10) >> 10 }

// Source returns the upper 5 bits of the lower 10 bits.
func (w Word) Source() Word { return (w & sourceMask) >> 5 }

// Dest returns the lower 5 bits.
func (w Word) Dest() Word { return w & destMask }
