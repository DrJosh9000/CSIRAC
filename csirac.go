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

import (
	"errors"
	"fmt"
	"time"
)

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
	// The upper half of S is the pointer into M.
	S Word

	// Intepreter register (the next instruction to execute).
	K Word

	// Input register (read from the input tape).
	I Word

	// Console switches (physical switches on the control console).
	NA, NB Word

	// Main store, also originally implemented with mercury delay-line memory.
	// While the total capacity was 1024 words, supposedly only 768 were in use
	// much of the time.
	M [1024]Word

	// Four magnetic storage disks of 1024 words each. Only one was implemented
	// initially. When CSIRAC moved to Melbourne a second disk was implemented
	// on the underside of the first. That said, the instruction set supports
	// four disks.
	MA, MB, MC, MD [1024]Word

	// Outputs
	Printer, TapePunch, Loudspeaker func(Word)
}

func (c *CSIRAC) String() string {
	return fmt.Sprintf("K:%v\tS:%v\nA:%v\tB:%v\tC:%v\tH:%v\n%s\n", c.K, c.S, c.A, c.B, c.C, c.H, c.K.InstructionString())
}

// Run runs the computer until it reaches a stop or an error. It runs one
// instruction per period. If period <= 0, it runs without any artificial delay.
// If the computer encounters a stop, Run will finish and return nil. (Run does
// not return ErrStop).
func (c *CSIRAC) Run(period time.Duration, trace bool) error {
	if period <= 0 {
		for {
			if trace {
				fmt.Println(c)
			}
			if err := c.Step(); err != nil {
				if errors.Is(err, ErrStop) {
					return nil
				}
				return err
			}
		}
	}
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for range ticker.C {
		if trace {
			fmt.Println(c)
		}
		if err := c.Step(); err != nil {
			if errors.Is(err, ErrStop) {
				return nil
			}
			return err
		}
	}
	return nil
}

// Step executes the instruction in K and fetches the next instruction.
func (c *CSIRAC) Step() error {
	inst := c.K
	src := c.ReadSource()
	// Three things could happen depending on the destination:
	// 1) The destination is neither S nor K. Increment S and then fetch the
	//    next instruction as normal here.
	// 2) The destination is S (S, PS, CS). WriteDest updates S and refetches K.
	//    (The programmer is warned to take into account the unconditional
	//    increment of S when sending values to PS or CS).
	// 3) The destination modifies K (PK). (It doesn't modify M[S], just K).
	//    Since it doesn't modify S, the next instruction is always the one
	//    fetched here.
	c.S += P(11)
	c.K = c.M[c.S.Hi()]
	return c.WriteDest(inst, src)
}

