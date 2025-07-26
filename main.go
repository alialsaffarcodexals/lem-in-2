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

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run . <file>")
		return
	}
	filename := os.Args[1]
	content, err := os.ReadFile(filename)
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
			if len(parts) != 2 {
				fmt.Println("ERROR: invalid data format")
				return
			}
			a, b := parts[0], parts[1]
			if rooms[a] == nil || rooms[b] == nil {
				fmt.Println("ERROR: invalid data format")
				return
			}
			graph[a] = append(graph[a], b)
			graph[b] = append(graph[b], a)
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			fmt.Println("ERROR: invalid data format")
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
