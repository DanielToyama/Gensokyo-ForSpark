package handlers

import "testing"

func TestMemberIDStrings(t *testing.T) {
	cases := []struct {
		in   interface{}
		want []string
	}{
		{"123", []string{"123"}},
		{float64(456), []string{"456"}},
		{[]string{"1", "2"}, []string{"1", "2"}},
		{[]interface{}{"3", float64(4)}, []string{"3", "4"}},
		{nil, nil},
	}
	for _, c := range cases {
		got := memberIDStrings(c.in)
		if len(got) != len(c.want) {
			t.Fatalf("memberIDStrings(%v) = %v, want %v", c.in, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("memberIDStrings(%v) = %v, want %v", c.in, got, c.want)
			}
		}
	}
}
