package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/jlouis/glicko2"
)

// Player represents raw stats for one player
type Player struct {
	Name              string
	Frags             int
	Deaths            int
	Suicides          int
	Damage            int
	DamageDealt       int
	WeaponDamage      [7]int
	WeaponDamageDealt [7]int
}

// Game represents raw stats for one duel
type Game struct {
	stage   string
	Time    int64
	Mode    int
	Map     string
	Players []Player
	Counts  bool
}

// Stats represents aggregated stats of one player
type Stats struct {
	Rank              int
	Name              string
	Elo               int
	opponents         []glicko2.Opponent
	Games             int
	Wins              int
	Losses            int
	WinRate           float32
	Frags             int
	Deaths            int
	Suicides          int
	Damage            int
	DamageDealt       int
	WeaponDamage      [7]int
	WeaponDamageDealt [7]int
	Accuracy          float32
	WeaponAccuracy    [7]float32
	KPD               float32
}

// Collection represents a collection of games, consisting of one or more files
type Collection struct {
	Path  string
	Games map[string]*Game
	Stats map[string]*[6]map[string]*Stats
}

// GlickoOpponent implements the glicko2.Opponent interface
type GlickoOpponent struct {
	Rx     float64
	Rdx    float64
	Sigmax float64
	Sjx    float64
}

// R returns R
func (e GlickoOpponent) R() float64 {
	return e.Rx
}

// RD returns Rd
func (e GlickoOpponent) RD() float64 {
	return e.Rdx
}

// Sigma returns Sigma
func (e GlickoOpponent) Sigma() float64 {
	return e.Sigmax
}

// SJ returns Sj, the result of the match
func (e GlickoOpponent) SJ() float64 {
	return e.Sjx
}

// NewGlickoOpponent creates a new Opponent with default values
func NewGlickoOpponent() glicko2.Opponent {
	return GlickoOpponent{1500, 300, 0.06, 0}
}

// GlickoOpponentFromMatch creates a new Opponent with the same values as challenger but with a new Sj
func GlickoOpponentFromMatch(challenger glicko2.Opponent, playerFrags int, challengerFrags int) glicko2.Opponent {
	w := 0.5
	if playerFrags > challengerFrags {
		w = 1
	} else if challengerFrags > playerFrags {
		w = 0
	}
	return GlickoOpponent{challenger.R(), challenger.RD(), challenger.Sigma(), w}
}

// ModeNames is a list of Sauerbraten mode names
var ModeNames = []string{"ffa", "coop", "teamplay", "insta", "instateam", "effic"}

// ModeNamesTitle is a list of Sauerbraten mode names in title case
var ModeNamesTitle = []string{"FFA", "Co-Op", "Teamplay", "Insta", "Instateam", "Effic"}

var lastEntryTime int64
var allowAfterTime int64
var globalStats map[string]*[6]map[string]*Stats
var globalElos [6]map[string]glicko2.Opponent

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./aggregator file|directory [file|directory [...]]")
	}

	globalStats = make(map[string]*[6]map[string]*Stats)
	for i := range globalElos {
		globalElos[i] = make(map[string]glicko2.Opponent)
	}
	ReadGlobalStats()

	for _, path := range os.Args[1:] {
		stat, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}
		if !stat.IsDir() {
			log.Fatalf("%s is not a directory\n", path)
		}

		fileinfos, err := ioutil.ReadDir(path)
		if err != nil {
			log.Fatal(err)
		}

		files := make([]string, 0)
		stages := make([]string, 0)

		for _, fileinfo := range fileinfos {
			path2 := filepath.Join(path, fileinfo.Name())

			stat, err = os.Stat(path2)
			if err != nil {
				log.Fatal(err)
			}

			if !stat.IsDir() {
				log.Fatalf("%s is not a directory\n", path)
			}

			fileinfos, err := ioutil.ReadDir(path2)
			if err != nil {
				log.Fatal(err)
			}

			for _, fileinfo2 := range fileinfos {
				files = append(files, filepath.Join(path2, fileinfo2.Name()))
				stages = append(stages, fileinfo.Name())
			}
		}

		c := Collection{path, make(map[string]*Game), make(map[string]*[6]map[string]*Stats)}

		for _, stage := range stages {
			if _, ok := c.Stats[stage]; !ok {
				c.Stats[stage] = &[6]map[string]*Stats{}
				for k := 0; k < 6; k++ {
					c.Stats[stage][k] = make(map[string]*Stats)
				}
			}
		}

		for k := range files {
			c.ProcessFile(files[k], stages[k])
		}
		c.ReduceDuplicates()
		c.ProcessGames()
		IntegrateStats(&c.Stats)
		UpdateStats(&c.Stats)
		SetRanks(&c.Stats)

		// WriteCSV(&c.Stats, filepart)
		// WriteHTML(&c.Stats, filepart)
	}

	UpdateStats(&globalStats)
	SetRanks(&globalStats)
	// WriteCSV(&globalStats)
	WriteHTML(&globalStats)
	WriteGlobalStats()
}

