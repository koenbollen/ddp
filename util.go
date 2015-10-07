package main

import (
	"math"
	"os"
	"strconv"
	"strings"
)

var byteUnits map[string]int64
var magnitudes = map[string]float64{"k": 1, "m": 2, "g": 3, "t": 4, "p": 6, "e": 7}

func init() {
	byteUnits = map[string]int64{}
	for magnitude, factor := range magnitudes {
		byteUnits[magnitude] = int64(math.Pow(1024, factor))
		byteUnits[magnitude+"b"] = int64(math.Pow(1000, factor))
	}
}

// GuessTargetSize will return a byte size based on commandline arguments
// intended for the `dd` command.
//
// TODO: This function is marked for rewrite.
func GuessTargetSize(args []string) int64 {
	var ifile string
	var bs, count, skip int64
	for _, arg := range args[1:] {
		parts := strings.Split(arg, "=")
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			switch key {
			case "if":
				ifile = value
			case "bs":
				bs = ParseByteUnits(value)
			case "count":
				count = ParseByteUnits(value)
			case "skip":
				skip = ParseByteUnits(value)
			case "iseek":
				skip = ParseByteUnits(value)
			}
		}
	}
	filesize := int64(0)
	stat, _ := os.Stat(ifile)
	if stat != nil {
		skipbytes := skip
		if bs != 0 {
			skipbytes = skip * bs
		}
		filesize = stat.Size() - skipbytes
	}

	size := bs * count
	if size == 0 || (filesize != 0 && filesize < size) {
		return filesize
	}
	return size
}

// ParseByteUnits will return a humanreadable bytesize as int.
func ParseByteUnits(in string) int64 {
	in = strings.ToLower(in)
	mul := int64(1)
	for unit := range byteUnits {
		if strings.HasSuffix(in, unit) {
			mul = byteUnits[unit]
			in = strings.TrimSuffix(in, unit)
			break
		}
	}
	i, err := strconv.ParseInt(in, 10, 64)
	if err != nil {
		return 0
	}
	return i * mul
}
