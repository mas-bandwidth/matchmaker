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

	// Clamp in [0,255]. Also, flip around black values to white so we can see bad values in the image
	for y := 0; y < LatencyMapHeight; y++ {
		for x := 0; x < LatencyMapWidth; x++ {
			index = x + y*LatencyMapWidth
			if floatArray[index] >= 255.0 {
				floatArray[index] = 255.0
			}
			if floatArray[index] < 1.0 {
				floatArray[index] = 255.0
			}
		} 
	}

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
