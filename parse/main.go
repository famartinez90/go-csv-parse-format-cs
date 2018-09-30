package main

import (
	"encoding/csv"
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

type groupsPerLaw struct {
	Parties    map[string]*groupVotesCount
	Provincies map[string]*groupVotesCount
}

type groupVotesCount struct {
	Afirmativos  int32
	Negativos    int32
	Abstenciones int32
	Ausencias    int32
}

func main() {
	votes, votesGroupedPerLaw := readAndParseFiles()
	formatted, mayorityAll, mayorityPartyOnly, mayorityPartyOnlyAndNoDips := formatForRules(votes, votesGroupedPerLaw)

	outputToCsv(formatted, "transactions")
	outputToCsv(mayorityAll, "mayorityAll")
	outputToCsv(mayorityPartyOnly, "mayorityPartyOnly")
	outputToCsv(mayorityPartyOnlyAndNoDips, "mayorityPartyOnlyAndNoDips")
}

func readAndParseFiles() ([]votingRegistry, map[int]*groupsPerLaw) {
	files, err := ioutil.ReadDir("./csv-votaciones-periodo-reunion-acta/")

	if err != nil {
		panic(err)
	}

	var votes []votingRegistry
	voteGroupsPerLaw := make(map[int]*groupsPerLaw)

	for index, f := range files {
		ley := index + 1
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
			if line[1] == "BLOQUE" ||
				line[0] == "DE VIDO, Julio (Suspendido Art 70 C.N.)" ||
				line[3] == "PRESIDENTE" {
				continue
			}

			vote := string(parseVote(line[3]))
			partido := strings.Replace(line[1], ",", "", -1)
			provincia := strings.Replace(line[2], ",", "", -1)

			votes = append(votes, votingRegistry{
				Diputado:  vote + "=" + minimizeName(line[0]),
				Partido:   vote + "=" + partido,
				Provincia: vote + "=" + provincia,
				Ley:       ley,
			})

			if _, ok := voteGroupsPerLaw[ley]; !ok {
				voteGroupsPerLaw[ley] = &groupsPerLaw{
					Parties:    make(map[string]*groupVotesCount),
					Provincies: make(map[string]*groupVotesCount),
				}
			}

			if _, ok := voteGroupsPerLaw[ley].Parties[partido]; !ok {
				voteGroupsPerLaw[ley].Parties[partido] = &groupVotesCount{
					Afirmativos:  0,
					Negativos:    0,
					Abstenciones: 0,
					Ausencias:    0,
				}
			}

			if _, ok := voteGroupsPerLaw[ley].Provincies[provincia]; !ok {
				voteGroupsPerLaw[ley].Provincies[provincia] = &groupVotesCount{
					Afirmativos:  0,
					Negativos:    0,
					Abstenciones: 0,
					Ausencias:    0,
				}
			}

			switch vote {
			case "A":
				voteGroupsPerLaw[ley].Parties[partido].Afirmativos++
				voteGroupsPerLaw[ley].Provincies[provincia].Afirmativos++
			case "U":
				voteGroupsPerLaw[ley].Parties[partido].Ausencias++
				voteGroupsPerLaw[ley].Provincies[provincia].Ausencias++
			case "N":
				voteGroupsPerLaw[ley].Parties[partido].Negativos++
				voteGroupsPerLaw[ley].Provincies[provincia].Negativos++
			case "B":
				voteGroupsPerLaw[ley].Parties[partido].Abstenciones++
				voteGroupsPerLaw[ley].Provincies[provincia].Abstenciones++
			}
		}
	}

	return votes, voteGroupsPerLaw
}

func formatForRules(votes []votingRegistry, votesGroupedPerLaw map[int]*groupsPerLaw) ([]string, []string, []string, []string) {
	var formatted []string
	var mayorityAll []string
	var mayorityPartyOnly []string
	var mayorityPartyOnlyAndNoDips []string

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

		var mayorityVoteParty []string

		for party, data := range votesGroupedPerLaw[law].Parties {
			mayorityVote := getMayorityVote(data)
			mayorityVoteParty = append(mayorityVoteParty, mayorityVote+"="+party)
		}

		var mayorityVoteProvince []string

		for province, data := range votesGroupedPerLaw[law].Provincies {
			mayorityVote := getMayorityVote(data)
			mayorityVoteProvince = append(mayorityVoteProvince, mayorityVote+"="+province)
		}

		formatted = append(formatted, strconv.Itoa(law)+","+strings.Join(dips, ",")+","+strings.Join(parts, ",")+","+strings.Join(provs, ","))
		mayorityAll = append(mayorityAll, strconv.Itoa(law)+","+strings.Join(dips, ",")+","+strings.Join(mayorityVoteParty, ",")+","+strings.Join(mayorityVoteProvince, ","))
		mayorityPartyOnly = append(mayorityPartyOnly, strconv.Itoa(law)+","+strings.Join(dips, ",")+","+strings.Join(mayorityVoteParty, ","))
		mayorityPartyOnlyAndNoDips = append(mayorityPartyOnlyAndNoDips, strconv.Itoa(law)+","+strings.Join(mayorityVoteParty, ","))
	}

	return formatted, mayorityAll, mayorityPartyOnly, mayorityPartyOnlyAndNoDips
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

func parseVote(vote string) rune {
	var parsed rune

	switch vote {
	case "AFIRMATIVO":
		parsed = 'A'
	case "AUSENTE":
		parsed = 'U'
	case "NEGATIVO":
		parsed = 'N'
	case "PRESIDENTE":
		parsed = 'P'
	case "ABSTENCION":
		parsed = 'B'
	}

	return parsed
}

func minimizeName(name string) string {
	apellidoNombre := strings.Split(name, ",")
	minimized := apellidoNombre[0]

	nombres := strings.Split(apellidoNombre[1], " ")

	for _, nombre := range nombres {
		if nombre != "" {
			minimized += " " + string(nombre[0])
		}
	}

	return minimized
}

func getMayorityVote(data *groupVotesCount) string {
	var maxVotes int32
	var maxVoteLetter string

	if data.Afirmativos > maxVotes {
		maxVotes = data.Afirmativos
		maxVoteLetter = "A"
	}

	if data.Negativos > maxVotes {
		maxVotes = data.Negativos
		maxVoteLetter = "N"
	}

	if data.Abstenciones > maxVotes {
		maxVotes = data.Abstenciones
		maxVoteLetter = "B"
	}

	if data.Ausencias > maxVotes {
		maxVotes = data.Ausencias
		maxVoteLetter = "U"
	}

	return maxVoteLetter
}
