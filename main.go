package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"

	"github.com/cheggaaa/pb"
)

func main() {
	stdout := os.Stdout
	os.Stdout = os.Stderr

	executable, err := exec.LookPath("dd")
	if err != nil {
		fmt.Printf("ddp: failed to find dd: %s\n", err)
		os.Exit(1)
	}

	// Create pipe attached to a reader:
	output, input, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// Setup process with _the_ three file descriptors:
	files := []*os.File{
		os.Stdin,
		stdout,
		input,
	}
	process, err := os.StartProcess(executable, os.Args, &os.ProcAttr{
		Files: files,
	})
	if err != nil {
		fmt.Printf("ddp: failed to start dd: %s\n", err)
		os.Exit(1)
	}

	Trap(process)

	target := GuessTargetSize(os.Args)
	bar := pb.New64(target)
	bar.SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.Output = os.Stderr
	started := false

	OutputScanner(io.Reader(output), func(bytes int64) {
		if !started {
			started = true
			bar.Start()
		}
		bar.Set64(bytes)
	})
	Interrupter(process, pb.DEFAULT_REFRESH_RATE)

	state, err := process.Wait()
	if err != nil {
		panic(err)
	}
	if started && state.Success() {
		bar.Finish()
	}
	output.Close()
	if !state.Success() {
		os.Exit(1)
	}
}

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
		filesize = stat.Size()
	}

	size := bs * (count - skip)
	if filesize > size {
		return filesize
	}
	return size
}

func ParseByteUnits(in string) int64 {
	units := map[string]int64{
		"kb": 1000,
		"mb": 1000 * 1000,
		"gb": 1000 * 1000 * 1000,
		"tb": 1000 * 1000 * 1000 * 1000,
		"k":  1024,
		"m":  1024 * 1024,
		"g":  1024 * 1024 * 1024,
		"t":  1024 * 1024 * 1024 * 1024,
	}
	in = strings.ToLower(in)
	mul := int64(1)
	for unit, _ := range units {
		if strings.HasSuffix(in, unit) {
			mul = units[unit]
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

func OutputScanner(reader io.Reader, callback func(int64)) {
	scanner := bufio.NewScanner(reader)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if !unicode.IsDigit(rune(line[0])) {
				fmt.Println(line)
				continue
			}
			var bytes int64
			var label string
			n, _ := fmt.Sscanf(line, "%d %s", &bytes, &label)
			if n == 2 && label == "bytes" {
				callback(bytes)
			}
		}
	}()
}

func Interrupter(process *os.Process, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			process.Signal(InfoSignal)
		}
	}()
}

func Trap(process *os.Process) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	go func() {
		process.Signal(<-c)
	}()
}
