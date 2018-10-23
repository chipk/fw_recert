package asa

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {

	r, err := os.Open(filepath.Join("test_data", "Cisco-ASA5506-config.txt"))

	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	defer r.Close()

	config := Parse(r)

	t.Logf("config = %#v", config)
}