type stateFile struct {
	LastTimestamp int64
	Stats         map[string]*[6]map[string]*Stats
	Elos          [6]map[string]GlickoOpponent
}

var sf = stateFile{}

// ReadGlobalStats unmarshals ./state.json into globalStats
func ReadGlobalStats() {
	file, err := ioutil.ReadFile("./state.json")
	if err == nil {
		json.Unmarshal(file, &sf)
	}

	allowAfterTime = sf.LastTimestamp
	for stage, modes := range sf.Stats {
		globalStats[stage] = new([6]map[string]*Stats)
		for mode, stats := range modes {
			if stats != nil {
				globalStats[stage][mode] = stats
			}
		}
	}
	for mode, elos := range sf.Elos {
		if elos != nil {
			for player, elo := range elos {
				globalElos[mode][player] = GlickoOpponent{elo.R(), elo.RD(), elo.Sigma(), 0}
			}
		}
	}
}

// WriteGlobalStats marshals globalStats into ./state.json
func WriteGlobalStats() {
	sf.LastTimestamp = lastEntryTime
	if sf.Stats == nil {
		sf.Stats = make(map[string]*[6]map[string]*Stats)
	}
	sf.Stats = globalStats
	for mode, elos := range globalElos {
		if sf.Elos[mode] == nil {
			sf.Elos[mode] = make(map[string]GlickoOpponent)
		}
		for player, elo := range elos {
			sf.Elos[mode][player] = GlickoOpponent{elo.R(), elo.RD(), elo.Sigma(), 0}
		}
	}

	sfJSON, err := json.Marshal(sf)
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile("./state.json", sfJSON, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// ProcessFile reads from a space separated file (SauerDuels style) and parses the values
// into c.Games
func (c Collection) ProcessFile(path string, stage string) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ' '
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, line := range records {
		gTime, _ := strconv.ParseInt(line[0], 10, 64)
		gMode, _ := strconv.Atoi(line[1])
		gMap := line[2]

		gKey := fmt.Sprintf("%s/%d", path, gTime)

		if _, ok := c.Games[gKey]; !ok {
			c.Games[gKey] = &Game{stage, gTime, gMode, gMap, make([]Player, 0), true}
		}

		p := Player{}
		p.Name = line[3]
		p.Frags, _ = strconv.Atoi(line[4])
		p.Deaths, _ = strconv.Atoi(line[5])
		p.Damage, _ = strconv.Atoi(line[6])
		p.DamageDealt, _ = strconv.Atoi(line[7])
		p.Suicides, _ = strconv.Atoi(line[8])

		for i := 0; i < 7; i++ {
			p.WeaponDamage[i], _ = strconv.Atoi(line[9+2*i])
			p.WeaponDamageDealt[i], _ = strconv.Atoi(line[10+2*i])
		}

		c.Games[gKey].Players = append(c.Games[gKey].Players, p)
	}

	for key, game := range c.Games {
		if len(game.Players) != 2 {
			delete(c.Games, key)
		}
	}
}

func pairString(g *Game) string {
	players := make([]string, 0)
	for _, p := range g.Players {
		players = append(players, p.Name)
	}
	sort.Strings(players)
	return strings.Join(players, " ")
}

// ReduceDuplicates reduces games with the same players into one game
func (c Collection) ReduceDuplicates() {
	gamesByPlayers := make(map[string][]string)

	for k, g := range c.Games {
		pairString := pairString(g)
		if _, ok := gamesByPlayers[pairString]; !ok {
			gamesByPlayers[pairString] = make([]string, 0)
		}
		gamesByPlayers[pairString] = append(gamesByPlayers[pairString], k)
	}

	for _, games := range gamesByPlayers {
		if len(games) > 1 {
			wins0 := 0
			wins1 := 0
			for _, g := range games {
				c.Games[g].Counts = false
				if c.Games[g].Players[0].Frags > c.Games[g].Players[1].Frags {
					wins0++
				} else if c.Games[g].Players[1].Frags > c.Games[g].Players[0].Frags {
					wins1++
				}
			}
			for _, g := range games {
				if c.Games[g].Players[0].Frags > c.Games[g].Players[1].Frags && wins0 > wins1 {
					c.Games[g].Counts = true
					break
				} else if c.Games[g].Players[1].Frags > c.Games[g].Players[0].Frags && wins1 > wins0 {
					c.Games[g].Counts = true
					break
				}
			}
		}
	}
}

// ProcessGames extracts and aggregates Stats from c.Games
func (c Collection) ProcessGames() {
	gs := make([]*Game, 0)
	for _, g := range c.Games {
		gs = append(gs, g)
	}
	sort.Sort(ByTime(gs))

	if len(gs) > 0 {
		if gs[0].Time <= allowAfterTime {
			// log.Printf("warning: processing '%s': found timestamp similar to or earlier than the one found in 'state.json'. Processing this file may result in inaccuracies or duplications\n", c.Path)
		}
		if gs[len(gs)-1].Time > lastEntryTime {
			lastEntryTime = gs[len(gs)-1].Time
		}
	}

	for stage := range c.Stats {
		if _, ok2 := globalStats[stage]; !ok2 {
			globalStats[stage] = new([6]map[string]*Stats)
			for mode := range globalStats[stage] {
				globalStats[stage][mode] = make(map[string]*Stats)
			}
		}
	}

	for _, g := range gs {
		for _, p := range g.Players {
			if _, ok := c.Stats[g.stage][g.Mode][p.Name]; !ok {
				c.Stats[g.stage][g.Mode][p.Name] = &Stats{}
				c.Stats[g.stage][g.Mode][p.Name].Name = p.Name

				if _, ok2 := globalElos[g.Mode][p.Name]; !ok2 {
					globalElos[g.Mode][p.Name] = NewGlickoOpponent()
				}
			}
			c.Stats[g.stage][g.Mode][p.Name].Deaths += p.Deaths
			c.Stats[g.stage][g.Mode][p.Name].Suicides += p.Suicides
			c.Stats[g.stage][g.Mode][p.Name].Damage += p.Damage
			c.Stats[g.stage][g.Mode][p.Name].DamageDealt += p.DamageDealt
			for i := 0; i < 7; i++ {
				c.Stats[g.stage][g.Mode][p.Name].WeaponDamage[i] += p.WeaponDamage[i]
				c.Stats[g.stage][g.Mode][p.Name].WeaponDamageDealt[i] += p.WeaponDamageDealt[i]
			}
		}
		// 99999 frags is a special entry that designates the winner of a forfeited game
		if g.Players[0].Frags < 99999 && g.Players[1].Frags < 99999 {
			c.Stats[g.stage][g.Mode][g.Players[0].Name].Games++
			c.Stats[g.stage][g.Mode][g.Players[0].Name].Frags += g.Players[0].Frags
			c.Stats[g.stage][g.Mode][g.Players[1].Name].Games++
			c.Stats[g.stage][g.Mode][g.Players[1].Name].Frags += g.Players[1].Frags
		}
		if g.Players[0].Frags > g.Players[1].Frags && g.Players[0].Frags < 99999 {
			c.Stats[g.stage][g.Mode][g.Players[0].Name].Wins++
			c.Stats[g.stage][g.Mode][g.Players[1].Name].Losses++
		} else if g.Players[1].Frags > g.Players[0].Frags && g.Players[1].Frags < 99999 {
			c.Stats[g.stage][g.Mode][g.Players[1].Name].Wins++
			c.Stats[g.stage][g.Mode][g.Players[0].Name].Losses++
		}
		if g.Counts {
			c.Stats[g.stage][g.Mode][g.Players[0].Name].opponents = append(c.Stats[g.stage][g.Mode][g.Players[0].Name].opponents,
				GlickoOpponentFromMatch(globalElos[g.Mode][g.Players[1].Name], g.Players[0].Frags, g.Players[1].Frags))

			c.Stats[g.stage][g.Mode][g.Players[1].Name].opponents = append(c.Stats[g.stage][g.Mode][g.Players[1].Name].opponents,
				GlickoOpponentFromMatch(globalElos[g.Mode][g.Players[0].Name], g.Players[1].Frags, g.Players[0].Frags))
		}
	}
}

// IntegrateStats integrates stats from stats into globalStats
func IntegrateStats(stats *map[string]*[6]map[string]*Stats) {
	for stage, stageStats := range *stats {
		for mode, modeStats := range stageStats {
			for player, stats := range modeStats {
				if _, ok2 := globalStats[stage][mode][player]; !ok2 {
					globalStats[stage][mode][player] = &Stats{}
					globalStats[stage][mode][player].Name = stats.Name
				}

				globalStats[stage][mode][player].Games += stats.Games
				globalStats[stage][mode][player].Frags += stats.Frags
				globalStats[stage][mode][player].Deaths += stats.Deaths
				globalStats[stage][mode][player].Suicides += stats.Suicides
				globalStats[stage][mode][player].Damage += stats.Damage
				globalStats[stage][mode][player].DamageDealt += stats.DamageDealt
				globalStats[stage][mode][player].Wins += stats.Wins
				globalStats[stage][mode][player].Losses += stats.Losses

				for i := 0; i < 7; i++ {
					globalStats[stage][mode][player].WeaponDamage[i] += stats.WeaponDamage[i]
					globalStats[stage][mode][player].WeaponDamageDealt[i] += stats.WeaponDamageDealt[i]
				}

				globalStats[stage][mode][player].opponents = append(globalStats[stage][mode][player].opponents, stats.opponents...)
			}
		}
	}

	for stage, stageStats := range globalStats {
		for mode, modeStats := range stageStats {
			if (*stats)[stage] == nil || len((*stats)[stage][mode]) == 0 {
				continue
			}
			for player, stats := range modeStats {
				if len(stats.opponents) == 0 {
					e := globalElos[mode][player]
					nrd := glicko2.Skip(e.R(), e.RD(), e.Sigma())
					globalElos[mode][player] = GlickoOpponent{e.R(), nrd, e.Sigma(), 0}
				} else {
					e := globalElos[mode][player]
					nr, nrd, nsigma := glicko2.Rank(e.R(), e.RD(), e.Sigma(), stats.opponents, 0.5)
					globalElos[mode][player] = GlickoOpponent{nr, nrd, nsigma, 0}
					globalStats[stage][mode][player].Elo = int(nr)
					stats.opponents = []glicko2.Opponent{}
				}
			}
		}
	}
}

// UpdateStats does post processing on stats to calculate WinRate, Accuracy, KPD, and Elo
func UpdateStats(stats *map[string]*[6]map[string]*Stats) {
	for _, stageStats := range *stats {
		for mode, modeStats := range stageStats {
			for _, s := range modeStats {
				s.WinRate = 0
				if s.Games != 0 {
					s.WinRate = float32(s.Wins) / float32(s.Games)
				}
				s.Accuracy = 0
				if s.Damage != 0 {
					s.Accuracy = float32(s.DamageDealt) / float32(s.Damage)
				}
				for i := 0; i < 7; i++ {
					s.WeaponAccuracy[i] = 0
					if s.WeaponDamage[i] != 0 {
						s.WeaponAccuracy[i] = float32(s.WeaponDamageDealt[i]) / float32(s.WeaponDamage[i])
					}
				}
				s.KPD = 0
				if s.Deaths != 0 {
					s.KPD = float32(s.Frags) / float32(s.Deaths)
				}
				s.Elo = int(globalElos[mode][s.Name].R())
			}
		}
	}
}

// SetRanks calculates the ranks for stats
func SetRanks(stats *map[string]*[6]map[string]*Stats) {
	for _, stageStats := range *stats {
		for _, modeStats := range stageStats {
			ps := make([]*Stats, 0)
			for _, s := range modeStats {
				ps = append(ps, s)
			}
			sort.Sort(ByElo(ps))

			counter := 1
			lastRank := 0
			lastElo := 0
			for _, s := range ps {
				if s.Elo != lastElo {
					lastRank = counter
				}
				s.Rank = lastRank
				lastElo = s.Elo
				counter++
			}
		}
	}
}

// WriteCSV writes stats to one CSV file per mode
func WriteCSV(stats *map[string]*[6]map[string]*Stats) {
	for stage, stageStats := range *stats {
		for k, modeStats := range stageStats {
			if len(modeStats) == 0 {
				continue
			}

			f, err := os.Create(ModeNames[k] + "_" + stage + ".csv")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			writer := csv.NewWriter(f)
			defer writer.Flush()

			writer.Write([]string{"Rank", "Player", "ELO", "Games", "Wins", "Losses", "Frags", "Deaths", "Suicides", "Damage", "Damage Dealt", "Chainsaw Damage", "Chainsaw Dealt", "Shotgun Damage", "Shotgun Damage Dealt", "CG Damage", "CG Damage Dealt", "RL Damage", "RL Damage Dealt", "Rifle Damage", "Rifle Damage Dealt", "GL Damage", "GL Damage Dealt", "Pistol Damage", "Pistol Damage Dealt"})

			ps := make([]*Stats, 0)
			for _, s := range modeStats {
				ps = append(ps, s)
			}
			sort.Sort(ByElo(ps))

			for _, s := range ps {
				line := strings.Split(fmt.Sprintf("%d %s %d %d %d %d %d %d %d %d %d", s.Rank, s.Name, s.Elo, s.Games, s.Wins, s.Losses, s.Frags, s.Deaths, s.Suicides, s.Damage, s.DamageDealt), " ")
				for i := 0; i < 7; i++ {
					line = append(line, fmt.Sprintf("%d", s.WeaponDamage[i]))
					line = append(line, fmt.Sprintf("%d", s.WeaponDamageDealt[i]))
				}
				writer.Write(line)
			}

			fmt.Printf("wrote %s\n", f.Name())
		}
	}
}

// WriteHTML writes stats to one HTML file per mode
func WriteHTML(stats *map[string]*[6]map[string]*Stats) {
	for stage, stageStats := range *stats {
		for k, modeStats := range stageStats {
			if len(modeStats) == 0 {
				continue
			}

			f, err := os.Create(ModeNames[k] + "_" + stage + ".html")
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			funcMap := template.FuncMap{
				"percent": func(i float32) string {
					return fmt.Sprintf("%02.0f", i*100) + "%"
				},
				"trim": func(i float32) string {
					return fmt.Sprintf("%.3f", i)
				},
				"modename": func() string {
					return ModeNamesTitle[k]
				},
				"finals": func() string {
					if stage == "finals" {
						return strings.Title(stage)
					}
					return ""
				},
			}

			tmpl, err := template.New("playerStats").Funcs(funcMap).Parse(`---
title:  {{modename}} {{finals}}
---
<table>
  <thead>
    <tr>
      <th>Rank</th><th>Player</th><th>ELO</th><th>Games</th><th>Wins</th><th>Losses</th><th>Win Ratio</th><th>Frags</th><th>Deaths</th><th>K/D</th><th>Suicides</th><th>Accuracy</th>
    </tr>
  </thead>
  <tbody>
    {{range .}}
      {{if .Games}}
        <tr>
          <td>{{.Rank}}</td><td class="left-align">{{.Name}}</td><td>{{.Elo}}</td><td>{{.Games}}</td><td>{{.Wins}}</td><td>{{.Losses}}</td><td>{{trim .WinRate}}</td><td>{{.Frags}}</td><td>{{.Deaths}}</td><td>{{trim .KPD}}</td><td>{{.Suicides}}</td><td>{{percent .Accuracy}}</td>
        </tr>
      {{end}}
    {{end}}
  </tbody>
</table>
<br />
<table>
  <thead>
    <tr>
      <th>Rank</th><th>Player</th><th>Accuracy</th><th>Chainsaw Accuracy</th><th>Shotgun Accuracy</th><th>CG Accuracy</th><th>RL Accuracy</th><th>Rifle Accuracy</th><th>GL Accuracy</th><th>Pistol Accuracy</th>
    </tr>
  </thead>
  <tbody>
    {{range .}}
      {{if .Games}}
        <tr>
          <td>{{.Rank}}</td><td class="left-align">{{.Name}}</td><td>{{percent .Accuracy}}</td>{{range .WeaponAccuracy}}<td>{{percent .}}</td>{{end}}
        </tr>
      {{end}}
    {{end}}
  </tbody>
</table>`)
			if err != nil {
				log.Fatal(err)
			}

			ps := make([]*Stats, 0)
			for _, s := range modeStats {
				ps = append(ps, s)
			}
			sort.Sort(ByElo(ps))

			err = tmpl.Execute(f, ps)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("wrote %s\n", f.Name())
		}
	}
}

// ByTime sorts Games by time
type ByTime []*Game

func (s ByTime) Len() int {
	return len(s)
}
func (s ByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByTime) Less(i, j int) bool {
	return s[i].Time < s[j].Time
}

// ByElo sorts Stats by Elo
type ByElo []*Stats

func (s ByElo) Len() int {
	return len(s)
}
func (s ByElo) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByElo) Less(i, j int) bool {
	return s[i].Elo > s[j].Elo
}

// ByName sorts Stats by player name
type ByName []*Stats

func (s ByName) Len() int {
	return len(s)
}
func (s ByName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByName) Less(i, j int) bool {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return strings.ToLower(reg.ReplaceAllString(s[i].Name, "")) < strings.ToLower(reg.ReplaceAllString(s[j].Name, ""))
}
