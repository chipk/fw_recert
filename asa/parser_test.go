package asa

import (
	"os"
	"testing"
)

func TestParse(t *testing.T) {

	r, err := os.Open("Cisco-ASA5506-config.txt")

	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}

	config := Parse(r)

	t.Logf("config = %#v", config)
}
