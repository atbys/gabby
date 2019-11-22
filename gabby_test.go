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

func TestSetDevice(t *testing.T) {
	e, err := New()
	if err != nil {
		log.Fatal(err)
	}
	e.SetDevice("AAAA")
	if e.Device != "AAAA" {
		t.Errorf("set device error")
	}
	e.SetDevice("enp0s31f6")
	if e.Device != "enp0s31f6" {
		t.Errorf("set device error")
	}
}
