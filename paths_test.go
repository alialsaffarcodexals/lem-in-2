package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)


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


func TestParseInputErrors(t *testing.T) {
	cases := []struct {
		name string
		data string
		msg  string
	}{
		{
			name: "SelfLoop",
			data: "4\n##start\n0 0 3\n2 2 5\n3 4 0\n##end\n1 8 3\n0-2\n2-3\n1-1\n",
			msg:  "ERROR: invalid data format",
		},
		{
			name: "DupCoords",
			data: "4\n##start\n0 8 3\n2 2 5\n3 4 0\n##end\n1 8 3\n0-2\n2-3\n3-1\n",
			msg:  "ERROR: invalid data format",
		},
		{
			name: "TooManyAnts",
			data: fmt.Sprintf("%d\n##start\nA 0 0\n##end\nB 1 0\nA-B\n", maxAnts+1),
			msg:  "ERROR: ant limit exceeded",
		},
	}
	for _, c := range cases {
		dir := t.TempDir()
		path := filepath.Join(dir, "in.txt")
		if err := os.WriteFile(path, []byte(c.data), 0644); err != nil {
			t.Fatalf("write temp file: %v", err)
		}
		_, _, err := parseInput(path)
		if err == nil || err.Error() != c.msg {
			t.Errorf("%s: expected %q, got %v", c.name, c.msg, err)
		}
	}
}

