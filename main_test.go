package main_test

import (
	"bytes"
	"testing"
	"time"

	. "./"
)

func TestOutputScanner(t *testing.T) {
	var result int64
	called := false
	in := bytes.NewBufferString(`19+0 records in
19+0 records in
error here
19+0 records out
741 bytes transferred in 0.32 secs (1337 bytes/sec)`)
	out := &bytes.Buffer{}

	OutputScanner(in, out, func(b int64) {
		result += b
		called = true
	})

	for !called {
		time.Sleep(1)
	}

	if result != 741 {
		t.Errorf("callback called with %d, want 741", result)
	}
	if out.String() != "error here\n" {
		t.Errorf("output recieved %q, want \"error here\n\"", out.String())
	}
}
