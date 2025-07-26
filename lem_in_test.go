package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

const ex00 = `4
##start
0 0 3
2 2 5
3 4 0
##end
1 8 3
0-2
2-3
3-1
L1-2
L1-3 L2-2
L1-1 L2-3 L3-2
L2-1 L3-3 L4-2
L3-1 L4-3
L4-1
`

const ex01 = `10
##start
start 1 6
0 4 8
o 6 8
n 6 6
e 8 4
t 1 9
E 5 9
a 8 9
m 8 6
h 4 6
A 5 2
c 8 1
k 11 2
##end
end 11 6
start-t
n-e
a-m
A-c
0-o
E-a
k-end
start-h
o-n
m-end
t-E
start-0
h-A
e-end
c-k
n-m
h-n
L1-t L2-h L3-0
L1-E L2-A L3-o L4-t L5-h L6-0
L1-a L2-c L3-n L4-E L5-A L6-o L7-t L8-h L9-0
L1-m L2-k L3-e L4-a L5-c L6-n L7-E L8-A L9-o L10-t
L1-end L2-end L3-end L4-m L5-k L6-e L7-a L8-c L9-n L10-E
L4-end L5-end L6-end L7-m L8-k L9-e L10-a
L7-end L8-end L9-end L10-m
L10-end
`

const ex02 = `20
##start
0 2 0
1 4 1
2 6 0
##end
3 5 3
0-1
0-3
1-2
3-2
L1-3 L4-1
L2-3 L4-2 L6-1
L3-3 L4-3 L6-2 L8-1
L5-3 L6-3 L8-2 L10-1
L7-3 L8-3 L10-2 L12-1
L9-3 L10-3 L12-2 L14-1
L11-3 L12-3 L14-2 L16-1
L13-3 L14-3 L16-2 L18-1
L15-3 L16-3 L18-2 L20-1
L17-3 L18-3 L20-2
L19-3 L20-3
`

const ex03 = `4
4 5 4
##start
0 1 4
1 3 6
##end
5 6 4
2 3 4
3 3 1
0-1
2-4
1-4
0-2
4-5
3-0
4-3
L1-1
L1-4 L2-1
L1-5 L2-4 L3-1
L2-5 L3-4 L4-1
L3-5 L4-4
L4-5
`

const ex04 = `9
##start
richard 0 6
gilfoyle 6 3
erlich 9 6
dinish 6 9
jimYoung 11 7
##end
peter 14 6
richard-dinish
dinish-jimYoung
richard-gilfoyle
gilfoyle-peter
gilfoyle-erlich
richard-erlich
erlich-jimYoung
jimYoung-peter
L1-gilfoyle L3-dinish
L1-peter L2-gilfoyle L3-jimYoung L5-dinish
L2-peter L3-peter L4-gilfoyle L5-jimYoung L7-dinish
L4-peter L5-peter L6-gilfoyle L7-jimYoung L9-dinish
L6-peter L7-peter L8-gilfoyle L9-jimYoung
L8-peter L9-peter
`

const ex05 = `9
#rooms
##start
start 0 3
##end
end 10 1
C0 1 0
C1 2 0
C2 3 0
C3 4 0
I4 5 0
I5 6 0
A0 1 2
A1 2 1
A2 4 1
B0 1 4
B1 2 4
E2 6 4
D1 6 3
D2 7 3
D3 8 3
H4 4 2
H3 5 2
F2 6 2
F3 7 2
F4 8 2
G0 1 5
G1 2 5
G2 3 5
G3 4 5
G4 6 5
H3-F2
H3-H4
H4-A2
start-G0
G0-G1
G1-G2
G2-G3
G3-G4
G4-D3
start-A0
A0-A1
A0-D1
A1-A2
A1-B1
A2-end
A2-C3
start-B0
B0-B1
B1-E2
start-C0
C0-C1
C1-C2
C2-C3
C3-I4
D1-D2
D1-F2
D2-E2
D2-D3
D2-F3
D3-end
F2-F3
F3-F4
F4-end
I4-I5
I5-end
L1-A0 L5-G0 L6-B0 L7-C0
L1-A1 L2-A0 L5-G1 L6-B1 L7-C1 L9-G0
L1-A2 L2-A1 L3-A0 L5-G2 L6-E2 L7-C2 L9-G1
L1-end L2-A2 L3-A1 L4-A0 L5-G3 L6-D2 L7-C3 L9-G2
L2-end L3-A2 L4-A1 L5-G4 L6-F3 L7-I4 L8-A0 L9-G3
L3-end L4-A2 L5-D3 L6-F4 L7-I5 L8-A1 L9-G4
L4-end L5-end L6-end L7-end L8-A2 L9-D3
L8-end L9-end
`

func TestExamples(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "lem-in")
	if out, err := exec.Command("go", "build", "-o", exe, ".").CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	cases := []struct {
		file string
		want string
	}{
		{"example00.txt", ex00},
		{"example01.txt", ex01},
		{"example02.txt", ex02},
		{"example03.txt", ex03},
		{"example04.txt", ex04},
		{"example05.txt", ex05},
	}
	for _, tc := range cases {
		out, err := exec.Command(exe, filepath.Join("examples", tc.file)).CombinedOutput()
		if err != nil {
			t.Fatalf("%s: %v\n%s", tc.file, err, out)
		}
		got := string(out)
		if got != tc.want {
			t.Errorf("%s output mismatch\nwant:\n%s\ngot:\n%s", tc.file, tc.want, got)
		}
	}
}
