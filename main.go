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

func allPaths(start, end *Room, limit int) [][]*Room {
	var paths [][]*Room
	var dfs func(*Room, map[*Room]bool, []*Room)
	dfs = func(cur *Room, visited map[*Room]bool, path []*Room) {
		if len(path) > limit {
			return
		}
		if cur == end {
			cp := append([]*Room{}, path...)
			paths = append(paths, cp)
			return
		}
		visited[cur] = true
		// explore neighbors in the input order to match example outputs
		for _, nb := range cur.Links {
			if !visited[nb] {
				dfs(nb, visited, append(path, nb))
			}
		}
		delete(visited, cur)
	}
	dfs(start, map[*Room]bool{}, []*Room{start})
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

func findPaths(g *Graph) [][]*Room {
	candidates := allPaths(g.Start, g.End, 20)
	var best [][]*Room
	bestTurns := 1<<31 - 1
	var explore func(int, [][]*Room, map[*Room]bool)
	explore = func(idx int, cur [][]*Room, used map[*Room]bool) {
		if idx == len(candidates) {
			if len(cur) == 0 {
				return
			}
			lengths := make([]int, len(cur))
			for i, p := range cur {
				lengths[i] = len(p) - 1
			}
			t := computeTurns(g.Ants, lengths)
			if t < bestTurns {
				bestTurns = t
				best = append([][]*Room{}, cur...)
			}
			return
		}
		p := candidates[idx]
		if disjoint(p, used) {
			for _, r := range p[1 : len(p)-1] {
				used[r] = true
			}
			explore(idx+1, append(cur, p), used)
			for _, r := range p[1 : len(p)-1] {
				delete(used, r)
			}
		}
		// skip
		explore(idx+1, cur, used)
	}
	explore(0, nil, map[*Room]bool{})
	sort.SliceStable(best, func(i, j int) bool {
		if len(best[i]) == len(best[j]) {
			return i < j
		}
		return len(best[i]) < len(best[j])
	})
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

func distributeAnts(paths [][]*Room, ants int) []int {
	n := len(paths)
	lengths := make([]int, n)
	for i, p := range paths {
		lengths[i] = len(p) - 1
	}

	turns := computeTurns(ants, lengths)
	counts := make([]int, n)
	for i, l := range lengths {
		if turns-l >= 0 {
			counts[i] = turns - l + 1
		}
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	for total > ants {
		for i := n - 1; i >= 0 && total > ants; i-- {
			if counts[i] > 0 {
				counts[i]--
				total--
			}
		}
	}

	offsets := make([]int, n)
	base := lengths[0]
	for i := 0; i < n; i++ {
		offsets[i] = lengths[i] - base
	}

	assigned := make([]int, n)
	assignment := make([]int, 0, ants)
	for len(assignment) < ants {
		best := -1
		bestVal := 0
		for i := 0; i < n; i++ {
			if assigned[i] >= counts[i] {
				continue
			}
			val := offsets[i] + assigned[i]
			if best == -1 || val < bestVal || (val == bestVal && offsets[i] > offsets[best]) {
				best = i
				bestVal = val
			}
		}
		if best == -1 {
			break
		}
		assignment = append(assignment, best)
		assigned[best]++
	}
	return assignment
}

func simulateMulti(g *Graph, paths [][]*Room) []string {
	if len(paths) == 0 {
		return nil
	}
	assignment := distributeAnts(paths, g.Ants)
	pos := make([]int, len(assignment))
	occupancy := map[*Room]int{}
	started := 0
	finished := 0
	var moves []string

	for finished < len(assignment) {
		var line []string

		// move ants already in motion
		for i := 0; i < started; i++ {
			p := paths[assignment[i]]
			if pos[i] < len(p)-1 {
				next := p[pos[i]+1]
				if next == g.End || occupancy[next] == 0 {
					if p[pos[i]] != g.Start {
						delete(occupancy, p[pos[i]])
					}
					pos[i]++
					if next != g.End {
						occupancy[next] = i + 1
					} else {
						finished++
					}
					line = append(line, fmt.Sprintf("L%d-%s", i+1, next.Name))
				}
			}
		}

		// spawn new ants in order
		spawned := make(map[int]bool)
		for started < len(assignment) {
			idx := assignment[started]
			if spawned[idx] {
				break
			}
			p := paths[idx]
			next := p[1]
			if next != g.End && occupancy[next] != 0 {
				break
			}
			spawned[idx] = true
			pos[started] = 1
			if next != g.End {
				occupancy[next] = started + 1
			} else {
				finished++
			}
			line = append(line, fmt.Sprintf("L%d-%s", started+1, next.Name))
			started++
		}

		if len(line) > 0 {
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
