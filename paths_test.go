package main

import "testing"

func TestExamplesTurns(t *testing.T) {
	cases := []struct {
		file string
		want int
	}{
		{"examples/example00.txt", 6},
		{"examples/example01.txt", 8},
		{"examples/example02.txt", 11},
		{"examples/example03.txt", 6},
		{"examples/example04.txt", 6},
		{"examples/example05.txt", 8},
	}
	for _, c := range cases {
		g, _, err := parseInput(c.file)
		if err != nil {
			t.Fatalf("parse %s: %v", c.file, err)
		}
		paths := findPaths(g)
		lengths := make([]int, len(paths))
		for i, p := range paths {
			lengths[i] = len(p) - 1
		}
		turns := computeTurns(g.Ants, lengths)
		if turns != c.want {
			t.Errorf("%s: got %d turns, want %d", c.file, turns, c.want)
		}
	}
}