// ReadSource reads the source field from K, and uses that to read a word from a
// variety of sources.
func (c *CSIRAC) ReadSource() Word {
	switch c.K.Source() {
	case 0: // n M - Read from main store
		// "Transmit the contents of cell number n of the main store (20 digits)"
		return c.M[c.K.Hi()]
	case 1: // I - Read input register
		// "Transmit the content of the input register (20 digits) and shift the
		// input tape"
		return c.I
	case 2: // NA - Read switch register 1
		// "Transmit the contents of hand set register No. 1 (20 digits)"
		return c.NA
	case 3: // NB - Read switch register 2
		// "Transmit the contents of hand set register No. 2 (20 digits)"
		return c.NB
	case 4: // A - Read the A register
		// "Transmit the contents of the A-register (20 digits)"
		return c.A
	case 5: // SA - Read the sign bit of the A register
		// While the "CSIRAC Hardware" article says the sign is returned as p1,
		// the programming manual implies this source does *not* translate from
		// bit p20 to bit p1.
		//
		// Appendix 3:
		// "Transmit the sign digit of A, i.e. the most significant digit of A."
		return c.A & signBit
	case 6: // HA - "Half A" - Read the A register shifted right
		// More specifically, an arithmetic (sign-preserving) shift, not a
		// logical shift.
		// "Transmit the contents of A divided by 2 (Half A)."
		return (c.A >> 1) | (c.A & signBit)
	case 7: // TA - Read the A register shifted left
		// "Transmit the contents of A multiplied by 2 (Twice A)."
		return (c.A << 1) & allBits
	case 8: // LA - Read the Least significant bit of the A register
		// "Transmit the least significant digit of A."
		return c.A & 1
	case 9: // CA - Read A then clear it
		// "Transmit the contents of A, and clear A to zero."
		a := c.A
		c.A = 0
		return a
	case 10: // ZA - Nonzero-test A, report result as a single bit
		// "If A = 0, transmit zero, otherwise transmit a PL digit."
		if c.A == 0 {
			return 0
		}
		return 1
	case 11: // B - Read the B register
		// "Transmit the contents of the B-register (20 digits)"
		return c.B
	case 12: // R - Read the sign bit of the B register
		// Both "CSIRAC Hardware" and the programming manual agree that this
		// transmits a p1 bit equal to p20 of B.
		// "If the most significant digit of B is 1, transmit PL, otherwise transmit
		// zero."
		return c.B.P(20)
	case 13: // RB - Read the B register shifted right (logical shift)
		// "Transmit the contents of the B-register shifted one place to the right,
		// with zero as the most significant bit."
		return c.B >> 1
	case 14: // C - Read the C register
		// "Transmit the contents of the C-register."
		return c.C
	case 15: // SC - Read the sign bit of the C register
		// While the "CSIRAC Hardware" article says the sign is returned as p1,
		// the programming manual implies this source does *not* translate from
		// bit p20 to bit p1.
		// "Transmit the sign bit of C, i.e. the most significant digit of C."
		return c.C & signBit
	case 16: // RC - Read the C register shifted right
		// "Transmit the contents of C shifted one place to the right, with zero in
		// the sign digit position."
		return c.C >> 1
	case 17: // n D - Read from one of the D registers
		// The programming manual says simultaneous operation on a store cell
		// and a D register if the lower four binary digits of the cell address
		// are the same as the D register address.
		// "Transmit the contents of the nth D-register (20 digits)."
		return c.D[c.K.Hi()&0xF]
	case 18: // n SD - Read the sign bit of one of the D registers
		// The programming manual implies this source does not translate from
		// bit p20 to bit p1.
		// "Transmit the sign bit of the nth D-register."
		return c.D[c.K.Hi()&0xF] & signBit
	case 19: // n RD - Read one of the D registers shifted right
		// "Transmit the contents of the nth D-register shifted one place to the
		// right, with zero in the sign digit position."
		d := c.D[c.K.Hi()&0xF]
		return d >> 1
	case 20: // Z - Read zero.
		// "Transmit zero (20 digits)."
		return 0
	case 21: // HL - Read the H register as a lower half
		// "Transmit the contents in the position group P1-P10 of the H-register"
		return c.H
	case 22: // HU - Read the H register as an upper half
		// "Transmit the contents in the position group P11-P20 of the H-register"
		return c.H << 10
	case 23: // S - Read sequence register
		// "Transmit the contents of the S-register (20 digits)."
		return c.S
	case 24: // PE - Read "upper" 1 (P-Eleven)
		// "Transmit 1 in the P11 position."
		return P(11)
	case 25: // PL - Read 1 (P-Least)
		// "Transmit 1 in the P1 position."
		return 1
	case 26: // n K - Read the upper half of the instruction (a literal)
		// "Transmit from the interpreter-register (K) the number n as the most
		// significant digits of a 20 digit number, with the least significant 10
		// digits zero."
		return c.K & hi10
	case 27: // n MA - Read disk 1
		// "Transmit the contents of cell No. n of the magnetic drum store No. 1."
		return c.MA[c.K.Hi()]
	case 28: // n MB - Read disk 2
		// "Transmit the contents of cell No. n of the magnetic drum store No. 2."
		return c.MB[c.K.Hi()]
	case 29: // n MC - Read disk 3
		// "Transmit the contents of cell No. n of the magnetic drum store No. 3."
		return c.MC[c.K.Hi()]
	case 30: // n MD - Read disk 4
		// "Transmit the contents of cell No. n of the magnetic drum store No. 4."
		return c.MD[c.K.Hi()]
	case 31: // PS - Read a number with 1 in the sign bit (P-Sign)
		// "Transmit 1 in the P20 digit position."
		return signBit
	}
	panic("k.Source returned a number outside [0, 31]")
}

