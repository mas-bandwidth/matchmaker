/*
	Copyright (c) 2022, Network Next, Inc. All rights reserved.

	This is open source software licensed under the BSD 3-Clause License.

	Redistribution and use in source and binary forms, with or without
	modification, are permitted provided that the following conditions are met:

	1. Redistributions of source code must retain the above copyright notice, this
	   list of conditions and the following disclaimer.

	2. Redistributions in binary form must reproduce the above copyright notice,
	   this list of conditions and the following disclaimer in the documentation
	   and/or other materials provided with the distribution.

	3. Neither the name of the copyright holder nor the names of its
	   contributors may be used to endorse or promote products derived from
	   this software without specific prior written permission.

	THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
	AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
	IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
	DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
	FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
	DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
	SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
	CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
	OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
	OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"math"
	"math/rand"
	"sort"
	"bufio"
	"strings"
	"strconv"
	"text/tabwriter"
	"encoding/binary"
)

const LatencyMapWidth = 360
const LatencyMapHeight = 180
const LatencyMapSize = LatencyMapWidth * LatencyMapHeight
const LatencyMapBytes = LatencyMapSize * 4
const PlayAgainPercent = 70
const MatchLengthSeconds = 198
const IdealTime = 60
const WarmBodyTime = 120
const OneIn = 10
const PlayersPerMatch = 4
const SecondsPerDay = 86400
const MinLatitude = -90
const MaxLatitude = +90
const MinLongitude = -180
const MaxLongitude = +180
const IdealCostThreshold = 50
const IdealCostSpread = 10
const WarmBodyCostSpread = 10

type NewPlayerData struct {
	latitude float64
	longitude float64
}

var newPlayerData [][]NewPlayerData

func percentChance(threshold int) bool {
	return randomInt(0,100) <= threshold
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
	return kilometers / 299792.458 * 1000.0 * 2.0 * ( 3.0 / 2.0 ) // speed of light is 2/3rds in fiber optic cables
}

type Datacenter struct {
	name string
	latitude float64
	longitude float64
	playerCount int
	playerQueue []*ActivePlayer
	averageLatency float64
	averageMatchingTime float64
	latencyMap []float32
}

var datacenters map[uint64]*Datacenter

func initialize() {

	fmt.Printf("initializing...\n")

	// load the players.csv file and parse it

    f, err := os.Open("players.csv")
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

	datacenters[100] = &Datacenter{name: "sanjose", latitude: 37.335480, longitude: -121.893028}
	datacenters[101] = &Datacenter{name: "chicago", latitude: 41.881832, longitude: -87.623177}
	datacenters[102] = &Datacenter{name: "dallas", latitude: 32.779167, longitude: -96.808891}
	datacenters[103] = &Datacenter{name: "miami", latitude: 25.793449, longitude: -80.139198}
	datacenters[104] = &Datacenter{name: "newyork", latitude: 40.730610, longitude: -73.935242}
	datacenters[105] = &Datacenter{name: "washingtondc", latitude: 38.889248, longitude: -77.050636}
	datacenters[106] = &Datacenter{name: "seattle", latitude: 47.608013, longitude: -122.335167}
	datacenters[107] = &Datacenter{name: "atlanta", latitude: 33.753746, longitude: -84.386330}
	datacenters[108] = &Datacenter{name: "montreal", latitude: 45.508888, longitude: -73.561668}
	datacenters[109] = &Datacenter{name: "toronto", latitude: 43.651070, longitude: -79.347015}
	datacenters[110] = &Datacenter{name: "saltlakecity", latitude: 40.758701, longitude: -111.876183}
	datacenters[111] = &Datacenter{name: "losangeles", latitude: 34.052235, longitude: -118.243683}
	datacenters[112] = &Datacenter{name: "ashburn", latitude: 39.0403, longitude: -77.4852}
	datacenters[112] = &Datacenter{name: "phoenix", latitude: 33.448376, longitude: -112.074036}
	datacenters[112] = &Datacenter{name: "denver", latitude: 39.742043, longitude: -104.991531}

	datacenters[200] = &Datacenter{name: "saopaulo", latitude: -23.533773, longitude: -46.625290}
	datacenters[201] = &Datacenter{name: "santiago", latitude: -33.447487, longitude: -70.673676}

	datacenters[300] = &Datacenter{name: "frankfurt", latitude: 50.110924, longitude: 8.682127}
	datacenters[301] = &Datacenter{name: "amsterdam", latitude: 52.377956, longitude: 4.897070}
	datacenters[302] = &Datacenter{name: "london", latitude: 51.5072, longitude: 0.1276}
	datacenters[303] = &Datacenter{name: "luxembourg", latitude: 49.611622, longitude: 6.131935}
	datacenters[304] = &Datacenter{name: "strasbourg", latitude: 48.580002, longitude: 7.750000}
	datacenters[305] = &Datacenter{name: "madrid", latitude: 40.416775, longitude: -3.703790}

	datacenters[400] = &Datacenter{name: "sydney", latitude: -33.865143, longitude: 151.209900}

	datacenters[500] = &Datacenter{name: "johannesburg", latitude: -26.195246, longitude: 28.034088}

	for _,v := range datacenters {
		v.playerQueue = make([]*ActivePlayer, 0, 1024)
	}

	// load latency maps for each datacenter

	for _,v := range datacenters {
		datacenterName := v.name
		filename := fmt.Sprintf("latency_%s.bin", datacenterName)
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
			integerValue := binary.LittleEndian.Uint32(data[index:index+4])
			floatArray[i] = math.Float32frombits(integerValue)
			index += 4
		}
		v.latencyMap = floatArray
	}

    // create active players hash (empty)

	activePlayers = make(map[uint64]*ActivePlayer)

	fmt.Printf("ready!\n")
}

const PlayerState_New = 0
const PlayerState_Ideal = 1
const PlayerState_WarmBody = 2
const PlayerState_Playing = 3
const PlayerState_Bots = 4

type DatacenterCostMapEntry struct {
	index int
	cost float64
}

type DatacenterCostEntry struct {
	datacenterId uint64
	cost float64
}

type ActivePlayer struct {
	playerId uint64
	state int
	latitude float64
	longitude float64
	datacenterCostMap map[uint64]DatacenterCostMapEntry
	datacenterCosts []DatacenterCostEntry
	counter int
	datacenterId uint64
}

var activePlayers map[uint64]*ActivePlayer

func runSimulation() {

	var seconds uint64
	var playerId uint64
	var totalBots uint64

	for {

		i := seconds % SecondsPerDay

		// add new players to the simulation

		for j := range newPlayerData[i] {
			
			if !chance(OneIn) {
				continue
			}

			activePlayer := ActivePlayer{}
			
			activePlayer.playerId = playerId
			activePlayer.latitude = newPlayerData[i][j].latitude
			activePlayer.longitude = newPlayerData[i][j].longitude
			activePlayer.datacenterCostMap = make(map[uint64]DatacenterCostMapEntry)
			activePlayer.datacenterCosts = make([]DatacenterCostEntry, len(datacenters))
			
			index := 0
			for k,v := range datacenters {
				kilometers := haversineDistance(activePlayer.latitude, activePlayer.longitude, v.latitude, v.longitude)
				milliseconds := kilometersToRTT(kilometers)
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
		numWarmBody := 0
		numPlaying := 0

		warmBodies := make(map[uint64]*ActivePlayer)

		for i := range activePlayers {

			if activePlayers[i].state == PlayerState_New {

				numNew++

				if activePlayers[i].datacenterCosts[0].cost <= IdealCostThreshold {

					activePlayers[i].state = PlayerState_Ideal

					costLimit := activePlayers[i].datacenterCosts[0].cost + IdealCostSpread
					
					for j := range activePlayers[i].datacenterCosts {
						datacenterId := activePlayers[i].datacenterCosts[j].datacenterId
						datacenterCost := activePlayers[i].datacenterCosts[j].cost
						if datacenterCost <= costLimit {
							datacenters[datacenterId].playerQueue = append(datacenters[datacenterId].playerQueue, activePlayers[i])
						}
					}

				} else {

					activePlayers[i].state = PlayerState_WarmBody
					activePlayers[i].counter = 0

				}

			} else if activePlayers[i].state == PlayerState_Ideal {

				numIdeal++

				activePlayers[i].counter++

				if activePlayers[i].counter > IdealTime {
					activePlayers[i].state = PlayerState_WarmBody
					activePlayers[i].counter = 0
				}

			} else if activePlayers[i].state == PlayerState_WarmBody {

				numWarmBody++

				activePlayers[i].counter++

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
						activePlayers[i].state = PlayerState_New
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
						activePlayers[i].state = PlayerState_New
						activePlayers[i].counter = 0
						activePlayers[i].datacenterId = 0
					} else {
						delete(activePlayers, i)
					}
				}
			}
		}

		// iterate across all datacenter queues

		for datacenterId, datacenter := range datacenters {

			playerCount := 0
			var matchPlayers [PlayersPerMatch]*ActivePlayer

			for i := range datacenter.playerQueue {

				if datacenter.playerQueue[i].state == PlayerState_Ideal || datacenter.playerQueue[i].state == PlayerState_WarmBody {
					matchPlayers[playerCount] = datacenter.playerQueue[i]
					playerCount++
				} else {
					continue
				}

				if playerCount == PlayersPerMatch {
					for j := 0; j < PlayersPerMatch; j++ {
						datacenter.playerCount++
						for k := range matchPlayers[j].datacenterCosts {
							if matchPlayers[j].datacenterCosts[k].datacenterId == datacenterId {
								datacenter.averageLatency += (matchPlayers[j].datacenterCosts[k].cost - datacenter.averageLatency) * 0.05
								datacenter.averageMatchingTime += (float64(matchPlayers[j].counter) - datacenter.averageMatchingTime) * 0.01
							}
						}
						matchPlayers[j].state = PlayerState_Playing
						matchPlayers[j].datacenterId = datacenterId
						matchPlayers[j].counter = 0
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
			t := float64(warmBody.counter) / WarmBodyTime
			a := warmBody.datacenterCosts[0].cost + WarmBodyCostSpread
			b := warmBody.datacenterCosts[len(warmBody.datacenterCosts)-1].cost - a
			c := a * t*2 * b
			for datacenterId, datacenter := range datacenters {
				datacenterCost := warmBody.datacenterCostMap[datacenterId].cost
				if datacenterCost <= c && len(datacenter.playerQueue) > 0 && len(datacenter.playerQueue) < PlayersPerMatch {
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
		}

		// print status for this iteration

		time := secondsToTime(seconds)

		fmt.Printf("--------------------------------------------------------------------------------------------\n")
		fmt.Printf("%s: %d playing, %d new, %d ideal, %d warmbody, %d bot matches\n", time.Format("2006-01-02 15:04:05"), numPlaying, numNew, numIdeal, numWarmBody, totalBots)
		fmt.Printf("--------------------------------------------------------------------------------------------\n")

		datacenterArray := make([]*Datacenter, len(datacenters))

		index := 0
		for _,v := range datacenters {
			datacenterArray[index] = v
			index++
		}

		sort.SliceStable(datacenterArray[:], func(i, j int) bool {
			return datacenterArray[i].name < datacenterArray[j].name
		})

		sort.SliceStable(datacenterArray[:], func(i, j int) bool {
			return datacenterArray[i].playerCount > datacenterArray[j].playerCount
		})

		w := new(tabwriter.Writer)
	
		w.Init(os.Stdout, 24, 8, 0, '\t', 0)
	
		for i := range datacenterArray {
			fmt.Fprintf(w, "%d\t%s\t%.1fms\t%.1fs\t\n", datacenterArray[i].playerCount, datacenterArray[i].name, datacenterArray[i].averageLatency, datacenterArray[i].averageMatchingTime)
		}

		w.Flush()

		seconds++
	}
}

func main() {

	initialize()

	go runSimulation()

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)
	<-termChan

	fmt.Printf("\nshutting down\n")

	fmt.Printf("shutdown completed\n")
}
