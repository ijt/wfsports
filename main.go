package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"html/template"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func main() {
	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Println(`usages: 

1. wfsports start names.csv   : generates the first pairings as round1.csv
2. wfsports next roundN.csv   : generates the next pairings as round{N+1}.csv or declares the winner
3. wfsports show roundN.csv   : outputs an HTML page displaying the pairings for round N

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

	case "show":
		return show(filename)

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
			return fmt.Errorf("given filename %s has an integer part that somehow doesn't parse (%v), probably a bug", doneRoundFilename, err)
		}

		roundFilename := fmt.Sprintf("round%d.csv", k+1)
		if err := generateRoundFile(roundFilename, names); err != nil {
			return errors.Wrap(err, "generating round file")
		}
		fmt.Println("wrote", roundFilename)
	}

	return nil
}

func show(roundFilename string) error {
	recs, err := getRecords(roundFilename)
	if err != nil {
		return errors.Wrap(err, "getting records")
	}

	for i := range recs {
		// Eliminate the winner column, because we aren't displaying winners here.
		recs[i] = recs[i][:2]
	}

	t, err := template.New("show").Parse(`
<!DOCTYPE html>
<html>
<head>
<style>
#players {
  font-family: "Trebuchet MS", Arial, Helvetica, sans-serif;
  border-collapse: collapse;
  width: 100%;
}

#players td, #customers th {
  border: 1px solid #ddd;
  padding: 8px;
}

#players tr:nth-child(even){background-color: #f2f2f2;}

#players tr:hover {background-color: #ddd;}

#players th {
  padding-top: 12px;
  padding-bottom: 12px;
  text-align: left;
  background-color: #4CAF50;
  color: white;
}
</style>
</head>
<body>

<table id="players">
  <tr>
    <th>Player 1</th>
    <th>Player 2</th>
  </tr>
{{range .}}
  <tr>
    {{range .}}
      <td>{{.}}</td>
    {{end}}
  </tr>
{{end}}
</table>

</body>
</html>
`)
	if err != nil {
		return errors.Wrap(err, "parsing HTML template")
	}
	htmlFilename := "table.html"
	f, err := os.Create(htmlFilename)
	if err != nil {
		return errors.Wrap(err, "creating HTML output file")
	}
	defer f.Close()
	if err := t.Execute(f, recs); err != nil {
		return errors.Wrap(err, "executing template")
	}
	fmt.Println("wrote", htmlFilename)
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
	var recs [][]string
	line := 1
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "reading line of CSv")
		}
		for _, name := range record {
			if len(strings.TrimSpace(name)) == 0 {
				return nil, fmt.Errorf("empty name encountered on line %d", line)
			}
		}
		recs = append(recs, record)
		line++
	}
	return recs, nil
}
