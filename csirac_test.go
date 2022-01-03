/*
   Copyright 2022 Josh Deprez

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

import "testing"

func TestCSIRACCountDownLoop(t *testing.T) {
	// A sample program from the programming guide that adds B to A 9 times,
	// using a "count down" loop.
	c := &CSIRAC{
		A: 13,
		B: 47,
		M: []Word{
			0: MustParseInstruction(" 0  8 K  C"),  // C = 8
			1: MustParseInstruction(" 0  0 B  PA"), // A += B
			2: MustParseInstruction(" 0  0 PE SC"), // C--
			3: MustParseInstruction(" 0  0 SC CS"), // if C < 0 { skip next }
			4: MustParseInstruction(" 0  1 K  S"),  // goto 1
			5: MustParseInstruction("31 31 K  T"),  // stop
			6: 0,
		},
	}
	c.K = c.M[0]

	if err := c.Run(0, false); err != nil {
		t.Errorf("c.Run(0) = %v, want nil", err)
	}
	if got, want := c.A, Word(13+9*47); got != want {
		t.Errorf("after Run: c.A = %d, want %d", got, want)
	}
}

func TestCSIRACCountUpLoop(t *testing.T) {
	// A sample program from the programming guide that adds B to A 9 times,
	// using a "count up" loop.
	c := &CSIRAC{
		A: 13,
		B: 47,
		M: []Word{
			0: MustParseInstruction("31 23 K  C"),  // C = -9
			1: MustParseInstruction(" 0  0 B  PA"), // A += B
			2: MustParseInstruction(" 0  0 PE PC"), // C++
			3: MustParseInstruction(" 0  0 SC CS"), // if C < 0 { skip next }
			4: MustParseInstruction(" 0  0 PE PS"), // skip next
			5: MustParseInstruction("31 27 K  PS"), // goto (line - 4)
			6: MustParseInstruction(" 0  0 PS T"),  // stop
			7: 0,
		},
	}
	c.K = c.M[0]

	if err := c.Run(0, false); err != nil {
		t.Errorf("c.Run(0) = %v, want nil", err)
	}
	if got, want := c.A, Word(13+9*47); got != want {
		t.Errorf("after Run: c.A = %d, want %d", got, want)
	}
}

func TestCSIRACStrobeLoop(t *testing.T) {
	// A sample program from the programming guide that adds B to A 9 times,
	// using a "strobe" loop.
	c := &CSIRAC{
		A: 13,
		B: 47,
		M: []Word{
			0: MustParseInstruction(" 0  0 PE C"),  // C = P11
			1: MustParseInstruction(" 0  0 B  PA"), // A += B
			2: MustParseInstruction(" 0  0 C  PC"), // C += C // alternatively C *= 2 or C <<= 1
			3: MustParseInstruction(" 0  0 SC CS"), // if C < 0 { skip next }
			4: MustParseInstruction("31 28 K  PS"), // goto (line - 3)
			5: MustParseInstruction(" 0  0 PL T"),  // stop
			6: 0,
		},
	}
	c.K = c.M[0]

	if err := c.Run(0, false); err != nil {
		t.Errorf("c.Run(0) = %v, want nil", err)
	}
	if got, want := c.A, Word(13+9*47); got != want {
		t.Errorf("after Run: c.A = %d, want %d", got, want)
	}
}
