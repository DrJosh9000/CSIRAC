package csirac

import "testing"

func TestWordString(t *testing.T) {
	tests := []struct {
		x    Word
		want string
	}{
		{
			x:    0,
			want: "0",
		},
		{
			x:    1,
			want: "1",
		},
		{
			x:    0b01111_11111_11111_11111,
			want: "524287",
		},
		{
			x:    0b11111_11111_11111_11111,
			want: "-1",
		},
		{
			x:    0b10000_00000_00000_00000,
			want: "-524288",
		},
	}

	for _, test := range tests {
		if got, want := test.x.String(), test.want; got != want {
			t.Errorf("%b.String() = %q, want %q", test.x, got, want)
		}
	}
}

func TestFracMul(t *testing.T) {
	tests := []struct {
		x, y, want Word
	}{
		{
			x:    0b01000_00000_00000_00000, // 1/2
			y:    0b01000_00000_00000_00000, // 1/2
			want: 0b00100_00000_00000_00000, // 1/4
		},
		{
			x:    0b01000_00000_00000_00000, // 1/2
			y:    0b11000_00000_00000_00000, // -1/2
			want: 0b10100_00000_00000_00000, // -1/4
		},
		{
			x:    0b11000_00000_00000_00000, // -1/2
			y:    0b11000_00000_00000_00000, // -1/2
			want: 0b00100_00000_00000_00000, // 1/4
		},
		{
			x:    0b01100_00000_00000_00000, // 3/4
			y:    0b01000_00000_00000_00000, // 1/2
			want: 0b00110_00000_00000_00000, // 3/8
		},
		{
			x:    0b01100_00000_00000_00000, // 3/4
			y:    0b01100_00000_00000_00000, // 3/4
			want: 0b01001_00000_00000_00000, // 9/16
		},
		{
			x:    0b01111_11111_11111_11111, // approximately 1
			y:    0b01111_11111_11111_11111, // approximately 1
			want: 0b01111_11111_11111_11110, // approximately approximately 1
		},
	}

	for _, test := range tests {
		if got, want := FracMul(test.x, test.y), test.want; got != want {
			t.Errorf("FracMul(%b,%b) = %b, want %b", test.x, test.y, got, want)
		}
	}
}
