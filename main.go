package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
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

	OutputScanner(io.Reader(output), os.Stderr, func(bytes int64) {
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

// OutputScanner will keep scanning for "%d bytes" in a io.Reader and call
// the supplied callback when it matches. Any line not starting with an integer
// will be printed to given output.
func OutputScanner(reader io.Reader, output io.Writer, callback func(int64)) {
	scanner := bufio.NewScanner(reader)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if !unicode.IsDigit(rune(line[0])) && output != nil {
				fmt.Fprintln(output, line)
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

// Interrupter will send the `InfoSignal` to the given process every interval.
func Interrupter(process *os.Process, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			process.Signal(InfoSignal)
		}
	}()
}

// Trap will listen for all stop signals and pass them along to the given
// process.
func Trap(process *os.Process) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGKILL, syscall.SIGQUIT)
	go func() {
		process.Signal(<-c)
	}()
}
