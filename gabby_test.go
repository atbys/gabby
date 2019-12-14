package gabby

import (
	"log"
	"testing"
)

func TestCreateEngine(t *testing.T) {
	e, err := New()
	if err != nil {
		log.Fatal(err)
	}
	if e.DeviceName != "" {
		t.Errorf("device error")
	}
}
