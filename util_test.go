package main_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	. "./"
)

func TestBlackboxGuessTargetSize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	f := func(bs, count bytesize) bool {
		command := fmt.Sprintf("dd if=/dev/zero bs=%s count=%s", bs, count)
		argv := strings.Split(command, " ")
		guessed := GuessTargetSize(argv)

		out, err := exec.Command("dd", argv[1:]...).Output()
		if err != nil {
			return false
		}

		return guessed == int64(len(out))
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestGuessTargetSize(t *testing.T) {
	t.Parallel()

	tempsize := int64(6 * 1024)
	file, _ := ioutil.TempFile("", "ddp-test")
	file.WriteAt([]byte{0}, tempsize-1)
	defer os.Remove(file.Name())

	tests := []struct {
		in  string
		out int64
	}{
		{"bs=1 count=42", 42},
		{"count=21 bs=2", 42},
		{"bs=1k count=42", 43008},
		{"bs=1m count=101", 105906176},
		{"if=" + file.Name(), tempsize},
		{"if=" + file.Name() + " skip=5", tempsize - 5},
		{"if=" + file.Name() + " skip=1k", tempsize - 1024},
		{"if=" + file.Name() + " bs=2k count=2", 1024 * 2 * 2},
		{"if=" + file.Name() + " bs=1k count=4 skip=3", 1024 * 3}, // overlap is 3k
	}
	for _, tt := range tests {
		in := strings.Split("ddp "+tt.in, " ")
		res := GuessTargetSize(in)
		if res != tt.out {
			t.Errorf("ParseByteUnits(%q) => %d, want %d", tt.in, res, tt.out)
		}
	}
}

func TestParseByteUnits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in  string
		out int64
	}{
		{"1", 1},
		{"16", 16},
		{"1k", 1024},
		{"1kb", 1000},
		{"42k", 1024 * 42},
		{"42kb", 1000 * 42},
		{"1337t", 1024 * 1024 * 1024 * 1024 * 1337},
		{"2MB", 1000 * 1000 * 2},
		{"invalid input", 0},
		{"e", 0},
	}
	for _, tt := range tests {
		res := ParseByteUnits(tt.in)
		if res != tt.out {
			t.Errorf("ParseByteUnits(%q) => %d, want %d", tt.in, res, tt.out)
		}
	}
}

type bytesize string

func (b bytesize) Generate(rand *rand.Rand, size int) reflect.Value {
	suffix := ""
	number := 1 + rand.Int31n(int32(size))
	chance := rand.Float32()
	if chance < 0.1 {
		suffix = "k"
	}
	// Can't test `kb`-style suffixes in darwin…
	return reflect.ValueOf(bytesize(fmt.Sprintf("%d%s", number, suffix)))
}
