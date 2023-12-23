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
	"os"
	"regexp"
	"encoding/binary"
	"math"
	/*
	"bufio"
	"strings"
	"strconv"
	*/
)

const LatencyMapWidth = 360
const LatencyMapHeight = 180
const LatencyMapSize = LatencyMapWidth * LatencyMapHeight
const LatencyMapBytes = LatencyMapSize * 4

const SecondsPerDay = 86400

const MinLatitude = -90
const MaxLatitude = +90
const MinLongitude = -180
const MaxLongitude = +180

func main() {

	// convert args into set of sums and counts filenames

	args := os.Args[1:]

	sums := make([]string, 0)
	counts := make([]string, 0)

	for i := range args {
		found, err := regexp.MatchString("^.*_counts.bin$", args[i])
		if err != nil {
			panic(err)
		}
		if found {
			counts = append(counts, args[i])
			sums = append(sums, args[i][:len(args[i])-11] + "_sums.bin")
		}
	}

	// load sums and counts files and add to totals as float64

	sums_total := make([]float64, LatencyMapSize)
	counts_total := make([]float64, LatencyMapSize)

	for i := range sums {

		fmt.Printf("%s | %s\n", sums[i], counts[i])

		sums_data, err := os.ReadFile(sums[i])
		if err != nil {
			fmt.Printf("missing sums file: %s\n", sums[i])
			continue
		}
		if len(sums_data) != LatencyMapBytes * 2 {
			panic(fmt.Sprintf("sums file %s is invalid size (%d bytes)", sums[i], len(sums_data)))
		}

		counts_data, err := os.ReadFile(counts[i])
		if err != nil {
			fmt.Printf("missing counts file: %s\n", counts[i])
			continue
		}
		if len(counts_data) != LatencyMapBytes * 2 {
			panic(fmt.Sprintf("counts file %s is invalid size (%d bytes)", counts[i], len(counts_data)))
		}

		sums_float64 := make([]float64, LatencyMapSize)
		counts_float64 := make([]float64, LatencyMapSize)

		index := 0
		for i := 0; i < LatencyMapSize; i++ {
			sums_integerValue := binary.LittleEndian.Uint64(sums_data[index : index+8])
			sums_float64[i] = math.Float64frombits(sums_integerValue)
			counts_integerValue := binary.LittleEndian.Uint64(counts_data[index : index+8])
			counts_float64[i] = math.Float64frombits(counts_integerValue)
			index += 8
		}

		for i := 0; i < LatencyMapSize; i++ {
			sums_total[i] += sums_float64[i]
			counts_total[i] += counts_float64[i]
		}
	}

	// convert the sums and totals into a latency map (float32)

	latencyMap := make([]float32, LatencyMapSize)

	for i := 0; i < LatencyMapSize; i++ {
		if counts_total[i] > 0.0 {
			latencyMap[i] = float32(sums_total[i]/counts_total[i])
		}
	}

	// write the latency map to output.bin

	data := make([]byte, LatencyMapBytes)
	index := 0
	for i := 0; i < LatencyMapSize; i++ {
		integerValue := math.Float32bits(latencyMap[i])
		binary.LittleEndian.PutUint32(data[index:index+4], integerValue)
		index += 4
	}

	os.WriteFile("output.bin", data, 0666)
}
