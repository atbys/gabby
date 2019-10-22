package gomamiso

import (
	"testing"
	"time"
)

func TestCreateEngine(t *testing.T) {
	e := New()
	if e.Device != "" {
		t.Errorf("device error")
	}
}

func TestSetDevice(t *testing.T) {
	e := New()
	e.SetDevice("AAAA")
	if e.Device != "AAAA"{
		t.Errorf("set device error")
	}
	e.SetDevice("enp0s31f6")
	if e.Device != "enp0s31f6"{
		t.Errorf("set device error")
	}
}

func TestDefaultSetting(t *testing.T) {
	e := Default()
	
	if e.Device != "enp0" || e.snapshot_len != 1024 || e.timeout != 30 * time.Second || e.promiscuous != true {
		t.Errorf("failed default setting")
	}
}
