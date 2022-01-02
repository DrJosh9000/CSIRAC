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
	"fmt"
)

const (
	allBits = 0b11111_11111_11111_11111
	lo10    = 0b00000_00000_11111_11111
	hi10    = 0b11111_11111_00000_00000
	signBit = 0b10000_00000_00000_00000

	sourceMask = 0x000003E0
	destMask   = 0x0000001F
)

// Word represents the basic numeric type used by CSIRAC, a 20-bit value.
// The bits in a CSIRAC word, from least to most significant, are called p1
// through p20.
type Word uint32

// P returns a Word with the Pn bit equal to 1, and all other bits equal to 0.
func P(n int) Word { return 1 << (n - 1) }

// String formats the word as an 11-character comma-separated decimal "number
// train". For example, the word 0b00010_10010_00100_00000 would be formatted as
// " 2,18, 4, 0".
func (w Word) String() string {
	return fmt.Sprintf("%2d,%2d,%2d,%2d", w>>15, (w>>10)&0x1f, (w>>5)&0x1f, w&0x1f)
}

// P returns the Pn bit of the word as a 0 or 1.
func (w Word) P(n int) Word {
	n--
	return (w & (1 << n)) >> n
}

// Lo returns the value in the lower 10 bits (p1 - p10).
func (w Word) Lo() Word { return w & lo10 }

// Hi returns the value in the upper 10 bits (p11 - p20).
func (w Word) Hi() Word { return (w & hi10) >> 10 }

// Source returns the upper 5 bits of the lower 10 bits (p6 - p10). When
// interpreting the word as an instruction, this value specifies the source.
func (w Word) Source() Word { return (w & sourceMask) >> 5 }

// Dest returns the lowest 5 bits (p1 - p5). When interpreting the word as an
// instruction, this value specifies the destination.
func (w Word) Dest() Word { return w & destMask }
