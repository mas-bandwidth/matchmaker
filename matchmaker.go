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
	"math/rand"
)

const SecondsPerDay = 86400
const MinLatitude = -90
const MaxLatitude = +90
const MinLongitude = -180
const MaxLongitude = +180

type NewPlayerData struct {
	latitude float32
	longitude float32
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

const PlayerState_Matchmaking = 0
const PlayerState_Playing = 1

type ActivePlayer struct {
	state int
	latitude float32
	longitude float32
}

var activePlayers map[uint64]*ActivePlayer

func runSimulation() {
	var seconds uint64
	var playerId uint64
	for {
		time := secondsToTime(seconds)
		i := seconds % SecondsPerDay
		for j := range newPlayerData[i] {
			activePlayer := ActivePlayer{}
			activePlayer.latitude = newPlayerData[i][j].latitude
			activePlayer.longitude = newPlayerData[i][j].longitude
			activePlayers[playerId] = &activePlayer
			playerId++
		}
		fmt.Printf("%s: %d players\n", time.Format("2006-01-02 15:04:05"), len(activePlayers))
		seconds++
	}
}

func initialize() {
	fmt.Printf("initializing...\n")
	newPlayerData = make([][]NewPlayerData, SecondsPerDay)
	for i := 0; i < SecondsPerDay; i++ {
		newPlayers := randomInt(0,5)
		newPlayerData[i] = make([]NewPlayerData, newPlayers)
		for j := 0; j < newPlayers; j++ {
			newPlayerData[i][j].latitude = float32(randomInt(MinLatitude, MaxLatitude))
			newPlayerData[i][j].longitude = float32(randomInt(MinLongitude, MaxLongitude))
		}
	}
	activePlayers = make(map[uint64]*ActivePlayer)
	fmt.Printf("ready!\n")
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
