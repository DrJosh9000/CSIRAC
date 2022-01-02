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
	// A sample program from the programming guide that adds B to A 9 times.
	c := &CSIRAC{
		A: 13,
		B: 47,
		M: [1024]Word{
			0: MustParseInstruction(" 0  8 K  C"),  // C = 8
			1: MustParseInstruction(" 0  0 B  PA"), // A += B
			2: MustParseInstruction(" 0  0 PE SC"), // C--
			3: MustParseInstruction(" 0  0 SC CS"), // if C < 0 { goto 5 }
			4: MustParseInstruction(" 0  1 K  S"),  // goto 1
			5: MustParseInstruction("31 31 K  T"),  // stop
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
