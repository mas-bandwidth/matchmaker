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
	"math"
	"os"
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

func main() {

	fmt.Printf("transformers. more than meets the eye\n")

	inputFilename := "input.bin"
	outputFilename := "output.bin"

	// temporary: sanjose
	datacenterLatitude := 37.335480
	datacenterLongitude := -121.893028

	data, err := os.ReadFile(inputFilename)
	if err != nil {
		panic(inputFilename)
	}

	if len(data) != LatencyMapBytes {
		panic(fmt.Sprintf("latency map %s is invalid size (%d bytes)", inputFilename, len(data)))
	}
	
	fmt.Printf("loaded %s\n", inputFilename)

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

	// For empty squares, look in a square around the sample, and see if any neighbour points are set.
	// Use the cheapest neighbour square latency as the empty sample latency, so we fill in holes.
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

	/*
	// Flip around black values to white to help the filter
	for y := 0; y < LatencyMapHeight; y++ {
		for x := 0; x < LatencyMapWidth; x++ {
			index = x + y*LatencyMapWidth
			if floatArray[index] < 1.0 {
				floatArray[index] = 255.0
			}
		} 
	}
	*/

	data = make([]byte, LatencyMapBytes)
	index = 0
	for i := 0; i < LatencyMapSize; i++ {
		integerValue := math.Float32bits(floatArray[i])
		binary.LittleEndian.PutUint32(data[index:index+4], integerValue)
		index += 4
	}

	os.WriteFile(outputFilename, data, 0666)

	fmt.Printf("wrote %s\n", outputFilename)
}
