package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
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

func bfs(start, end *Room) []*Room {
	type node struct {
		r *Room
		p *node
	}
	visited := map[*Room]bool{start: true}
	queue := []node{{r: start}}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if n.r == end {
			var path []*Room
			for cur := &n; cur != nil; cur = cur.p {
				path = append([]*Room{cur.r}, path...)
			}
			return path
		}
		for _, nb := range n.r.Links {
			if !visited[nb] {
				visited[nb] = true
				queue = append(queue, node{r: nb, p: &n})
			}
		}
	}
	return nil
}

func simulate(g *Graph, path []*Room) []string {
	if len(path) < 2 {
		return nil
	}
	positions := make([]int, len(path)) // 0 means empty
	moves := []string{}
	antsAtEnd := 0
	nextAnt := 1
	for antsAtEnd < g.Ants {
		turn := []string{}
		// move ants from end backwards
		for i := len(path) - 1; i > 0; i-- {
			if positions[i-1] != 0 && (i == len(path)-1 || positions[i] == 0) {
				positions[i] = positions[i-1]
				turn = append(turn, fmt.Sprintf("L%d-%s", positions[i], path[i].Name))
				positions[i-1] = 0
				if i == len(path)-1 {
					antsAtEnd++
				}
			}
		}
		// spawn new ant into path[1]
		if nextAnt <= g.Ants && positions[1] == 0 {
			positions[1] = nextAnt
			turn = append(turn, fmt.Sprintf("L%d-%s", nextAnt, path[1].Name))
			if len(path) == 2 {
				antsAtEnd++
			}
			nextAnt++
		}
		if len(turn) > 0 {
			moves = append(moves, strings.Join(turn, " "))
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
	path := bfs(graph.Start, graph.End)
	if path == nil {
		fmt.Println("ERROR: invalid data format")
		return
	}
	for _, move := range simulate(graph, path) {
		fmt.Println(move)
	}
}
