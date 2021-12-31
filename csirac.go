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

// Package csirac implements the historical CSIRAC computer as a virtual
// machine.
package csirac

import "errors"

var (
	// ErrStop is returned by the machine if it has stopped. It's not an error
	// - the machine could stop normally due to a T destination.
	ErrStop = errors.New("machine stopped")
)

// CSIRAC represents the entire CSIRAC machine state.
type CSIRAC struct {
	// Registers (originally implemented using mercury delay-line memory).
	A Word     // Supports +,-,<<,>>,AND,XOR,NAND
	B Word     // Supports >>,*
	C Word     // Supports >>,+,-
	H Word     // 10 bits - stores upper or lower half word
	D [16]Word // Each supports >>,+,-

	// Sequence register (a.k.a. program counter/instruction pointer).
	S Word

	// Intepreter register (the current instruction).
	K Word

	// Input switches (originally physical switches on the control console).
	I, N1, N2 Word

	// Main store, also originally implemented with mercury delay-line memory.
	// While the total capacity was 1024 words, usually only 768 were in use.
	M [1024]Word

	// Four magnetic storage disks of 1024 words each. Only one was implemented
	// initially. When CSIRAC moved to Melbourne a second disk was implemented
	// on the underside of the first. That said, the instruction set supports
	// four disks.
	MA, MB, MC, MD [1024]Word

	// Outputs
	Printer, TapePunch, Loudspeaker func(Word)
}

// Step executes one instruction.
func (c *CSIRAC) Step() {
	inst := c.K
	src := c.ReadSource(inst)
	// Three things could happen depending on the destination:
	// 1) The destination is neither S nor K. Increment S and then fetch the
	//    next instruction as normal here.
	// 2) The destination is S. WriteDest updates S and fetches K again.
	// 3) The destination modifies K (PK). Since it doesn't modify S, the
	//    next instruction is always the one fetched here.
	c.S++
	c.K = c.M[c.S]
	c.WriteDest(inst, src)
}

// ReadSource reads the source field from the instruction k, and uses that to
// read a word from a variety of sources.
func (c *CSIRAC) ReadSource(inst Word) Word {
	switch inst.Source() {
	case 0: // n M - Read from main store
		return c.M[inst.Hi()]
	case 1: // I - Read input switches
		return c.I
	case 2: // NA - Read switch register 1
		return c.N1
	case 3: // NB - Read switch register 2
		return c.N2
	case 4: // A - Read the A register
		return c.A
	case 5: // SA - Read the sign bit of the A register
		return c.A.Sign()
	case 6: // HA - Read the A register shifted right
		return c.A >> 1
	case 7: // TA - Read the A register shifted left
		return (c.A << 1) & allBits
	case 8: // LA - Read the least significant bit of the A register
		return c.A & 1
	case 9: // CA - Read A then clear it
		a := c.A
		c.A = 0
		return a
	case 10: // ZA - Nonzero-test A, report result as a single bit
		if c.A == 0 {
			return 0
		}
		return 1
	case 11: // B - Read the B register
		return c.B
	case 12: // R - Read the sign bit of the B register
		return c.B.Sign()
	case 13: // RB - Read the B register shifted right
		return c.B >> 1
	case 14: // C - Read the C register
		return c.C
	case 15: // SC - Read the sign bit of the C register
		return c.C.Sign()
	case 16: // RC - Read the C register shifted right
		return c.C >> 1
	case 17: // n D - Read from one of the D registers
		return c.D[inst.Hi()]
	case 18: // n SD - Read the sign bit of one of the D registers
		return c.D[inst.Hi()].Sign()
	case 19: // n RD - Read one of the D registers shifted right
		return c.D[inst.Hi()] >> 1
	case 20: // Z - Read zero.
		return 0
	case 21: // HL - Read the H register as a lower half
		return c.H
	case 22: // HU - Read the H register as an upper half
		return c.H << 10
	case 23: // S - Read sequence register as upper half
		// TODO: check
		return c.S << 10
	case 24: // PE - Read "upper" 1
		return 1 << 10
	case 25: // PL - Read 1
		return 1
	case 26: // n K - Read the upper half of the instruction (a literal)
		// TODO: check
		return inst.Hi()
	case 27: // n MA - Read disk 1
		return c.MA[inst.Hi()]
	case 28: // n MB - Read disk 2
		return c.MB[inst.Hi()]
	case 29: // n MC - Read disk 3
		return c.MC[inst.Hi()]
	case 30: // n MD - Read disk 4
		return c.MD[inst.Hi()]
	case 31: // PS - Read a number with 1 in the sign bit
		return 1 << 19
	}
	panic("k.Source returned a number outside [0, 31]")
}

