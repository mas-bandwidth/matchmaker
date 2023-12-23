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
	"image"
	"image/png"
    "image/color"
    "math"
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

	f, err := os.Open("data/datacenters.csv")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	filenames := make([]string, 0)

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
		_ = latitude
		_ = longitude
		filenames = append(filenames, fmt.Sprintf("data/latency_%s.bin", city))
	}

	latencyMaps := make([][]float32, 0)

	for i := range filenames {
		filename := filenames[i]
		fmt.Printf("'%s'\n", filename)
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("missing binfile: %s\n", filename)
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
		latencyMaps = append(latencyMaps, floatArray)
	}

	combined := make([]float32, LatencyMapSize)

	for i := range combined {
		value := float32(1000.0)
		for j := range latencyMaps {
			if latencyMaps[j][i] >= 1.0 && latencyMaps[j][i] < value {
				value = latencyMaps[j][i]
			}
		}
		if value <= 200.0 {
			combined[i] = value
		}
	}

	data := make([]byte, LatencyMapBytes)
	index := 0
	for i := 0; i < LatencyMapSize; i++ {
		integerValue := math.Float32bits(combined[i])
		binary.LittleEndian.PutUint32(data[index:index+4], integerValue)
		index += 4
	}

	os.WriteFile("combined.bin", data, 0666)

	// write out as color png

	imageData := image.NewRGBA(image.Rectangle{image.Point{0,0},image.Point{LatencyMapWidth,LatencyMapHeight}})

	for x := 0; x < LatencyMapWidth; x++ {
		for y := 0; y < LatencyMapHeight; y++ {
			index := y*LatencyMapWidth + x
			intensity := combined[index]
			if intensity <= 50 {
		    	c := color.RGBA{uint8(0),uint8(intensity*4),uint8(0),255}
		    	imageData.Set(x,y,c)
			} else if intensity <= 100 {
		    	c := color.RGBA{uint8(intensity*4),uint8(intensity*4*0.647),uint8(0),255}
		    	imageData.Set(x,y,c)
			} else {
		    	c := color.RGBA{uint8(intensity*2),uint8(0),uint8(0),255}
		    	imageData.Set(x,y,c)
			}
		}
	}
	 
	imageFile, _ := os.Create("color.png")

    png.Encode(imageFile, imageData)

	// write out as javascript array for visualization

	const JSArrayBlockSize = 3
	const JSArrayWidth = LatencyMapWidth / JSArrayBlockSize
	const JSArrayHeight = LatencyMapHeight / JSArrayBlockSize
	const JSArraySize = JSArrayWidth * JSArrayHeight

	jsArray := make([]float32, JSArraySize)

	for y := 0; y < JSArrayHeight; y++ {
		for x := 0; x < JSArrayWidth; x++ {
			bx := x * JSArrayBlockSize
			by := y * JSArrayBlockSize
			sum := float32(0.0)
			count := float32(0.0)
			for j := 0; j < JSArrayBlockSize; j++ {
				for i := 0; i < JSArrayBlockSize; i++ {
					index := (by+j) * LatencyMapWidth + (bx+i)
					if combined[index] >= 1.0 {
						sum += combined[index]
						count++
					}
				}
			}
			if count > 0 {
				index := x+y*JSArrayWidth
				jsArray[index] = float32(sum / count) / 255.0
			}
		}
	}

	for i := 0; i < len(jsArray); i++ {
		fmt.Printf("%.5f,", jsArray[i])
	}
	fmt.Printf("\n")

	/*
	const JSArrayBlockSize = 3
	const JSArrayWidth = LatencyMapWidth / JSArrayBlockSize
	const JSArrayHeight = LatencyMapHeight / JSArrayBlockSize
	const JSArraySize = JSArrayWidth * JSArrayHeight

	jsArray := make([]byte, JSArraySize)

	for y := 0; y < JSArrayHeight; y++ {
		for x := 0; x < JSArrayWidth; x++ {
			bx := x * JSArrayBlockSize
			by := y * JSArrayBlockSize
			numGreen := 0
			numOrange := 0
			numRed := 0
			for j := 0; j < JSArrayBlockSize; j++ {
				for i := 0; i < JSArrayBlockSize; i++ {
					index := (by+j) * LatencyMapWidth + (bx+i)
					if combined[index] >= 1.0 {
						if combined[index] <= 50.0 {
							numGreen++
						} else if combined[index] <= 100.0 {
							numOrange++
						} else {
							numRed++
						}
					}
				}
			}
			index := x+y*JSArrayWidth
			if numRed > 0 {
				jsArray[index] = 3
			} else if numOrange > 0 {
				jsArray[index] = 2
			} else if numGreen > 0 {
				jsArray[index] = 1
			}
		}
	}

	for i := 0; i < len(jsArray); i++ {
		fmt.Printf("%d,", jsArray[i])
	}
	fmt.Printf("\n")
	*/
}
