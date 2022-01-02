package csirac

import "testing"

func TestWordString(t *testing.T) {
	tests := []struct {
		x    Word
		want string
	}{
		{
			x:    0,
			want: " 0, 0, 0, 0",
		},
		{
			x:    1,
			want: " 0, 0, 0, 1",
		},
		{
			x:    0b01111_11111_11111_11111,
			want: "15,31,31,31",
		},
		{
			x:    0b11111_11111_11111_11111,
			want: "31,31,31,31",
		},
		{
			x:    0b10000_00000_00000_00000,
			want: "16, 0, 0, 0",
		},
	}

	for _, test := range tests {
		if got, want := test.x.String(), test.want; got != want {
			t.Errorf("%b.String() = %q, want %q", test.x, got, want)
		}
	}
}
