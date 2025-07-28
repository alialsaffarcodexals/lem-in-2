package utils

import (
	"fmt"
	"sort"
	"strings"
)

type antState struct {
	id   int
	path int
	pos  int
}

func SimulateMulti(g *Graph, paths [][]*Room) []string {
	if len(paths) == 0 {
		return nil
	}
	route := assignPaths(paths, g.Ants)
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
