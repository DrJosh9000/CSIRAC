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
	"bufio"
	"fmt"
	"io"
	"strings"
)

const (
	allBits    = 0b11111_11111_11111_11111
	lo10       = 0b00000_00000_11111_11111
	hi10       = 0b11111_11111_00000_00000
	signBit    = 0b10000_00000_00000_00000
	sourceMask = 0b00000_00000_11111_00000
	destMask   = 0b00000_00000_00000_11111
)

// Word represents the basic numeric type used by CSIRAC, a 20-bit value.
// The bits in a CSIRAC word, from least to most significant, are called p1
// through p20. The underlying type is uint32 since that is overall easier to
// work with than a signed integer type.
type Word uint32

// P returns a Word with the Pn bit equal to 1, and all other bits equal to 0.
func P(n int) Word { return 1 << (n - 1) }

// IntWord returns x as a Word (truncated to 20 bits).
func IntWord(x int) Word { return Word(x) & allBits }

// String formats the word as an 11-character comma-separated decimal "number
// train". For example, the word 0b00010_10010_00100_00000 would be formatted as
// " 2,18, 4, 0".
func (w Word) String() string {
	return fmt.Sprintf("(%2d,%2d,%2d,%2d)", w>>15, (w>>10)&0x1f, (w>>5)&0x1f, w&0x1f)
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

// InstructionString formats the word as an instruction (U. Melbourne symbols).
func (w Word) InstructionString() string {
	return fmt.Sprintf("%2d %2d %2s %2s", w>>15, (w>>10)&0x1f, sourceToMnemonic[w.Source()], destToMnemonic[w.Dest()])
}

// ParseInstruction parses an instruction string.
func ParseInstruction(k string) (Word, error) {
	var n0, n1 int
	var s, d string
	if _, err := fmt.Sscan(k, &n0, &n1, &s, &d); err != nil {
		return 0, err
	}
	if n0 < 0 || n0 > 31 {
		return 0, fmt.Errorf("first number %d out of valid range [0,31]", n0)
	}
	if n1 < 0 || n1 > 31 {
		return 0, fmt.Errorf("second number %d out of valid range [0,31]", n1)
	}
	sv, ok := mnemonicToSource[s]
	if !ok {
		return 0, fmt.Errorf("invalid source %q", s)
	}
	dv, ok := mnemonicToDest[d]
	if !ok {
		return 0, fmt.Errorf("invalid destination %q", d)
	}
	return Word(n0<<15 + n1<<10 + sv<<5 + dv), nil
}

// MustParseInstruction parses an instruction, or panics.
func MustParseInstruction(k string) Word {
	w, err := ParseInstruction(k)
	if err != nil {
		panic(err)
	}
	return w
}

// ParseProgram parses a (mnemonic-form) program. Programs can include comments
// (starting with semicolon).
func ParseProgram(program io.Reader) ([]Word, error) {
	// TODO: implement offsets
	var m []Word
	lc := 0
	sc := bufio.NewScanner(program)
	for sc.Scan() {
		lc++
		cspl := strings.SplitN(sc.Text(), ";", 2) // trim off comment
		code := strings.TrimSpace(cspl[0])
		if code == "" {
			continue
		}
		ins, err := ParseInstruction(code)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lc, err)
		}
		m = append(m, ins)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

// MustParseProgram parses a (mnemonic form) program or panics.
func MustParseProgram(program string) []Word {
	m, err := ParseProgram(strings.NewReader(program))
	if err != nil {
		panic(err)
	}
	return m
}

var (
	sourceToMnemonic = [32]string{
		0: "M", 1: "I", 2: "NA", 3: "NB", 4: "A",
		5: "SA", 6: "HA", 7: "TA", 8: "LA", 9: "CA",
		10: "ZA", 11: "B", 12: "R", 13: "RB", 14: "C",
		15: "SC", 16: "RC", 17: "D", 18: "SD", 19: "RD",
		20: "Z", 21: "HL", 22: "HU", 23: "S", 24: "PE",
		25: "PL", 26: "K", 27: "MA", 28: "MB", 29: "MC",
		30: "MD", 31: "PS",
	}
	mnemonicToSource = map[string]int{
		"M": 0, "I": 1, "NA": 2, "NB": 3, "A": 4,
		"SA": 5, "HA": 6, "TA": 7, "LA": 8, "CA": 9,
		"ZA": 10, "B": 11, "R": 12, "RB": 13, "C": 14,
		"SC": 15, "RC": 16, "D": 17, "SD": 18, "RD": 19,
		"Z": 20, "HL": 21, "HU": 22, "S": 23, "PE": 24,
		"PL": 25, "K": 26, "MA": 27, "MB": 28, "MC": 29,
		"MD": 30, "PS": 31,
	}

	destToMnemonic = [32]string{
		0: "M", 1: "Q", 2: "OT", 3: "OP", 4: "A",
		5: "PA", 6: "SA", 7: "CA", 8: "DA", 9: "NA",
		10: "P", 11: "B", 12: "XB", 13: "L", 14: "C",
		15: "PC", 16: "SC", 17: "D", 18: "PD", 19: "SD",
		20: "Z", 21: "HL", 22: "HU", 23: "S", 24: "PS",
		25: "CS", 26: "PK", 27: "MA", 28: "MB", 29: "MC",
		30: "MD", 31: "T",
	}
	mnemonicToDest = map[string]int{
		"M": 0, "Q": 1, "OT": 2, "OP": 3, "A": 4,
		"PA": 5, "SA": 6, "CA": 7, "DA": 8, "NA": 9,
		"P": 10, "B": 11, "XB": 12, "L": 13, "C": 14,
		"PC": 15, "SC": 16, "D": 17, "PD": 18, "SD": 19,
		"Z": 20, "HL": 21, "HU": 22, "S": 23, "PS": 24,
		"CS": 25, "PK": 26, "MA": 27, "MB": 28, "MC": 29,
		"MD": 30, "T": 31,
	}
)
