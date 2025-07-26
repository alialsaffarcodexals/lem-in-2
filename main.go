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
		// explore neighbors in alphabetical order for determinism
		sort.Slice(cur.Links, func(i, j int) bool { return cur.Links[i].Name < cur.Links[j].Name })
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
	sort.Slice(best, func(i, j int) bool {
		if len(best[i]) == len(best[j]) {
			return strings.Join(pathNames(best[i]), " ") < strings.Join(pathNames(best[j]), " ")
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

func assignAnts(ants int, lengths []int, turns int) []int {
	assigns := make([]int, len(lengths))
	for i, l := range lengths {
		if turns-l >= 0 {
			assigns[i] = turns - l + 1
		}
	}
	sum := 0
	for _, a := range assigns {
		sum += a
	}
	for sum > ants {
		for i := len(assigns) - 1; i >= 0 && sum > ants; i-- {
			if assigns[i] > 0 {
				assigns[i]--
				sum--
			}
		}
	}
	return assigns
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

func simulateMulti(g *Graph, paths [][]*Room) []string {
	if len(paths) == 0 {
		return nil
	}
	lengths := make([]int, len(paths))
	for i, p := range paths {
		lengths[i] = len(p) - 1
	}
	turns := computeTurns(g.Ants, lengths)
	assigns := assignAnts(g.Ants, lengths, turns)

	var moves []string
	occupancy := map[*Room]int{}
	var active []antState
	nextAnt := 1
	remaining := g.Ants
	for remaining > 0 || len(active) > 0 {
		var line []string

		// move existing ants
		for i := range active {
			a := &active[i]
			path := paths[a.path]
			if a.pos < len(path)-1 {
				nextRoom := path[a.pos+1]
				if nextRoom == g.End || occupancy[nextRoom] == 0 {
					if path[a.pos] != g.Start {
						delete(occupancy, path[a.pos])
					}
					a.pos++
					if nextRoom != g.End {
						occupancy[nextRoom] = a.id
					}
					line = append(line, fmt.Sprintf("L%d-%s", a.id, nextRoom.Name))
				}
			}
		}

		// remove finished ants
		j := 0
		for _, a := range active {
			if a.pos < len(paths[a.path])-1 {
				active[j] = a
				j++
			}
		}
		active = active[:j]

		// spawn new ants
		for i, p := range paths {
			if assigns[i] > 0 && nextAnt <= g.Ants {
				nextRoom := p[1]
				if nextRoom == g.End || occupancy[nextRoom] == 0 {
					active = append(active, antState{id: nextAnt, path: i, pos: 1})
					if nextRoom != g.End {
						occupancy[nextRoom] = nextAnt
					}
					line = append(line, fmt.Sprintf("L%d-%s", nextAnt, nextRoom.Name))
					assigns[i]--
					nextAnt++
					remaining--
				}
			}
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
	if len(paths) == 0 {
		fmt.Println("ERROR: invalid data format")
		return
	}

	for _, move := range simulateMulti(graph, paths) {
		fmt.Println(move)
	}
}
