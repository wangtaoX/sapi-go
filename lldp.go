package sapi

import "time"
import "strings"
import "encoding/json"

import "strconv"
import g "github.com/soniah/gosnmp"
import "encoding/hex"

var (
	LocalDataOid  = "1.0.8802.1.1.2.1.3"
	RemoteDataOid = "1.0.8802.1.1.2.1.4"

	lldpLocChassisIdSubtype = "1.0.8802.1.1.2.1.3.1.0"
	lldpLocChassisId        = "1.0.8802.1.1.2.1.3.2.0"
	lldpLocSysName          = "1.0.8802.1.1.2.1.3.3.0"
	lldpLocSysDesc          = "1.0.8802.1.1.2.1.3.4.0"
	lldpLocPorts            = "1.0.8802.1.1.2.1.3.7.1.3"
	lldpLocManAddress       = "1.0.8802.1.1.2.1.3.8.1.5"
	lldpLocal               = map[string]string{
		lldpLocChassisIdSubtype: lldpLocChassisIdSubtype,
		lldpLocChassisId:        lldpLocChassisId,
		lldpLocSysDesc:          lldpLocSysDesc,
		lldpLocPorts:            lldpLocPorts,
		lldpLocSysName:          lldpLocSysName,
		lldpLocManAddress:       lldpLocManAddress,
	}

	lldpRemChassisIdSubtype = "1.0.8802.1.1.2.1.4.1.1.4"
	lldpRemChassisId        = "1.0.8802.1.1.2.1.4.1.1.5"
	lldpRemPortDesc         = "1.0.8802.1.1.2.1.4.1.1.8"
	lldpRemSysName          = "1.0.8802.1.1.2.1.4.1.1.9"
	lldpRemSysDesc          = "1.0.8802.1.1.2.1.4.1.1.10"
	lldpRemManAddress       = "1.0.8802.1.1.2.1.4.2.1.3"
	lldpRemote              = map[string]string{
		lldpRemChassisIdSubtype: lldpRemChassisIdSubtype,
		lldpRemChassisId:        lldpRemChassisId,
		lldpRemPortDesc:         lldpRemPortDesc,
		lldpRemSysDesc:          lldpRemSysDesc,
		lldpRemSysName:          lldpRemSysName,
		lldpRemManAddress:       lldpRemManAddress,
	}
	VeryCloudy = []string{
		"             ",
		"\033[38;5;240;1m     .--.    \033[0m",
		"\033[38;5;240;1m  .-(    ).  \033[0m",
		"\033[38;5;240;1m (___.__)__) \033[0m",
		"             "}
)

type LocalData struct {
	MacAddress string            `json:"mac_address"`
	SysName    string            `json:"sysname"`
	SysDesc    string            `json:"sysdesc"`
	ManAddress map[string]string `json:"mgr"`
	Ports      map[string]string `json:"ports"`
}

func (ld *LocalData) String() string {
	b, _ := json.Marshal(*ld)
	return string(b)
}

func newl() *LocalData {
	l := &LocalData{}

	l.ManAddress = make(map[string]string)
	l.Ports = make(map[string]string)

	return l
}

type RemoteData struct {
	RemMacAddress string `json:"rem_mac_address"`
	RemPortDesc   string `json:"rem_desc"`
	RemSysName    string `json:"rem_sysname"`
	RemSysDesc    string `json:"rem_sysdesc"`
	RemManAddress string `json:"rem_mgr"`
	Index         string `json:"index"`
}

func (rd *RemoteData) String() string {
	b, _ := json.Marshal(*rd)
	return string(b)
}

func newr() *RemoteData {
	return &RemoteData{}
}

type LLDP struct {
	Local  *LocalData             `json:"local"`
	Remote map[string]*RemoteData `json:"remote"`
}

func (lldp *LLDP) String() string {
	b, _ := json.Marshal(*lldp)
	return string(b)
}

func isOid(oid string, ll map[string]string) string {
	for k, _ := range ll {
		if strings.HasPrefix(oid, "."+k) {
			return k
		}
	}
	return ""
}

func parseLocalData(res []g.SnmpPDU) *LocalData {
	l := newl()

	for _, pdu := range res {
		t := isOid(pdu.Name, lldpLocal)
		switch t {
		case lldpLocChassisId:
			l.MacAddress = hex.EncodeToString(pdu.Value.([]byte))
		case lldpLocSysName:
			l.SysName = string(pdu.Value.([]byte))
		case lldpLocSysDesc:
			l.SysDesc = string(pdu.Value.([]byte))
		case lldpLocPorts:
			tmp := strings.Split(pdu.Name, ".")
			index := tmp[len(tmp)-1]
			_, ok := l.Ports[index]
			if !ok {
				l.Ports[index] = string(pdu.Value.([]byte))
			}
		case lldpLocManAddress:
			tmp := strings.Split(pdu.Name, ".")
			ip := strings.Join(tmp[len(tmp)-4:], ".")
			index := strconv.Itoa(pdu.Value.(int))
			_, ok := l.ManAddress[index]
			if !ok {
				l.ManAddress[index] = ip
			}
		}
	}
	return l
}

