/*
	Matchmaker

	Copyright (c) 2023 - 2024, Mas Bandwidth LLC. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"sort"
	"sync"
	"strconv"
	"strings"
	"syscall"
	"time"
	"net/http"
	"encoding/binary"

	"github.com/gorilla/mux"
)

const PlayersPerMatch = 4
const DelaySeconds = 50
const PlayAgainPercent = 96
const MatchLengthSeconds = 180
const IdealTime = 60
const ExpandTime = 60
const WarmBodyTime = 60
const SpeedOfLightFactor = 2

const IdealCostThreshold = 300
const IdealCostSpread = 10
const ExpandMaxCost = 200
const ExpandCostSpread = 10
const WarmBodyCostThreshold = 255

const SampleDays = 20           // the number of days worth of samples contained in players.csv

const LatencyMapWidth = 360
const LatencyMapHeight = 180
const LatencyMapSize = LatencyMapWidth * LatencyMapHeight
const LatencyMapBytes = LatencyMapSize * 4

const SecondsPerDay = 86400

const MinLatitude = -90
const MaxLatitude = +90
const MinLongitude = -180
const MaxLongitude = +180

var mapDataMutex sync.RWMutex
var mapData      []byte

type NewPlayerData struct {
	latitude  float64
	longitude float64
}

var newPlayerData [][]NewPlayerData

func percentChance(threshold int) bool {
	return randomInt(0, 100) <= threshold
}

func chance(n int) bool {
	if rand.Intn(n) == 0 {
		return true
	}
	return false
}

func randomInt(min int, max int) int {
	difference := max - min
	value := rand.Intn(difference + 1)
	return value + min
}

func secondsToTime(second uint64) time.Time {
	startTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	return startTime.Add(time.Duration(second) * time.Second)
}

func haversineDistance(lat1 float64, long1 float64, lat2 float64, long2 float64) float64 {
	lat1 *= math.Pi / 180
	lat2 *= math.Pi / 180
	long1 *= math.Pi / 180
	long2 *= math.Pi / 180
	delta_lat := lat2 - lat1
	delta_long := long2 - long1
	lat_sine := math.Sin(delta_lat / 2)
	long_sine := math.Sin(delta_long / 2)
	a := lat_sine*lat_sine + math.Cos(lat1)*math.Cos(lat2)*long_sine*long_sine
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	r := 6371.0
	d := r * c
	return d // kilometers
}

func kilometersToRTT(kilometers float64) float64 {
	return kilometers / 299792.458 * 1000.0 * 2.0 * (3.0 / 2.0) // speed of light is 2/3rds in fiber optic cables
}

func datacenterRTT(datacenter *Datacenter, playerLatitude float64, playerLongitude float64) float64 {
	lat := playerLatitude
	long := playerLatitude
	if lat < MinLatitude {
		lat = MinLatitude
	}
	if lat > MaxLatitude {
		lat = MaxLatitude
	}
	if long < 0 {
		long = MaxLongitude + long
	}
	long = math.Mod(long, LatencyMapWidth)
	x := int(math.Floor(long)) + MaxLongitude
	y := MaxLatitude - int(math.Floor(lat))
	if x >= LatencyMapWidth {
		x = LatencyMapWidth - 1
	}
	if y >= LatencyMapHeight {
		y = LatencyMapHeight - 1
	}
	index := x + y*LatencyMapWidth
	if datacenter.latencyMap != nil && datacenter.latencyMap[index] > 0.0 {
		return float64(datacenter.latencyMap[index])
	} else {
		kilometers := haversineDistance(playerLatitude, playerLongitude, datacenter.latitude, datacenter.longitude)
		return kilometersToRTT(kilometers) * SpeedOfLightFactor
	}
}

type Datacenter struct {
	name                string
	latitude            float64
	longitude           float64
	playerCount         int
	playerQueue         []*ActivePlayer
	averageLatency      float64
	averageMatchingTime float64
	latencyMap          []float32
}

var datacenters map[uint64]*Datacenter

var matchesFile *os.File
var statsFile *os.File

func initialize() {

	fmt.Printf("initializing...\n")

	// load the players.csv file and parse it

	f, err := os.Open("new/players.csv")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	newPlayerData = make([][]NewPlayerData, SecondsPerDay)

	for scanner.Scan() {
		values := strings.Split(scanner.Text(), ",")
		if len(values) != 3 {
			continue
		}
		time := values[0]
		latitude, _ := strconv.ParseFloat(values[1], 64)
		longitude, _ := strconv.ParseFloat(values[2], 64)
		time_values := strings.Split(time, ":")
		time_hours, _ := strconv.Atoi(time_values[0])
		time_minutes, _ := strconv.Atoi(time_values[1])
		time_seconds, _ := strconv.Atoi(time_values[2])
		seconds := uint64(0)
		seconds += uint64(time_seconds) + uint64(time_minutes)*60 + uint64(time_hours)*60*60
		newPlayerData[seconds] = append(newPlayerData[seconds], NewPlayerData{latitude: latitude, longitude: longitude})
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	// initialize datacenters for the simulation

	datacenters = make(map[uint64]*Datacenter)

	f, err = os.Open("new/datacenters.csv")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	scanner = bufio.NewScanner(f)

	for scanner.Scan() {
		values := strings.Split(scanner.Text(), ",")
		if len(values) != 4 {
			continue
		}
		datacenterId, _ := strconv.Atoi(values[0])
		city := values[1]
		latitude, _ := strconv.ParseFloat(values[2], 64)
		longitude, _ := strconv.ParseFloat(values[3], 64)
		datacenters[uint64(datacenterId)] = &Datacenter{name: city, latitude: latitude, longitude: longitude}
	}

	for _, v := range datacenters {
		v.playerQueue = make([]*ActivePlayer, 0, 1024)
	}

	// load latency maps for each datacenter

	for _, v := range datacenters {
		datacenterName := v.name
		filename := fmt.Sprintf("new/latency_%s.bin", datacenterName)
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}
		if len(data) != LatencyMapBytes {
			panic(fmt.Sprintf("latency map %s is invalid size (%d bytes)", filename, len(data)))
		}
		fmt.Printf("loaded %s\n", filename)
		index := 0
		floatArray := make([]float32, LatencyMapSize)
		for i := 0; i < LatencyMapSize; i++ {
			integerValue := binary.LittleEndian.Uint32(data[index : index+4])
			floatArray[i] = math.Float32frombits(integerValue)
			index += 4
		}
		v.latencyMap = floatArray
	}

	// create active players hash (empty)

	activePlayers = make(map[uint64]*ActivePlayer)

	fmt.Printf("ready!\n")

	// create output files

	matchesFile, err = os.Create("matches.csv")
	if err != nil {
		panic(err)
	}

	statsFile, err = os.Create("stats.csv")
	if err != nil {
		panic(err)
	}
}

const PlayerState_New = 0
const PlayerState_Ideal = 1
const PlayerState_Expand = 2
const PlayerState_WarmBody = 3
const PlayerState_Playing = 4
const PlayerState_Bots = 5
const PlayerState_Delay = 6

type DatacenterCostMapEntry struct {
	index int
	cost  float64
}

type DatacenterCostEntry struct {
	datacenterId uint64
	cost         float64
}

type ActivePlayer struct {
	playerId          uint64
	state             int
	latitude          float64
	longitude         float64
	datacenterCostMap map[uint64]DatacenterCostMapEntry
	datacenterCosts   []DatacenterCostEntry
	counter           int
	matchingTime      float64
	datacenterId      uint64
	latency           float64
}

var activePlayers map[uint64]*ActivePlayer

func runSimulation() {

	var seconds uint64
	var playerId uint64
	var totalBots uint64

	const MapWidth = 120
	const MapHeight = 64
	const MapSize = MapWidth * MapHeight

	countData := make([]float64, MapSize)

	for {

		i := seconds % SecondsPerDay

		// add new players to the simulation

		for j := range newPlayerData[i] {

			if !chance(SampleDays) {
				continue
			}

			activePlayer := ActivePlayer{}

			activePlayer.playerId = playerId
			activePlayer.latitude = newPlayerData[i][j].latitude
			activePlayer.longitude = newPlayerData[i][j].longitude
			activePlayer.datacenterCostMap = make(map[uint64]DatacenterCostMapEntry)
			activePlayer.datacenterCosts = make([]DatacenterCostEntry, len(datacenters))

			index := 0
			for k, v := range datacenters {
				milliseconds := datacenterRTT(v, activePlayer.latitude, activePlayer.longitude)
				activePlayer.datacenterCostMap[k] = DatacenterCostMapEntry{cost: milliseconds, index: index}
				activePlayer.datacenterCosts[index].datacenterId = k
				activePlayer.datacenterCosts[index].cost = milliseconds
				index++
			}

			sort.SliceStable(activePlayer.datacenterCosts[:], func(i, j int) bool {
				return activePlayer.datacenterCosts[i].cost < activePlayer.datacenterCosts[j].cost
			})

			activePlayers[playerId] = &activePlayer

			playerId++
		}

		// iterate across all active players

		numNew := 0
		numIdeal := 0
		numExpand := 0
		numWarmBody := 0
		numPlaying := 0
		numDelay := 0

		warmBodies := make(map[uint64]*ActivePlayer)

		for i := range activePlayers {

			if activePlayers[i].state == PlayerState_New {

				numNew++

				cost := activePlayers[i].datacenterCosts[0].cost

				activePlayers[i].counter = 0
				activePlayers[i].matchingTime = 0.0

				if cost <= IdealCostThreshold {

					activePlayers[i].state = PlayerState_Ideal

					costLimit := activePlayers[i].datacenterCosts[0].cost + IdealCostSpread

					for j := range activePlayers[i].datacenterCosts {
						datacenterId := activePlayers[i].datacenterCosts[j].datacenterId
						datacenterCost := activePlayers[i].datacenterCosts[j].cost
						if datacenterCost <= costLimit {
							datacenters[datacenterId].playerQueue = append(datacenters[datacenterId].playerQueue, activePlayers[i])
						}
					}

				} else if cost < WarmBodyCostThreshold {

					activePlayers[i].state = PlayerState_Expand

				} else {

					activePlayers[i].state = PlayerState_WarmBody

				}

			} else if activePlayers[i].state == PlayerState_Ideal {

				numIdeal++

				activePlayers[i].counter++
				activePlayers[i].matchingTime += 1.0

				if activePlayers[i].counter > IdealTime {
					activePlayers[i].state = PlayerState_WarmBody
					activePlayers[i].counter = 0
				}

			} else if activePlayers[i].state == PlayerState_Expand {

				numExpand++

				activePlayers[i].counter++
				activePlayers[i].matchingTime += 1.0

				t := float64(activePlayers[i].counter) / ExpandTime

				expandStartCost := activePlayers[i].datacenterCosts[0].cost + ExpandCostSpread

				cost := expandStartCost + t*(ExpandMaxCost-expandStartCost)

				for datacenterId, datacenter := range datacenters {
					datacenterCost := activePlayers[i].datacenterCostMap[datacenterId].cost
					if datacenterCost <= cost {
						found := false
						for j := range datacenter.playerQueue {
							if datacenter.playerQueue[j].playerId == activePlayers[i].playerId {
								found = true
								break
							}
						}
						if !found {
							datacenter.playerQueue = append(datacenter.playerQueue, activePlayers[i])
						}
					}
				}

				if activePlayers[i].counter > ExpandTime {
					activePlayers[i].state = PlayerState_WarmBody
					activePlayers[i].counter = 0
				}

			} else if activePlayers[i].state == PlayerState_WarmBody {

				numWarmBody++

				activePlayers[i].counter++
				activePlayers[i].matchingTime += 1.0

				if activePlayers[i].counter > WarmBodyTime {
					activePlayers[i].state = PlayerState_Bots
					activePlayers[i].counter = MatchLengthSeconds
					totalBots++
				} else {
					warmBodies[i] = activePlayers[i]
				}

			} else if activePlayers[i].state == PlayerState_Playing {

				numPlaying++

				activePlayers[i].counter++

				if activePlayers[i].counter > MatchLengthSeconds {
					datacenters[activePlayers[i].datacenterId].playerCount--
					if percentChance(PlayAgainPercent) {
						activePlayers[i].state = PlayerState_Delay
						activePlayers[i].counter = 0
						activePlayers[i].datacenterId = 0
					} else {
						delete(activePlayers, i)
					}
				}

			} else if activePlayers[i].state == PlayerState_Bots {

				activePlayers[i].counter++

				if activePlayers[i].counter > MatchLengthSeconds {
					if percentChance(PlayAgainPercent) {
						activePlayers[i].state = PlayerState_Delay
						activePlayers[i].counter = 0
						activePlayers[i].datacenterId = 0
					} else {
						delete(activePlayers, i)
					}
				}

			} else if activePlayers[i].state == PlayerState_Delay {

				numDelay++

				activePlayers[i].counter++

				if activePlayers[i].counter > DelaySeconds {
					activePlayers[i].state = PlayerState_New
					activePlayers[i].counter = 0
					activePlayers[i].datacenterId = 0
				}
			}
		}

		// iterate across all datacenter queues

		for datacenterId, datacenter := range datacenters {

			playerCount := 0
			var matchPlayers [PlayersPerMatch]*ActivePlayer

			for i := range datacenter.playerQueue {

				if datacenter.playerQueue[i].state == PlayerState_Ideal || datacenter.playerQueue[i].state == PlayerState_Expand || datacenter.playerQueue[i].state == PlayerState_WarmBody {
					matchPlayers[playerCount] = datacenter.playerQueue[i]
					playerCount++
				} else {
					continue
				}

				if playerCount == PlayersPerMatch {
					for j := 0; j < PlayersPerMatch; j++ {
						datacenter.playerCount++
						latency := 0.0
						for k := range matchPlayers[j].datacenterCosts {
							if matchPlayers[j].datacenterCosts[k].datacenterId == datacenterId {
								latency = matchPlayers[j].datacenterCosts[k].cost
								break
							}
						}
						datacenter.averageLatency += (latency - datacenter.averageLatency) * 0.05
						datacenter.averageMatchingTime += (matchPlayers[j].matchingTime - datacenter.averageMatchingTime) * 0.01
						matchPlayers[j].state = PlayerState_Playing
						matchPlayers[j].datacenterId = datacenterId
						matchPlayers[j].latency = latency
						matchPlayers[j].counter = 0
						// fmt.Fprintf(matchesFile, "%d,%.1f,%.1f,%s,%.1f,%.1f\n", seconds, matchPlayers[j].latitude, matchPlayers[j].longitude, datacenter.name, latency, matchPlayers[j].matchingTime)
					}
					playerCount = 0
				}

			}

			newPlayerQueue := make([]*ActivePlayer, 0, 1024)

			for i := range datacenter.playerQueue {
				if datacenter.playerQueue[i].state == PlayerState_Ideal {
					newPlayerQueue = append(newPlayerQueue, datacenter.playerQueue[i])
				}
			}

			datacenter.playerQueue = newPlayerQueue
		}

		// feed warm bodies back into datacenter queues to fill matches

		for _, warmBody := range warmBodies {
			for _, datacenter := range datacenters {
				found := false
				for i := range datacenter.playerQueue {
					if datacenter.playerQueue[i].playerId == warmBody.playerId {
						found = true
						break
					}
				}
				if !found {
					datacenter.playerQueue = append(datacenter.playerQueue, warmBody)
				}
			}
		}

		// write stats

		// time := secondsToTime(seconds)

		// fmt.Printf("%s: %10d playing %10d delay %5d new %5d ideal %5d expand %4d warmbody %4d bot matches\n", time.Format("2006-01-02 15:04:05"), numPlaying, numDelay, numNew, numIdeal, numExpand, numWarmBody, totalBots)

		fmt.Printf("%s\n", secondsToTime(seconds))

		/*
		datacenterArray := make([]*Datacenter, len(datacenters))

		index := 0
		for _, v := range datacenters {
			datacenterArray[index] = v
			index++
		}

		sort.SliceStable(datacenterArray[:], func(i, j int) bool {
			return datacenterArray[i].name < datacenterArray[j].name
		})

		sort.SliceStable(datacenterArray[:], func(i, j int) bool {
			return datacenterArray[i].playerCount > datacenterArray[j].playerCount
		})

		for i := range datacenterArray {
			fmt.Fprintf(statsFile, "%d,%s,%d,%.1f,%.1f\n", seconds, datacenterArray[i].name, datacenterArray[i].playerCount, datacenterArray[i].averageLatency, datacenterArray[i].averageMatchingTime)
		}
		*/

		seconds++

		// update map data

		for i := range activePlayers {
			if activePlayers[i].state != PlayerState_Playing {
				continue
			}
			ix := int( ( activePlayers[i].longitude + MaxLongitude ) / 3.0 )
			if ix < 0 {
				ix = 0
			} else if ix >= MapWidth {
				ix = MapWidth - 1
			}
			iy := int( ( MaxLatitude - activePlayers[i].latitude ) / 3.0 )
			if iy < 0 {
				iy = 0
			} else if iy >= MapHeight {
				iy = MapHeight - 1
			}
			index := ix + iy*MapWidth
			countData[index]++
		}

		if ( seconds % 10) == 0 {

			data := make([]uint8, MapSize*4)
			for i := 0; i < MapSize; i++ {
				intData := uint32(countData[i])
				binary.LittleEndian.PutUint32(data[i*4:], intData)
			}

			mapDataMutex.Lock()
			mapData = data
			mapDataMutex.Unlock()

			countData = make([]float64, MapSize)
		}
	}
}

func shutdown() {
	matchesFile.Close()
	statsFile.Close()
}

func main() {

	go func() {
		var router mux.Router
		router.HandleFunc("/data", dataHandler).Methods("GET")
		router.HandleFunc("/", serveFile("index.html")).Methods("GET")
		router.HandleFunc("/map.js", serveFile("map.js")).Methods("GET")
		router.HandleFunc("/styles.css", serveFile("styles.css")).Methods("GET")
		fmt.Printf("starting web server\n")
		err := http.ListenAndServe("127.0.0.1:8000", &router)
		if err != nil {
			fmt.Printf("error starting http server: %v\n", err)
			os.Exit(1)
		}
	}()

	initialize()

	go runSimulation()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan

	fmt.Printf("\nshutting down\n")

	shutdown()

	fmt.Printf("shutdown completed\n")
}

func serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) { fmt.Printf("serve %s\n", filename); http.ServeFile(w, r, fmt.Sprintf("map/%s", filename)) }
}

func dataHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/octet-stream")
	mapDataMutex.RLock()
	data := mapData
	mapDataMutex.RUnlock()
	w.Write(data)
}
