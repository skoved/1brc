package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type StatHolder struct {
	Max, Min, Sum float64
	Count         int
}

func main() {
	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(fmt.Sprintf("cannot open file measurements.txt: %s\n", err))
	}
	defer file.Close()

	temps := make(map[string]*StatHolder)
	fReader := bufio.NewScanner(file)
	for fReader.Scan() {
		line := fReader.Text()
		city, temp, found := strings.Cut(line, ";")
		if !found {
			fmt.Printf("did not find ; in %s\n", line)
			continue
		}
		nTemp, err := strconv.ParseFloat(temp, 64)
		if err != nil {
			fmt.Printf("could not convert temp %s to float64: %s\n", temp, err)
			continue
		}
		cityStat, exists := temps[city]
		if !exists {
			temps[city] = &StatHolder{
				Max:   nTemp,
				Min:   nTemp,
				Sum:   nTemp,
				Count: 1,
			}
		} else {
			cityStat.Sum += nTemp
			cityStat.Count++
			if cityStat.Max < nTemp {
				cityStat.Max = nTemp
			}
			if cityStat.Min > nTemp {
				cityStat.Min = nTemp
			}
		}
	}
	for city, cityStat := range temps {
		avg := cityStat.Sum / float64(cityStat.Count)
		fmt.Printf("%s: %.1f/%.1f/%.1f\n", city, cityStat.Min, avg, cityStat.Max)
	}
}
