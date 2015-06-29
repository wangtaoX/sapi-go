package sapi

import "time"
import g "github.com/soniah/gosnmp"

type SnmpAgent struct {
	gs *g.GoSNMP
}

func NewSnmpAgent() *SnmpAgent {
	return &SnmpAgent{gs: &g.GoSNMP{Port: 161}}
}

func (sa *SnmpAgent) Target(target string) *SnmpAgent {
	sa.gs.Target = target
	return sa
}

func (sa *SnmpAgent) Community(community string) *SnmpAgent {
	sa.gs.Community = community
	return sa
}

func (sa *SnmpAgent) Version(v g.SnmpVersion) *SnmpAgent {
	sa.gs.Version = v
	return sa
}

func (sa *SnmpAgent) Timeout(t time.Duration) *SnmpAgent {
	sa.gs.Timeout = t
	return sa
}

func (sa *SnmpAgent) GS() *g.GoSNMP {
	return sa.gs
}