func parseRemoteData(res []g.SnmpPDU) map[string]*RemoteData {
	var index string
	r := make(map[string]*RemoteData)

	for _, pdu := range res {
		t := isOid(pdu.Name, lldpRemote)
		tmp := strings.Split(pdu.Name, ".")
		index = tmp[len(tmp)-2]
		if t == lldpRemManAddress {
			index = tmp[len(tmp)-8]
		}
		if t == "" {
			continue
		}
		_, ok := r[index]
		if !ok {
			r[index] = newr()
		}
		r[index].Index = index
		switch t {
		case lldpRemChassisId:
			r[index].RemMacAddress = hex.EncodeToString(pdu.Value.([]byte))
		case lldpRemManAddress:
			r[index].RemManAddress = strings.Join(tmp[len(tmp)-4:len(tmp)], ".")
		case lldpRemSysName:
			r[index].RemSysName = string(pdu.Value.([]byte))
		case lldpRemSysDesc:
			r[index].RemSysDesc = string(pdu.Value.([]byte))
		case lldpRemPortDesc:
			r[index].RemPortDesc = string(pdu.Value.([]byte))
		}
	}

	return r
}

func retrieve(target string, community string, t time.Duration) (*LLDP, error) {
	gs := NewSnmpAgent().
		Target(target).
		Community(community).
		Timeout(t).
		Version(g.Version2c).GS()

	err := gs.Connect()
	if err != nil {
		return nil, err
	}
	defer gs.Conn.Close()

	res, err := gs.WalkAll(LocalDataOid)
	if err != nil {
		return nil, err
	}
	l := parseLocalData(res)

	res, err = gs.WalkAll(RemoteDataOid)
	if err != nil {
		return nil, err
	}
	d := parseRemoteData(res)

	return &LLDP{l, d}, nil
}

func GetTopology(tors []string, community string, t time.Duration) (map[string][]*Topology, map[string][]string) {
	var tp1 = make(map[string][]*Topology)
	var tp2 = make(map[string][]string)

	for _, tor := range tors {
		lldp, err := retrieve(tor, community, t)
		if err != nil {
			continue
		}
		tp1Tmp := lldp.resolveTopology()
		tp2Tmp := topologySimple(tp1Tmp)

		for _, v := range tp1Tmp {
			tp1[v.TorIp] = append(tp1[v.TorIp], v)
		}
		for k, v := range tp2Tmp {
			tp2[k] = v
		}
	}
	return tp1, tp2
}

//Topology got by lldp information
type Topology struct {
	Index         string
	IndexName     string
	Tor           string
	TorIp         string
	Host          string
	HostInterface string
	Ip            string
	Mac           string
	SysDesc       string
}

func (t *Topology) String() string {
	b, _ := json.Marshal(*t)
	return string(b)
}

func (lldp *LLDP) resolveTopology() map[string]*Topology {
	var index, ip string
	topo := make(map[string]*Topology)

	tor := lldp.Local.SysName
	for index, ip = range lldp.Local.ManAddress {
		if _, ok := lldp.Local.Ports[index]; ok {
			break
		}
	}
	torIp := ip

	for index, rem := range lldp.Remote {
		if !checkForLinux(rem.RemSysDesc) {
			continue
		}

		if _, ok := topo[index]; !ok {
			topo[index] = &Topology{
				Tor:   tor,
				TorIp: torIp}
		}
		topo[index].Host = rem.RemSysName
		topo[index].HostInterface = rem.RemPortDesc
		topo[index].SysDesc = rem.RemSysDesc
		topo[index].Mac = rem.RemMacAddress
		topo[index].Ip = rem.RemManAddress
		topo[index].IndexName = lldp.Local.Ports[index]
		topo[index].Index = index
	}

	return topo
}

func checkForLinux(desc string) bool {
	var linux = []string{"Linux", "Ubuntu"}

	for _, s1 := range linux {
		if strings.Contains(desc, s1) {
			return true
		}
	}
	return false
}

func topologySimple(topo map[string]*Topology) map[string][]string {
	topoSimple := make(map[string][]string)

	for _, tp := range topo {
		if _, ok := topoSimple[tp.TorIp]; !ok {
			topoSimple[tp.TorIp] = []string{}
		}
		topoSimple[tp.TorIp] = append(topoSimple[tp.TorIp], tp.Host)
	}
	return topoSimple
}
