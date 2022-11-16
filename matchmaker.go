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
)

const SecondsPerDay = 86400
const MinLatitude = -90
const MaxLatitude = +90
const MinLongitude = -180
const MaxLongitude = +180
const IdealCostThreshold = 25
const IdealCostSpread = 10

type NewPlayerData struct {
	latitude float64
	longitude float64
}

var newPlayerData [][]NewPlayerData

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
	return kilometers / 299792.458 * 1000.0 * ( 3.0 / 2.0 ) // ~ 2/3rds speed of light in fiber optic cables
}

type Datacenter struct {
	name string
	latitude float64
	longitude float64
}

var datacenters map[uint64]*Datacenter

func initialize() {

	fmt.Printf("initializing...\n")

	newPlayerData = make([][]NewPlayerData, SecondsPerDay)
	for i := 0; i < SecondsPerDay; i++ {
		newPlayers := randomInt(0,5)
		newPlayerData[i] = make([]NewPlayerData, newPlayers)
		for j := 0; j < newPlayers; j++ {
			newPlayerData[i][j].latitude = float64(randomInt(MinLatitude, MaxLatitude))
			newPlayerData[i][j].longitude = float64(randomInt(MinLongitude, MaxLongitude))
		}
	}

	activePlayers = make(map[uint64]*ActivePlayer)

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

	datacenters[200] = &Datacenter{name: "saopaulo", latitude: -23.533773, longitude: -46.625290}
	datacenters[201] = &Datacenter{name: "santiago", latitude: -33.447487, longitude: -70.673676}

	datacenters[300] = &Datacenter{name: "frankfurt", latitude: 50.110924, longitude: 8.682127}
	datacenters[301] = &Datacenter{name: "amsterdam", latitude: 52.377956, longitude: 4.897070}
	datacenters[302] = &Datacenter{name: "london", latitude: 51.5072, longitude: 0.1276}
	datacenters[303] = &Datacenter{name: "luxembourg", latitude: 49.611622, longitude: 6.131935}
	datacenters[304] = &Datacenter{name: "strasbourg", latitude: 48.580002, longitude: 7.750000}
	datacenters[305] = &Datacenter{name: "madrid", latitude: 40.416775, longitude: -3.703790}

	datacenters[400] = &Datacenter{name: "sydney", latitude: -33.865143, longitude: 151.209900}

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
	state int
	latitude float64
	longitude float64
	datacenterCostMap map[uint64]DatacenterCostMapEntry
	datacenterCosts []DatacenterCostEntry
}

var activePlayers map[uint64]*ActivePlayer

func runSimulation() {

	var seconds uint64
	var playerId uint64

	for {

		i := seconds % SecondsPerDay

		// add new players to the simulation

		for j := range newPlayerData[i] {
			
			activePlayer := ActivePlayer{}
			
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
		// numPlaying := 0
		// numBots := 0

		for i := range activePlayers {

			if activePlayers[i].state == PlayerState_New {

				numNew++

				if activePlayers[i].datacenterCosts[0].cost > IdealCostThreshold {
					activePlayers[i].state = PlayerState_WarmBody
				} else {
					activePlayers[i].state = PlayerState_Ideal
					// todo: add player to set of datacenter queues within spread of best datacenter
				}

			} else if activePlayers[i].state == PlayerState_Ideal {

				numIdeal++

			} else if activePlayers[i].state == PlayerState_WarmBody {

				numWarmBody++

			}
		}

		// todo: iterate across all datacenter queues (random order)

		// ...

		// print status for this iteration

		time := secondsToTime(seconds)

		fmt.Printf("%s: %d new, %d ideal, %d warmbody\n", time.Format("2006-01-02 15:04:05"), numNew, numIdeal, numWarmBody)

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
