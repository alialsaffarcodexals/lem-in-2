package utils

import "sort"

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
    bestTurns := int(^uint(0) >> 1)
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
            t := ComputeTurns(ants, lengths)
            better := false
            if t < bestTurns {
                better = true
            } else if t == bestTurns {
                for k := 0; k < len(idxs) && k < len(bestIdx); k++ {
                    if idxs[k] < bestIdx[k] {
                        better = true
                        break
                    } else if idxs[k] > bestIdx[k] {
                        break
                    }
                }
                if !better && len(idxs) < len(bestIdx) {
                    better = true
                }
            }
            if better {
                bestTurns = t
                best = append([][]*Room{}, cur...)
                bestIdx = append([]int{}, idxs...)
            }
            return
        }
        rec(i+1, cur, idxs, used)
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

func FindPaths(g *Graph) [][]*Room {
    all := allPaths(g, MaxPaths)
    sort.SliceStable(all, func(i, j int) bool {
        li, lj := len(all[i]), len(all[j])
        if li == lj {
            return i < j
        }
        return li < lj
    })
    return bestDisjointPaths(all, g.Ants)
}

func ComputeTurns(ants int, lengths []int) int {
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

func assignPaths(paths [][]*Room, ants int) []int {
    n := len(paths)
    lengths := make([]int, n)
    for i, p := range paths {
        lengths[i] = len(p) - 1
    }
    t := ComputeTurns(ants, lengths)
    counts := make([]int, n)
    for i, l := range lengths {
        c := t - l + 1
        if c < 0 {
            c = 0
        }
        counts[i] = c
    }
    total := 0
    for _, c := range counts {
        total += c
    }
    for i := n - 1; total > ants && i >= 0; i-- {
        if counts[i] > 0 {
            take := counts[i]
            if take > total-ants {
                take = total - ants
            }
            counts[i] -= take
            total -= take
        }
    }
    var order []int
    active := true
    for active {
        active = false
        for i := 0; i < n; i++ {
            if counts[i] > 0 {
                order = append(order, i)
                counts[i]--
                active = true
            }
        }
    }
    return order
}
