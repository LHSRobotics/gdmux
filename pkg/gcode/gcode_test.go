package gcode

import (
	"testing"
	"io"
	"os"
)

func TestSamples(t *testing.T) {
	correctfiles := []string{
		"samples/spiral.gcode",
		"samples/h.gcode",
		"samples/gopro.gcode",
		"samples/glasses.gcode",
		"samples/gopro.nc",
	}
	for _, f := range correctfiles {
		r, err := os.Open(f)
		if err != nil {
			t.Logf("couldn't open test input: %v", err)
			t.Fail()
		}
		p := NewParser(r)
		
		for {
			_, err = p.Next()
			if err == io.EOF {
				break
			} else if err != nil {
				t.Logf("parse error: %v", err)
				t.Fail()
			}
		}
	}
}
