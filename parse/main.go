package main

import (
	"encoding/csv"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type votingRegistry struct {
	Diputado  string
	Partido   string
	Provincia string
	Ley       int
}

func main() {
	version := flag.String("version", "v1", "Version of the application")

	flag.Parse()

	votes := readAndParseFiles(*version)
	formatted, formattedSinProvincias, formattedSinProvinciasNiPartidos := formatForRules(votes)

	outputToCsv(formatted, "transactions")
	outputToCsv(formattedSinProvincias, "sinProvincias")
	outputToCsv(formattedSinProvinciasNiPartidos, "sinProvinciasNiPartidos")
}

func readAndParseFiles(version string) []votingRegistry {
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

			vote := line[3]

			if version == "v2" {
				votes = append(votes, votingRegistry{
					Diputado:  vote + "=" + strings.Replace(line[0], ",", "", -1),
					Partido:   vote + "=" + strings.Replace(line[1], ",", "", -1),
					Provincia: vote + "=" + strings.Replace(line[2], ",", "", -1),
					Ley:       ley,
				})
			} else {
				votes = append(votes, votingRegistry{
					Diputado:  strings.Replace(line[0], ",", "", -1) + " [" + vote + "]",
					Partido:   strings.Replace(line[1], ",", "", -1) + " [" + vote + "]",
					Provincia: strings.Replace(line[2], ",", "", -1) + " [" + vote + "]",
					Ley:       ley,
				})
			}
		}
	}

	return votes
}

func formatForRules(votes []votingRegistry) ([]string, []string, []string) {
	var formatted []string
	var formattedSinProvincias []string
	var formattedSinProvinciasNiPartidos []string

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

		formatted = append(formatted, strconv.Itoa(law)+","+strings.Join(dips, ",")+","+strings.Join(parts, ",")+","+strings.Join(provs, ","))
		formattedSinProvincias = append(formattedSinProvincias, strconv.Itoa(law)+","+strings.Join(dips, ",")+","+strings.Join(parts, ","))
		formattedSinProvinciasNiPartidos = append(formattedSinProvinciasNiPartidos, strconv.Itoa(law)+","+strings.Join(dips, ","))
	}

	return formatted, formattedSinProvincias, formattedSinProvinciasNiPartidos
}

func outputToCsv(f []string, name string) {
	file, err := os.Create("./" + name + ".csv")

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
