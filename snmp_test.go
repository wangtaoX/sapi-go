package sapi

import "testing"
import "time"
import g "github.com/soniah/gosnmp"

func TestSnmpAgent(t *testing.T) {
	var lc = &SnmpAgent{
		gs: &g.GoSNMP{
			Port:      161,
			Target:    "127.0.0.1",
			Community: "test",
			Version:   g.Version2c,
			Timeout:   time.Second * 2,
		},
	}

	c1 := NewSnmpAgent().Target("127.0.0.1").Timeout(time.Second * 2).Community("test").Version(g.Version2c).GS()
	c2 := lc.gs

	if c1.Target != c2.Target {
		t.Errorf("Expected Target %s, got %s", c2.Target, c1.Target)
	}

	if c1.Community != c2.Community {
		t.Errorf("Expected Community %s, got %s", c2.Community, c1.Community)
	}

	if c1.Timeout != c2.Timeout {
		t.Errorf("Expected Timeout %s, got %s", c2.Timeout, c1.Timeout)
	}

	if c1.Version != c2.Version {
		t.Errorf("Expected Version %s, got %s", c2.Version, c1.Version)
	}
}
