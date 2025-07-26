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

const maxPaths = 100

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
		visited[r] = true
		path = append(path, r)
		for _, nb := range r.Links {
			if !visited[nb] {
				dfs(nb)
			}
		}
		path = path[:len(path)-1]
		visited[r] = false
	}
	dfs(g.Start)
	return res
}

func bestDisjointPaths(all [][]*Room, ants int) [][]*Room {
	bestTurns := 1<<31 - 1
	var best [][]*Room
	var bestIdx []int
	var rec func(int, [][]*Room, []int, map[*Room]bool)
	rec = func(i int, cur [][]*Room, idxs []int, used map[*Room]bool) {
		if i == len(all) {
			if len(cur) == 0 {
				return
			}
			lengths := make([]int, len(cur))
			for j, p := range cur {
				lengths[j] = len(p) - 1
			}
			t := computeTurns(ants, lengths)
			better := false
			if t < bestTurns {
				better = true
			} else if t == bestTurns {
				if len(cur) > len(best) {
					better = true
				} else if len(cur) == len(best) {
					for k := 0; k < len(idxs) && k < len(bestIdx); k++ {
						if idxs[k] < bestIdx[k] {
							better = true
							break
						} else if idxs[k] > bestIdx[k] {
							break
						}
					}
				}
			}
			if better {
				bestTurns = t
				best = append([][]*Room{}, cur...)
				bestIdx = append([]int{}, idxs...)
			}
			return
		}
		// skip current path
		rec(i+1, cur, idxs, used)

		// try including current path if disjoint
		p := all[i]
		valid := true
		for _, r := range p[1 : len(p)-1] {
			if used[r] {
				valid = false
				break
			}
		}
		if valid {
			for _, r := range p[1 : len(p)-1] {
				used[r] = true
			}
			rec(i+1, append(cur, p), append(idxs, i), used)
			for _, r := range p[1 : len(p)-1] {
				delete(used, r)
			}
		}
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
// assignPaths distributes ants based on the optimal turn count computed from
// the path lengths. Ants are spawned in a round-robin fashion starting from the
// shortest path so that no path launches more than one ant per turn.
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
	excess := sum(counts) - ants
	idx := n - 1
	for excess > 0 {
		if counts[idx] > 0 {
			counts[idx]--
			excess--
		}
		idx--
		if idx < 0 {
			idx = n - 1
		}
	}

	start := 0
	for i := 1; i < n; i++ {
		if lengths[i] < lengths[start] {
			start = i
		}
	}

	var order []int
	base := lengths[start]
	for step := 0; ; step++ {
		added := false
		for j := 0; j < n; j++ {
			i := (j + start) % n
			offset := lengths[i] - base
			if step >= offset && step-offset < counts[i] {
				order = append(order, i)
				added = true
			}
		}
		if !added {
			break
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
