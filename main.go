package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Println(`usages: 

1. wfsports start names.csv
2. wfsports next roundN.csv

The CSV files do not have headers.
names.csv has only a single column: the names of the players.
roundN.csv has three columns: player1,player2,winner.
`)
		os.Exit(1)
	}

	command := flag.Arg(0)
	filename := flag.Arg(1)
	if err := run(command, filename); err != nil {
		fmt.Printf("wfsports: error: %v\n", err)
		os.Exit(1)
	}
}

func run(command, filename string) error {
	switch command {
	case "start":
		return start(filename)

	case "next":
		return next(filename)

	default:
		return fmt.Errorf("unrecognized command: %s, want start or next", command)
	}
}

func start(filename string) error {
	recs, err := getRecords(filename)
	if err != nil {
		return errors.Wrap(err, "getting records")
	}

	// Randomize order of the names by making them the keys of a map.
	namesMap := make(map[string]bool)
	for _, r := range recs {
		namesMap[r[0]] = true
	}
	var names []string
	for name := range namesMap {
		names = append(names, name)
	}

	switch len(names) {
	case 0:
		return errors.New("no players listed")
	case 1:
		fmt.Println(names[0], "is the winner!")
	default:
		roundFilename := "round1.csv"
		if err := generateRoundFile(roundFilename, names); err != nil {
			return errors.Wrap(err, "generating round file")
		}
		fmt.Println("wrote", roundFilename)
	}

	return nil
}

var roundRx = regexp.MustCompile(`round(\d+).csv`)

func next(doneRoundFilename string) error {
	recs, err := getRecords(doneRoundFilename)
	if err != nil {
		return errors.Wrap(err, "getting records")
	}

	// Get the winners of the done round, and randomize their order by
	// making them the keys of a map.
	namesMap := make(map[string]bool)
	for _, r := range recs {
		namesMap[r[2]] = true
	}
	var names []string
	for name := range namesMap {
		names = append(names, name)
	}

	switch len(names) {
	case 0:
		return errors.New("no players listed")
	case 1:
		fmt.Println(names[0], "is the winner!")
	default:
		m := roundRx.FindStringSubmatch(doneRoundFilename)
		if len(m) != 2 {
			return fmt.Errorf("given filename %s failed to match pattern %s", doneRoundFilename, roundRx)
		}
		k, err := strconv.Atoi(m[1])
		if err != nil {
			return fmt.Errorf("given filename %s has an integer part that somehow doesn't parse (%v). This is a bug.", doneRoundFilename, err)
		}

		roundFilename := fmt.Sprintf("round%d.csv", k+1)
		if err := generateRoundFile(roundFilename, names); err != nil {
			return errors.Wrap(err, "generating round file")
		}
		fmt.Println("wrote", roundFilename)
	}

	return nil
}

func generateRoundFile(roundFilename string, names []string) error {
	roundFile, err := os.Create(roundFilename)
	if err != nil {
		return errors.Wrap(err, "creating round file")
	}
	defer roundFile.Close()
	n := roundDownToPowerOfTwo(len(names))
	// Output pairings so that n players are playing.
	if n > 1 {
		for i := 0; i < n; i += 2 {
			fmt.Fprintf(roundFile, "%s,%s,\n", names[i], names[i+1])
		}
	}

	// Output trivial matches where the remaining players play against themselves and win.
	for i := n; i < len(names); i++ {
		fmt.Fprintf(roundFile, "%s,%s,%s\n", names[i], names[i], names[i])
	}

	return nil
}

// roundDownToPowerOfTwo finds the closest power of 2 below n.
func roundDownToPowerOfTwo(k int) int {
	return int(math.Pow(2.0, math.Floor(math.Log2(float64(k)))))
}

func getRecords(filename string) ([][]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "opening CSV file")
	}
	defer f.Close()
	r := csv.NewReader(f)
	return r.ReadAll()
}
