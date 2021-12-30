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

// CSIRAC represents the entire CSIRAC machine state.
type CSIRAC struct {
	// Registers (originally implemented using mercury delay-line memory).
	A, B, C Word
	H       Word
	D       [16]Word

	// Sequence register (a.k.a. program counter/instruction pointer).
	S Word

	// Intepreter register (the current instruction).
	K Word

	// Input switches (originally physical switches on the control console).
	I, N1, N2 Word

	// Main store, also originally implemented with mercury delay-line memory.
	// While the total capacity was 1024 words, usually only 768 were in use.
	M [1024]Word

	// Four storage disks. Only one was implemented initially.
	// When CSIRAC moved to Melbourne a second disk was implemented on the
	// underside of the first. That said, the instruction set supports four
	// disks.
	MA, MB, MC, MD [1024]Word
}
