package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")

type StatHolder struct {
	Max, Min, Sum, Count int
}

func split(line []byte) ([]byte, []byte, bool) {
	var index = 0
	for ; index < len(line); index++ {
		if line[index] == ';' {
			break
		}
	}
	if index == len(line) {
		return nil, nil, false
	}
	return line[:index], line[index+1:], true
}

func bytesToNum(bTemp []byte) (int, int) {
	i := 0
	temp := 0
	negative := false

	if bTemp[i] == '-' {
		i++
		negative = true
	}
	temp = int(bTemp[i] - '0')
	i++
	if bTemp[i] != '.' {
		temp = temp*10 + int(bTemp[i]-'0')
		i++
	}
	i++
	temp = temp*10 + int(bTemp[i]-'0')
	if negative {
		temp = -temp
	}
	i += 2
	return temp, i
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(fmt.Sprintf("cannot open file measurements.txt: %s\n", err))
	}
	defer file.Close()

	temps := make(map[string]*StatHolder)
	buf := make([]byte, 1024)
	pos := 0
	for {
		read, err := file.Read(buf[pos:])
		if err != nil && err != io.EOF {
			panic("could not read from file: " + err.Error())
		}
		if pos+read == 0 {
			break
		}
		chunk := buf[:pos+read]
		newline := bytes.LastIndexByte(chunk, '\n')
		if newline < 0 {
			break
		}
		remaining := chunk[newline+1:]
		chunk = chunk[:newline+1]
		for {
			bCity, after, hasSemi := bytes.Cut(chunk, []byte{';'})
			if !hasSemi {
				break
			}
			index := 0
			temp, forward := bytesToNum(after)
			index += forward
			chunk = after[index:]
			city := string(bCity)
			cityStat, found := temps[city]
			if !found {
				temps[city] = &StatHolder{
					Max:   temp,
					Min:   temp,
					Sum:   temp,
					Count: 1,
				}
			} else {
				cityStat.Sum += temp
				cityStat.Count++
				if cityStat.Max < temp {
					cityStat.Max = temp
				}
				if cityStat.Min > temp {
					cityStat.Min = temp
				}
			}
		}

		pos = copy(buf, remaining)
	}
	for city, cityStat := range temps {
		avg := float64(cityStat.Sum) / float64(cityStat.Count) / 10
		fmt.Printf("%s: %.1f/%.1f/%.1f\n", city, float64(cityStat.Min)/10, avg, float64(cityStat.Max)/10)
	}

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up to date stats
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}
}