// WriteDest reads the dest field from the instruction k, and uses that to
// write the word v to a variety of sources.
func (c *CSIRAC) WriteDest(inst, src Word) error {
	switch inst.Dest() {
	case 0: // n M - Write to main store
		c.M[inst.Hi()] = src
	case 1: // Q - Set binary or decimal input
		// "in Melb no-op"
	case 2: // OT - Write to console printer
		c.Printer(src)
	case 3: // OP - Write to tape punch
		c.TapePunch(src)
	case 4: // A - Write to A register
		c.A = src
	case 5: // PA - Add into A register
		c.A = (c.A + src) & allBits
	case 6: // SA - Subtract into A register
		c.A = (c.A - src) & allBits
	case 7: // CA - AND with A register
		c.A &= src
	case 8: // DA - XOR with A register
		c.A ^= src
	case 9: // NA - NAND with A register
		c.A = ^(c.A & src) & allBits
	case 10: // P - Loudspeaker
		c.Loudspeaker(src)
	case 11: // B - Write into B register
		c.B = src
	case 12: // XB - B = A + v*C
		// TODO: Fix this
		// Page 10 of "The last of the first" says that for the multiplier unit,
		// numbers are signed 19-bit fractions. Is that the source,
		// the C register, or both?
		c.B = (c.A + src*c.C) & allBits
	case 13: // L - If bit 20 is set, shift A and B left
		if src.Sign() == 1 {
			c.A = (c.A << 1) & allBits
			c.B = (c.B << 1) & allBits
		}
	case 14: // C - Write into C register
		c.C = src
	case 15: // PC - Add into C register
		c.C = (c.C + src) & allBits
	case 16: // SC - Subtract into C register
		c.C = (c.C - src) & allBits
	case 17: // n D - Write into a D register
		c.D[inst.Hi()] = src
	case 18: // n PD - Add into a D register
		c.D[inst.Hi()] = (c.D[inst.Hi()] + src) & allBits
	case 19: // n SD - Subtract into a D register
		c.D[inst.Hi()] = (c.D[inst.Hi()] - src) & allBits
	case 20: // Z - Null
		// no-op I guess.
	case 21: // HL - H as lower half
		c.H = src.Lo()
	case 22: // HU - H as upper half
		c.H = src.Hi()
	case 23: // S - Write into sequence register (absolute jump)
		c.S = src
		c.K = c.M[c.S]
	case 24: // PS - Add into sequence register (relative jump)
		c.S += src
		c.K = c.M[c.S]
	case 25: // CS - Count into sequence register
		// TODO: check this
		c.S += src.Ones()
		c.K = c.M[c.S]
	case 26: // PK - Add into instruction register (add as upper half to next instruction)
		// The "table of numbers" example is illuminating. For a program:
		//     A PK
		// 3 4 M B     // 3*32 + 4 = 100
		// where A holds an index value into a table, PK then modifies the next
		// instruction (3 4 M B, i.e. B = M[100]) to instead read from M[100+A].
		c.K = (c.K + src<<10) & allBits
	case 27: // n MA - Disk 1
		c.MA[inst.Hi()] = src
	case 28: // n MB - Disk 2
		c.MB[inst.Hi()] = src
	case 29: // n MC - Disk 3
		c.MC[inst.Hi()] = src
	case 30: // n MD - Disk 4
		c.MD[inst.Hi()] = src
	case 31: // T - Stop if non-zero
		if src != 0 {
			return ErrStop
		}
	default:
		panic("k.Dest returned a number outside [0, 31]")
	}
	return nil
}
