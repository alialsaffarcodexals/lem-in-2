package main

import (
	"fmt"
	"os"

	"lem-in/internal/utils"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: lem-in <file>")
		os.Exit(1)
	}
	graph, lines, err := utils.ParseInput(os.Args[1])
	if err != nil {
		if e, ok := err.(utils.LemError); ok {
			fmt.Println(e.Msg)
			fmt.Println("Reason: " + e.Reason)
		} else {
			fmt.Println(err.Error())
		}
		os.Exit(1)
	}
	paths := utils.FindPaths(graph)
	if len(paths) == 0 {
		fmt.Println("ERROR: invalid data format")
		fmt.Println("Reason: no path from start to end")
		os.Exit(1)
	}
	for _, l := range lines {
		fmt.Println(l)
	}
	fmt.Println()
	for _, m := range utils.SimulateMulti(graph, paths) {
		fmt.Println(m)
	}
}
