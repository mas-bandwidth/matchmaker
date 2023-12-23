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
	"math"
	"os"
	"bufio"
	"strings"
	"strconv"
	"encoding/binary"
)

const LatencyMapWidth = 360
const LatencyMapHeight = 180
const LatencyMapSize = LatencyMapWidth * LatencyMapHeight
const LatencyMapBytes = LatencyMapSize * 4
const ConservativeFactor = 2.0

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
	return kilometers / 300 * ConservativeFactor
}

func getSample(inputArray []float32, x int, y int) float32 {
	if x < 0 {
		x += LatencyMapWidth
	}
	if x > LatencyMapWidth - 1 {
		x = LatencyMapWidth - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= LatencyMapHeight - 1 {
		y = LatencyMapHeight - 1
	}
	index := x + y*LatencyMapWidth
	return inputArray[index]
}

func fillInHolesFilter(inputArray []float32, outputArray []float32, x int, y int, latitude float64, longitude float64) {
	index := x + y*LatencyMapWidth
	if inputArray[index] >= 1.0 {
		outputArray[index] = inputArray[index]
		return
	}
	latency_sum := float32(0.0)
	latency_count := 0
	latency_minimum := float32(1000.0)
	latency_maximum := float32(0.0)
	for i := -4; i <= +4; i++ {
		for j := -4; j <= +4; j++ {
			sample_x := x + j
			sample_y := y + i
			sample_latitude := latitude - float64(i)
			sample_longitude := longitude + float64(j)
			sample_latency := getSample(inputArray, sample_x, sample_y)
			if sample_latency < 1.0 {
				continue
			}
			distance := haversineDistance(latitude, longitude, sample_latitude, sample_longitude)
			rtt_to_sample := kilometersToRTT(distance)
			sample_latency += float32(rtt_to_sample)
			if sample_latency < latency_minimum {
				latency_minimum = sample_latency
			}
			if sample_latency > latency_maximum {
				latency_maximum = sample_latency
			}
			latency_sum += sample_latency
			latency_count++
		}
	}

	_ = latency_minimum
	_ = latency_maximum
	_ = latency_count
	_ = latency_sum

	/*
	if latency_maximum > 0.0 {
		outputArray[index] = latency_maximum
	}
	*/

	if latency_count > 0 {
		outputArray[index] = latency_sum / float32(latency_count)
	}
	
	/*
	if latency_minimum < 1000.0 {
		outputArray[index] = latency_minimum
	}
	*/
}

func transform(inputFilename string, outputFilename string, datacenterLatitude float64, datacenterLongitude float64) {

	data, err := os.ReadFile(inputFilename)
	if err != nil {
		fmt.Printf("missing binfile: %s\n", inputFilename)
		return
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

	// IMPORTANT: clear "null island" at ~(0,0) lat/long
	index = LatencyMapWidth/2 + LatencyMapHeight/2 * LatencyMapWidth
	floatArray[index-LatencyMapWidth] = 0.0
	floatArray[index-LatencyMapWidth-1] = 0.0
	floatArray[index-LatencyMapWidth+1] = 0.0
	floatArray[index] = 0.0
	floatArray[index-1] = 0.0
	floatArray[index+1] = 0.0
	floatArray[index+LatencyMapWidth-1] = 0.0
	floatArray[index+LatencyMapWidth] = 0.0
	floatArray[index+LatencyMapWidth+1] = 0.0

	// IMPORTANT: Clear highly improbable low latency samples that are far away from the datacenter
	latitude := +90.0
	for y := 0; y < LatencyMapHeight; y++ {
		longitude := -180.0
		for x := 0; x < LatencyMapWidth; x++ {
			index = x + y*LatencyMapWidth
			distance := haversineDistance(latitude, longitude, datacenterLatitude, datacenterLongitude)
			if floatArray[index] < 50 && distance > 1500 {
				floatArray[index] = 0.0
			}
			longitude += 1.0
		}
		latitude -= 1.0
	}

	// Clamp in [0,255]
	for y := 0; y < LatencyMapHeight; y++ {
		for x := 0; x < LatencyMapWidth; x++ {
			index = x + y*LatencyMapWidth
			if floatArray[index] >= 255.0 {
				floatArray[index] = 255.0
			}
			if floatArray[index] < 1.0 {
				floatArray[index] = 0.0
			}
		} 
	}

	// Filter so we fill in holes where we don't have samples, where there are surrounding samples
	outputArray := make([]float32, LatencyMapSize)
	latitude = +90.0
	for y := 0; y < LatencyMapHeight; y++ {
		longitude := -180.0
		for x := 0; x < LatencyMapWidth; x++ {
			fillInHolesFilter(floatArray, outputArray, x, y, latitude, longitude)
			longitude += 1.0
		}
		latitude -= 1.0
	}
	floatArray = outputArray

	data = make([]byte, LatencyMapBytes)
	index = 0
	for i := 0; i < LatencyMapSize; i++ {
		integerValue := math.Float32bits(floatArray[i])
		binary.LittleEndian.PutUint32(data[index:index+4], integerValue)
		index += 4
	}

	os.WriteFile(outputFilename, data, 0666)
}

func main() {

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

	for i := range cities {
		city := cities[i]
		source_filename := fmt.Sprintf("./data/latency_%s.bin", city)
		dest_filename := fmt.Sprintf("latency_%s_transformed.bin", city)
		fmt.Printf("%s\n", dest_filename)
		transform(source_filename, dest_filename, latitudes[i], longitudes[i])
	}
}
