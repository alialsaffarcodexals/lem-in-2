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
	lineNum := 0
	var pendingStart bool
	var pendingEnd bool
	var links [][2]string

	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
		if strings.HasPrefix(line, "#") {
			if line == "##start" {
				pendingStart = true
			} else if line == "##end" {
				pendingEnd = true
			}
			continue
		}
		if lineNum == 0 {
			ants, err := strconv.Atoi(strings.TrimSpace(line))
			if err != nil || ants <= 0 {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			g.Ants = ants
			lineNum++
			continue
		}
		if strings.Count(line, " ") == 2 && !strings.Contains(line, "-") {
			parts := strings.Split(line, " ")
			if len(parts) != 3 {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			x, err1 := strconv.Atoi(parts[1])
			y, err2 := strconv.Atoi(parts[2])
			if err1 != nil || err2 != nil {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			room := &Room{Name: parts[0], X: x, Y: y}
			g.Rooms[room.Name] = room
			if pendingStart {
				g.Start = room
				pendingStart = false
			}
			if pendingEnd {
				g.End = room
				pendingEnd = false
			}
		} else if strings.Contains(line, "-") && !strings.Contains(line, " ") {
			parts := strings.Split(line, "-")
			if len(parts) != 2 {
				return nil, lines, errors.New("ERROR: invalid data format")
			}
			links = append(links, [2]string{parts[0], parts[1]})
		} else {
			return nil, lines, errors.New("ERROR: invalid data format")
		}
		lineNum++
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

func disjoint(p []*Room, used map[*Room]bool) bool {
	for _, r := range p[1 : len(p)-1] {
		if used[r] {
			return false
		}
	}
	return true
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

	order := make([]int, ants)
	assigned := make([]int, n)
	for k := 0; k < ants; k++ {
		best := 0
		for i := 1; i < n; i++ {
			if lengths[i]+assigned[i] < lengths[best]+assigned[best] {
				best = i
			}
		}
		assigned[best]++
		order[k] = best
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
	for _, l := range lines {
		fmt.Println(l)
	}
	paths := findPaths(graph)
	if os.Getenv("DEBUG") == "1" {
		for i, p := range paths {
			fmt.Fprintln(os.Stderr, i, pathNames(p))
		}
	}
	if len(paths) == 0 {
		fmt.Println("ERROR: invalid data format")
		return
	}

	for _, move := range simulateMulti(graph, paths) {
		fmt.Println(move)
	}
}
