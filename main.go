package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Room struct {
	name string
	x    int
	y    int
}


type Graph struct {
	Ants  int
	Rooms map[string]*Room
	Start *Room
	End   *Room
}

const maxPaths = 100

func parseInput(path string) (*Graph, []string, error) {
	file, err := os.Open(path)

	if err != nil {
		fmt.Println("ERROR: invalid data format")
		return
	}
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")
	for _, l := range lines {
		if l != "" {
			fmt.Println(l)
		}
	}
	if len(lines) == 0 {
		fmt.Println("ERROR: invalid data format")
		return
	}
	numAnts, err := strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil || numAnts <= 0 {
		fmt.Println("ERROR: invalid data format")
		return
	}

	rooms := make(map[string]*Room)
	graph := make(map[string][]string)
	var start, end string

	var prevIsStart, prevIsEnd bool
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			if line == "##start" {
				prevIsStart = true
				prevIsEnd = false
			} else if line == "##end" {
				prevIsEnd = true
				prevIsStart = false
			}
			continue
		}
		if strings.Contains(line, "-") { // link
			parts := strings.Split(line, "-")

			links = append(links, [2]string{parts[0], parts[1]})
		} else {
			return nil, lines, errors.New("ERROR: invalid data format")
		}
	}
	if scanner.Err() != nil {
		return nil, lines, scanner.Err()
	}

	if g.Start == nil || g.End == nil || g.Ants <= 0 {
		return nil, lines, errors.New("ERROR: invalid data format")
	}

	for _, l := range links {
		a, ok1 := g.Rooms[l[0]]
		b, ok2 := g.Rooms[l[1]]
		if !ok1 || !ok2 {
			return nil, lines, errors.New("ERROR: invalid data format")
		}
		a.Links = append(a.Links, b)
		b.Links = append(b.Links, a)
	}
	return g, lines, nil
}

func allPaths(g *Graph, limit int) [][]*Room {
	var res [][]*Room
	path := []*Room{}
	visited := map[*Room]bool{}
	var dfs func(*Room)
	dfs = func(r *Room) {
		if len(res) >= limit {
			return
		}
		if r == g.End {
			p := append(append([]*Room{}, path...), r)
			res = append(res, p)
			return
		}
		x, err1 := strconv.Atoi(fields[1])
		y, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil || strings.HasPrefix(fields[0], "L") || strings.HasPrefix(fields[0], "#") {
			fmt.Println("ERROR: invalid data format")
			return
		}
		room := &Room{name: fields[0], x: x, y: y}
		rooms[room.name] = room
		if prevIsStart {
			start = room.name
		}
		if prevIsEnd {
			end = room.name
		}
		prevIsStart = false
		prevIsEnd = false
	}

	rec(0, nil, nil, map[*Room]bool{})
	return best
}

// gatherPaths repeatedly searches for the shortest path while removing the
// intermediate rooms of each discovered path. This yields a set of disjoint
// paths ordered by their discovery order.

// findPaths gathers disjoint paths and selects the prefix that minimises the
// number of turns required to send all ants across the colony.
func findPaths(g *Graph) [][]*Room {
	all := allPaths(g, maxPaths)
	sort.SliceStable(all, func(i, j int) bool {
		li, lj := len(all[i]), len(all[j])
		if li == lj {
			return i < j
		}
		return li < lj
	})
	return bestDisjointPaths(all, g.Ants)
}


	if start == "" || end == "" {
		fmt.Println("ERROR: invalid data format")
		return
	}

	path, ok := bfs(graph, start, end)
	if !ok {
		fmt.Println("ERROR: invalid data format")
		return
	}

	simulate(path, numAnts)
}

func bfs(graph map[string][]string, start, end string) ([]string, bool) {
	queue := []string{start}
	visited := map[string]bool{start: true}
	parent := make(map[string]string)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node == end {
			break
		}
		for _, n := range graph[node] {
			if !visited[n] {
				visited[n] = true
				parent[n] = node
				queue = append(queue, n)
			}
		}
	}
	if _, ok := visited[end]; !ok {
		return nil, false
	}
	var path []string
	for v := end; v != start; v = parent[v] {
		path = append([]string{v}, path...)
	}
	path = append([]string{start}, path...)
	return path, true
}

func simulate(path []string, ants int) {
	if len(path) < 2 || ants <= 0 {
		return
	}
	L := len(path) - 1
	totalSteps := ants + L - 1
	for step := 1; step <= totalSteps; step++ {
		moves := []string{}
		for a := 1; a <= ants; a++ {
			pos := step - a + 1
			if pos >= 1 && pos < len(path) {
				moves = append(moves, fmt.Sprintf("L%d-%s", a, path[pos]))
			}
		}
		if len(moves) > 0 {
			fmt.Println(strings.Join(moves, " "))
		}
	}
}
