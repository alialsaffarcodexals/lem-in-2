package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Room struct {
	Name  string
	X, Y  int
	Links []*Room
}

type Graph struct {
	Ants  int
	Rooms map[string]*Room
	Start *Room
	End   *Room
}

func parseInput(path string) (*Graph, []string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	g := &Graph{Rooms: make(map[string]*Room)}
	var lines []string
	scanner := bufio.NewScanner(file)
	var pendingStart bool
	var pendingEnd bool
	var links [][2]string
	parsedAnts := false

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if strings.HasPrefix(line, "#") {
			if line == "##start" {
				if pendingStart || g.Start != nil {
					return nil, lines, errors.New("ERROR: invalid data format")
				}
				pendingStart = true
			} else if line == "##end" {
				if pendingEnd || g.End != nil {
					return nil, lines, errors.New("ERROR: invalid data format")
				}
				pendingEnd = true
			}
			continue
		}

		if !parsedAnts {
			ants, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil || ants <= 0 {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			g.Ants = ants
			parsedAnts = true
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 3 {
			if strings.HasPrefix(fields[0], "L") || strings.HasPrefix(fields[0], "#") {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			if _, ok := g.Rooms[fields[0]]; ok {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			x, err1 := strconv.Atoi(fields[1])
			y, err2 := strconv.Atoi(fields[2])
			if err1 != nil || err2 != nil {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			room := &Room{Name: fields[0], X: x, Y: y}
			g.Rooms[room.Name] = room
			if pendingStart {
				g.Start = room
				pendingStart = false
			}
			if pendingEnd {
				g.End = room
				pendingEnd = false
			}
		} else if strings.Count(line, "-") == 1 && !strings.Contains(line, " ") {
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

// bfs finds the shortest path from start to end while avoiding nodes in banned
// (except for start and end themselves). The neighbors are explored in the
// order they appear in the input so the resulting paths match the examples.
func bfs(start, end *Room, banned map[*Room]bool) []*Room {
	type item struct {
		r    *Room
		prev *item
	}
	q := []*item{{r: start}}
	visited := map[*Room]bool{start: true}
	for len(q) > 0 {
		it := q[0]
		q = q[1:]
		if it.r == end {
			var rev []*Room
			for p := it; p != nil; p = p.prev {
				rev = append(rev, p.r)
			}
			path := make([]*Room, len(rev))
			for i := range rev {
				path[i] = rev[len(rev)-1-i]
			}
			return path
		}
		for _, nb := range it.r.Links {
			if (nb == end || !banned[nb]) && !visited[nb] {
				visited[nb] = true
				q = append(q, &item{r: nb, prev: it})
			}
		}
	}
	return nil
}

// gatherPaths repeatedly searches for the shortest path while removing the
// intermediate rooms of each discovered path. This yields a set of disjoint
// paths ordered by their discovery order.
func gatherPaths(g *Graph) [][]*Room {
	banned := map[*Room]bool{}
	var paths [][]*Room
	for {
		p := bfs(g.Start, g.End, banned)
		if p == nil {
			break
		}
		paths = append(paths, p)
		for _, r := range p[1 : len(p)-1] {
			banned[r] = true
		}
	}
	return paths
}

// findPaths gathers disjoint paths and selects the prefix that minimises the
// number of turns required to send all ants across the colony.
func findPaths(g *Graph) [][]*Room {
	all := gatherPaths(g)
	var best [][]*Room
	bestTurns := 1<<31 - 1
	for i := 1; i <= len(all); i++ {
		subset := all[:i]
		lengths := make([]int, i)
		for j, p := range subset {
			lengths[j] = len(p) - 1
		}
		t := computeTurns(g.Ants, lengths)
		if t <= bestTurns {
			bestTurns = t
			best = append([][]*Room{}, subset...)
		} else {
			break
		}
	}
	return best
}

func computeTurns(ants int, lengths []int) int {
	for t := 1; ; t++ {
		total := 0
		for _, l := range lengths {
			if t-l >= 0 {
				total += t - l + 1
			}
		}
		if total >= ants {
			return t
		}
	}
}

func sum(v []int) int {
	s := 0
	for _, x := range v {
		s += x
	}
	return s
}

type antState struct {
	id   int
	path int
	pos  int
}

func pathNames(p []*Room) []string {
	names := make([]string, len(p))
	for i, r := range p {
		names[i] = r.Name
	}
	return names
}

// distributeAnts calculates how many ants each path should receive based on the
// optimal number of turns and returns an ordered assignment list. At most one
// ant is spawned on a path per turn.
// distributeAnts calculates how many ants each path should receive based on the
// optimal turn count. It returns a slice with the number of ants per path.
// assignPaths returns, for each ant in numeric order, the index of the path it
// should take. Ants are assigned one by one to the path that minimises the sum
// of its length and the number of ants already assigned to it. This approach
// reproduces the distribution visible in the official examples.
func assignPaths(paths [][]*Room, ants int) []int {
	n := len(paths)
	lengths := make([]int, n)
	for i, p := range paths {
		lengths[i] = len(p) - 1
	}

	turns := computeTurns(ants, lengths)
	counts := make([]int, n)
	for i, l := range lengths {
		if turns >= l {
			counts[i] = turns - l + 1
		}
	}
	for sum(counts) > ants {
		idx := n - 1
		for idx > 0 && counts[idx] == 0 {
			idx--
		}
		counts[idx]--
	}

	var order []int
	for step := 0; len(order) < ants; step++ {
		for i := 0; i < n; i++ {
			if step < counts[i] {
				order = append(order, i)
			}
		}
	}
	return order
}

func simulateMulti(g *Graph, paths [][]*Room) []string {
	if len(paths) == 0 {
		return nil
	}

	route := assignPaths(paths, g.Ants)
	// build per-path queues of ants
	queues := make([][]int, len(paths))
	for ant, p := range route {
		queues[p] = append(queues[p], ant)
	}

	pos := make([]int, len(route))
	started := make([]bool, len(route))
	occupancy := map[*Room]int{}
	finished := 0
	var moves []string

	for finished < len(route) {
		type evt struct {
			id   int
			room *Room
		}
		var evts []evt

		// move ants already started in ID order
		for id := 0; id < len(route); id++ {
			if !started[id] {
				continue
			}
			p := paths[route[id]]
			if pos[id] < len(p)-1 {
				next := p[pos[id]+1]
				if next == g.End || occupancy[next] == 0 {
					if p[pos[id]] != g.Start {
						delete(occupancy, p[pos[id]])
					}
					pos[id]++
					if next != g.End {
						occupancy[next] = id + 1
					} else {
						finished++
					}
					evts = append(evts, evt{id: id + 1, room: next})
				}
			}
		}

		// spawn at most one ant per path
		for i, q := range queues {
			if len(q) == 0 {
				continue
			}
			ant := q[0]
			next := paths[i][1]
			if next == g.End || occupancy[next] == 0 {
				started[ant] = true
				pos[ant] = 1
				if next != g.End {
					occupancy[next] = ant + 1
				} else {
					finished++
				}
				queues[i] = q[1:]
				evts = append(evts, evt{id: ant + 1, room: next})
			}
		}

		if len(evts) > 0 {
			sort.Slice(evts, func(i, j int) bool { return evts[i].id < evts[j].id })
			line := make([]string, len(evts))
			for i, e := range evts {
				line[i] = fmt.Sprintf("L%d-%s", e.id, e.room.Name)
			}
			moves = append(moves, strings.Join(line, " "))
		}
	}
	return moves
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: lem-in <file>")
		os.Exit(1)
	}
	graph, lines, err := parseInput(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	paths := findPaths(graph)
	if len(paths) == 0 {
		fmt.Println("ERROR: invalid data format")
		return
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	if os.Getenv("DEBUG") == "1" {
		for i, p := range paths {
			fmt.Fprintln(os.Stderr, i, pathNames(p))
		}
	}
	for _, move := range simulateMulti(graph, paths) {
		fmt.Println(move)
	}
}
