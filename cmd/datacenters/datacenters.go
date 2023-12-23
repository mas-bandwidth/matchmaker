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
)

func main() {

	f, err := os.Open("data/datacenters.csv")
	if err != nil {
		panic(err)
	}

	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		values := strings.Split(scanner.Text(), ",")
		if len(values) != 4 {
			continue
		}
		datacenterId, _ := strconv.Atoi(values[0])
		city := values[1]
		latitude, _ := strconv.ParseFloat(values[2], 64)
		longitude, _ := strconv.ParseFloat(values[3], 64)
		fmt.Printf("%s\n", city)
		_ = datacenterId
		_ = latitude
		_ = longitude
	}
}
