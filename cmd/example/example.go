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
	"fmt"
	"bufio"
	"os"
	"strings"
	"strconv"
	"encoding/binary"
	"math"
)

const SpeedOfLightFactor = 2.0

const LatencyMapWidth = 360
const LatencyMapHeight = 180
const LatencyMapSize = LatencyMapWidth * LatencyMapHeight
const LatencyMapBytes = LatencyMapSize * 4

const MinLatitude = -90
const MaxLatitude = +90
const MinLongitude = -180
const MaxLongitude = +180

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

func datacenterRTT(latencyMap []float32, datacenterLatitude float64, datacenterLongitude float64, playerLatitude float64, playerLongitude float64) float64 {
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
	index := x + y*LatencyMapWidth
	if latencyMap[index] > 0.0 {
		return float64(latencyMap[index])
	} else {
		kilometers := haversineDistance(playerLatitude, playerLongitude, datacenterLatitude, datacenterLongitude)
		return kilometersToRTT(kilometers) * SpeedOfLightFactor
	}
}

func main() {

	// load datacenters

	f, err := os.Open("data/datacenters.csv")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	cities := make([]string, 0)
	latitudes := make([]float64, 0)
	longitudes := make([]float64, 0)

	for scanner.Scan() {
		values := strings.Split(scanner.Text(), ",")
		if len(values) != 4 {
			continue
		}
		datacenterId, _ := strconv.Atoi(values[0])
		city := values[1]
		latitude, _ := strconv.ParseFloat(values[2], 64)
		longitude, _ := strconv.ParseFloat(values[3], 64)
		_ = datacenterId
		cities = append(cities, city)
		latitudes = append(latitudes, latitude)
		longitudes = append(longitudes, longitude)
	}

	// load latency maps

	latencyMaps := make([][]float32, 0)

	for i := range cities {

		inputFilename := fmt.Sprintf("./data/latency_%s.bin", cities[i])

		data, err := os.ReadFile(inputFilename)
		if err != nil {
			fmt.Printf("missing binfile: %s\n", inputFilename)
			latencyMaps = append(latencyMaps, make([]float32, LatencyMapSize)) // empty file
			continue
		}

		if len(data) != LatencyMapBytes {
			panic(fmt.Sprintf("latency map %s is invalid size (%d bytes)", inputFilename, len(data)))
		}
		
		index := 0
		floatArray := make([]float32, LatencyMapSize)
		for i := 0; i < LatencyMapSize; i++ {
			integerValue := binary.LittleEndian.Uint32(data[index : index+4])
			floatArray[i] = math.Float32frombits(integerValue)
			index += 4
		}

		latencyMaps = append(latencyMaps, floatArray)
	}

	// print latencies between all datacenters

	for i := range cities {
		fmt.Printf("-------------------------------\n")		
		for j := range cities {
			rtt := datacenterRTT(latencyMaps[i], latitudes[i], longitudes[i], latitudes[j], longitudes[j])
			fmt.Printf("%5.1fms: %s <-> %s\n", rtt, cities[i], cities[j])
		}
	}
	fmt.Printf("-------------------------------\n")		
}