// WriteDest reads the dest field from inst, and uses that to write src to a
// variety of destinations.
func (c *CSIRAC) WriteDest(inst, src Word) error {
	switch inst.Dest() {
	case 0: // n M - Write to main store
		// "Replace the content of cell n of the main store by the digit entering."
		c.M[inst.Hi()] = src
	case 1: // Q - Set binary or decimal input
		// Programming guide appendix 3: "Has no effect"
	case 2: // OT - Write to console printer
		// "Print on the teleprinter the character corresponding to digits 1 to 5
		// of the output register."
		if c.Printer != nil {
			c.Printer(src)
		}
	case 3: // OP - Write to tape punch
		// "Output to the five hole punch the digits in positions 1-5 of the output
		// register."
		if c.TapePunch != nil {
			c.TapePunch(src)
		}
	case 4: // A - Write to A register
		// "Replace the contents of the A-register by the 20 entering digits."
		c.A = src
	case 5: // PA - Add into A register
		// "Add to the contents of A and hold the sum."
		c.A = (c.A + src) & allBits
	case 6: // SA - Subtract into A register
		// "Subtract from the contents of A and hold the difference."
		c.A = (c.A - src) & allBits
	case 7: // CA - AND with A register (C for Conjunction)
		// "Replace the contents of A by the digit by digit logical product of its
		// contents and the entering digits (i.e. conjunction)."
		c.A &= src
	case 8: // DA - OR with A register (D for Disjunction)
		// Programming manual says this is OR, not XOR.
		// "Replace the contents of A by the digit by digit logical sum of its
		// contents and the entering digits (i.e. disjunction)."
		c.A |= src
	case 9: // NA - XOR with A register (N for negation)
		// Programming manual says this is XOR, not NAND.
		// "Compare digit by digit the contents of A with the entering digits,
		// placing 0 or 1 in the digit position as the digits compared are the
		// same or different."
		c.A ^= src
	case 10: // P - Loudspeaker
		// "Transmit the entering bit stream to the loudspeaker."
		c.Loudspeaker(src)
	case 11: // B - Write into B register
		// "Replace the content of the B-register by the entering 20 digits."
		c.B = src
	case 12: // XB - Multiplication.
		// "CSIRAC Hardware" doesn't describe this well at all. The
		// destinations table simply says
		// "B - multiply: B = A + source X register C".
		// It also says that for the multiplier unit, numbers are signed 19-bit
		// fractions. Fortunately the programming manual (CMANUAL.pdf) is
		// clearer and more accurate.
		// After determining the output sign bit, the remaining 19-bit integers
		// (source and C) are multiplied, with the most-significant 19 bits
		// *added* into A, and the least-significant 19 bits replacing
		// bits 20 through 2 of B (and bit 1 cleared).
		// The interpretation of the result is then up to the programmer, rather
		// than assuming the multiplicands are always fractions. There are
		// three standard interpretations depending on the multiplicands (two
		// integers, two fractions, or integer x fraction).
		//
		// Appendix 3:
		// "Substitute the entering number into the B-register.  Then form the
		// product of the contents of B and C in A and B, the top 20
		// digits of the product being added to A and placing the lower 19 bits in
		// B with a zero in the PL position."
		sign := (src & signBit) ^ (c.C & signBit)
		prod := uint64(src&^signBit) * uint64(c.C&^signBit)
		c.A = (c.A + sign + Word(prod>>19)) & allBits
		c.B = Word(prod<<1) & allBits
	case 13: // L - "A and B shifted 1 left IF source bit 20 is set"
		// This is more accurately called "40-bit left rotate".
		// Again, the programming manual is clearer. This destination treats A
		// and B as a combined 40-bit register, where bits leave the left of B
		// to become the right of A, and leave the left of A to become the right
		// of B. Up to 7 shifts can be called depending on the source. From the
		// programming manual:
		// 16  0 K L - one left shift
		// 16  2 K L - two left shifts
		// ...
		// 16 12 K L - seven left shifts
		//
		// Appendix 3:
		// "Left shift the contents of A and B n places, 1 ≤ n ≤ 7. (See Ch 2)"
		if src.P(20) != 1 {
			break
		}
		n := (src.Hi()>>1)&0xf + 1
		a := c.A << n
		b := c.B << n
		c.A = (a + (b >> 20)) & allBits
		c.B = (b + (a >> 20)) & allBits
	case 14: // C - Write into C register
		// "Replace the contents of the C-register by the 20 entering digits."
		c.C = src
	case 15: // PC - Add into C register
		// "Add to the contents of C and hold the sum."
		c.C = (c.C + src) & allBits
	case 16: // SC - Subtract into C register
		// "Subtract from the contents of C and hold the difference."
		c.C = (c.C - src) & allBits
	case 17: // n D - Write into a D register
		// "Replace the contents of the nth D-register by the 20 entering digits"
		c.D[inst.Hi()&0xf] = src
	case 18: // n PD - Add into a D register
		// "Add to the contents of the nth D-register and hold the sum."
		c.D[inst.Hi()&0xf] = (c.D[inst.Hi()&0xf] + src) & allBits
	case 19: // n SD - Subtract into a D register
		// "Subtract from the contents of nth D-register and hold the difference."
		c.D[inst.Hi()&0xf] = (c.D[inst.Hi()&0xf] - src) & allBits
	case 20: // Z - Null
		// "Has no effect."
	case 21: // HL - H as lower half
		// "Replace the 10 digits of the H-register by the P1-P10 bits of the
		// entering number."
		c.H = src.Lo()
	case 22: // HU - H as upper half
		// "Replace the 10 digits of the H-register by the P11-P20 digits of the
		// entering number"
		c.H = src.Hi()
	case 23: // S - Write into sequence register (absolute jump)
		// "Replace the contents of the S-register by the 20 entering digit."
		// Since this is commonly used with the K source, which transmits the
		// top half of src as the most sigificant bits, we can infer that
		// the upper half of S is what points into M, not the lower half.
		c.S = src
		c.K = c.M[c.S.Hi()]
	case 24: // PS - Add into sequence register (relative jump)
		// "Add to the contents of the S-register."
		c.S += src
		c.K = c.M[c.S.Hi()]
	case 25: // CS - Conditionally increase sequence register
		// This is specifically for conditional execution of the following
		// instruction.
		//
		// Per the programming manual:
		// If bits are recieved in either p1-p11 or p15-p20 (one of the ranges
		// are non-zero), an extra unit is added to S. If bits in both ranges
		// are received, a further unit is added.
		//
		// Note that a "unit" is actually a p11, since the *upper* half of S
		// indexes into M (not the lower half).
		//
		// Appendix 3:
		// "Add one P11 digit to the S-register if either the P1-P11 or the P15-P20
		// group of the entering digits is not completely zero, and 2P11 if
		// neither group is entirely zero."
		//
		// The ranges and possible double-increment seem confusing and
		// arbitrary, but whatever.
		if src&0b00000_00001_11111_11111 != 0 { // p1 - p11
			c.S += P(11)
		}
		if src&0b11111_10000_00000_00000 != 0 { // p15 - p20
			c.S += P(11)
		}
		c.K = c.M[c.S.Hi()]
	case 26: // PK - Add into instruction register
		// "CSIRAC Hardware" doesn't fully explain what happens here - further-
		// more, the "upper half" wording seems to be a mistake.
		// The programming manual says a number transmitted to PK is held and
		// the next command is added to it. (I do it the other way around.)
		// It spells out that the source and destination of the next instruction
		// can also be changed with PK, not just the address (but )
		//
		// Appendix 3:
		// " Replace the content of the K-register by the 20 digits entering.
		// Add the digits forming the next command and obey the command represented
		// by this sum."
		c.K = (c.K + src) & allBits
	case 27: // n MA - Disk 1
		// "Replace the 20 bits of cell No. n of the magnetic drum store No.1 by the
		// entering digits."
		c.MA[inst.Hi()] = src
	case 28: // n MB - Disk 2
		// "As for 27 but using auxiliary store No. 2"
		c.MB[inst.Hi()] = src
	case 29: // n MC - Disk 3
		// "As for 27 but using auxiliary store No. 3"
		c.MC[inst.Hi()] = src
	case 30: // n MD - Disk 4
		// "As for 27 but using auxiliary store No. 4"
		c.MD[inst.Hi()] = src
	case 31: // T - Stop if non-zero
		// CSIRAC remains ready to continue with the next instruction.
		// "If one or more digits received, computer; do not proceed to the next
		// command"
		if src != 0 {
			return ErrStop
		}
	default:
		panic("k.Dest returned a number outside [0, 31]")
	}
	return nil
}
