package utils

import (
    "bufio"
    "os"
    "strconv"
    "strings"
)

func ParseInput(path string) (*Graph, []string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, nil, err
    }
    defer file.Close()

    g := &Graph{Rooms: make(map[string]*Room)}
    scanner := bufio.NewScanner(file)
    var lines []string
    var pendingStart, pendingEnd bool
    var links [][2]string
    parsedAnts := false
    coords := map[[2]int]bool{}
    linkSeen := map[string]struct{}{}

    for scanner.Scan() {
        line := scanner.Text()
        lines = append(lines, line)
        if strings.HasPrefix(line, "#") {
            if line == "##start" {
                if pendingStart || g.Start != nil {
                    return nil, lines, LemError{"ERROR: invalid data format", "duplicate start"}
                }
                pendingStart = true
            } else if line == "##end" {
                if pendingEnd || g.End != nil {
                    return nil, lines, LemError{"ERROR: invalid data format", "duplicate end"}
                }
                pendingEnd = true
            }
            continue
        }

        if !parsedAnts {
            ants, err := strconv.Atoi(strings.TrimSpace(line))
            if err != nil || ants <= 0 {
                return nil, lines, LemError{"ERROR: invalid data format", "invalid ants count"}
            }
            if ants > MaxAnts {
                return nil, lines, LemError{"ERROR: ant limit exceeded", "ant count greater than 100000"}
            }
            g.Ants = ants
            parsedAnts = true
            continue
        }

        fields := strings.Fields(line)
        if len(fields) == 3 {
            name := fields[0]
            if strings.HasPrefix(name, "L") || strings.HasPrefix(name, "#") {
                return nil, lines, LemError{"ERROR: invalid data format", "invalid room name '" + name + "'"}
            }
            if _, ok := g.Rooms[name]; ok {
                return nil, lines, LemError{"ERROR: invalid data format", "duplicate room name '" + name + "'"}
            }
            x, err1 := strconv.Atoi(fields[1])
            y, err2 := strconv.Atoi(fields[2])
            if err1 != nil || err2 != nil {
                return nil, lines, LemError{"ERROR: invalid data format", "invalid room line"}
            }
            if coords[[2]int{x, y}] {
                return nil, lines, LemError{"ERROR: invalid data format", "duplicate coordinates " + fields[1] + " " + fields[2]}
            }
            coords[[2]int{x, y}] = true
            r := &Room{Name: name, X: x, Y: y}
            g.Rooms[name] = r
            if pendingStart {
                g.Start = r
                pendingStart = false
            }
            if pendingEnd {
                g.End = r
                pendingEnd = false
            }
            continue
        }

        if strings.Count(line, "-") == 1 && !strings.Contains(line, " ") {
            parts := strings.Split(line, "-")
            if parts[0] == parts[1] {
                return nil, lines, LemError{"ERROR: invalid data format", "self-loop link " + parts[0] + "-" + parts[1]}
            }
            a, b := parts[0], parts[1]
            if b < a {
                a, b = b, a
            }
            key := a + "-" + b
            if _, ok := linkSeen[key]; ok {
                return nil, lines, LemError{"ERROR: invalid data format", "duplicate link " + key}
            }
            linkSeen[key] = struct{}{}
            links = append(links, [2]string{parts[0], parts[1]})
            continue
        }

        return nil, lines, LemError{"ERROR: invalid data format", "invalid line"}
    }
    if err := scanner.Err(); err != nil {
        return nil, lines, err
    }

    if g.Start == nil || g.End == nil {
        return nil, lines, LemError{"ERROR: invalid data format", "missing start or end"}
    }

    for _, l := range links {
        a, ok1 := g.Rooms[l[0]]
        b, ok2 := g.Rooms[l[1]]
        if !ok1 || !ok2 {
            return nil, lines, LemError{"ERROR: invalid data format", "unknown room in link '" + l[0] + "'"}
        }
        if !hasNeighbor(a, b) {
            a.Links = append(a.Links, b)
        }
        if !hasNeighbor(b, a) {
            b.Links = append(b.Links, a)
        }
    }
    return g, lines, nil
}

func hasNeighbor(r, other *Room) bool {
    for _, nb := range r.Links {
        if nb == other {
            return true
        }
    }
    return false
}
