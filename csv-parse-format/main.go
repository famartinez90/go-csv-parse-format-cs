package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type votingRegistry struct {
	Diputado  string
	Partido   string
	Provincia string
	Voto      string
	Ley       int
}

func (v *votingRegistry) print() {
	fmt.Printf("Diputado: %s, Partido: %s, Provincia: %s, Voto: %s, Ley: %d\n", v.Diputado, v.Partido, v.Provincia, v.Voto, v.Ley)
}

func main() {
	votes := readAndParseFiles()
	formatted := formatForRules(votes)

	outputToCsv(formatted)
}

func readAndParseFiles() []votingRegistry {
	files, err := ioutil.ReadDir("./csv-votaciones-periodo-reunion-acta/")

	if err != nil {
		panic(err)
	}

	var votes []votingRegistry

	for ley, f := range files {
		csvFile, err := os.Open("./csv-votaciones-periodo-reunion-acta/" + f.Name())

		if err != nil {
			panic(err)
		}

		defer csvFile.Close()

		reader := csv.NewReader(csvFile)
		reader.LazyQuotes = true

		lines, err := reader.ReadAll()

		if err != nil {
			panic(err)
		}

		for _, line := range lines {
			if line[1] == "BLOQUE" {
				continue
			}

			if line[3] == "AFIRMATIVO" {
				votes = append(votes, votingRegistry{
					Diputado:  strings.Replace(line[0], ",", "", -1),
					Partido:   strings.Replace(line[1], ",", "", -1),
					Provincia: strings.Replace(line[2], ",", "", -1),
					Voto:      strings.Replace(line[3], ",", "", -1),
					Ley:       ley,
				})
			}
		}
	}

	return votes
}

func formatForRules(votes []votingRegistry) []string {
	var formatted []string
	lawsVotes := make(map[int][]votingRegistry)

	for _, vote := range votes {
		lawsVotes[vote.Ley] = append(lawsVotes[vote.Ley], vote)
	}

	for law, data := range lawsVotes {
		var dips []string
		var parts []string
		var provs []string

		for _, vote := range data {
			if !contains(dips, vote.Diputado) {
				dips = append(dips, vote.Diputado)
			}

			if !contains(parts, vote.Partido) {
				parts = append(parts, vote.Partido)
			}

			if !contains(provs, vote.Provincia) {
				provs = append(provs, vote.Provincia)
			}
		}

		formatted = append(formatted, strconv.Itoa(law)+","+strings.Join(dips, ",")+strings.Join(parts, ",")+strings.Join(provs, ","))
	}

	return formatted
}

func outputToCsv(f []string) {
	file, err := os.Create("./transactions.csv")

	if err != nil {
		panic(err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, line := range f {
		err := writer.Write(strings.Split(line, ","))

		if err != nil {
			panic(err)
		}
	}
}

func contains(s []string, search string) bool {
	for _, value := range s {
		if value == search {
			return true
		}
	}

	return false
}
