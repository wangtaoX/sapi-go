package sapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	sjson "github.com/bitly/go-simplejson"
	"github.com/gorilla/mux"
)

var (
	tors           = []string{}
	tp             map[string][]*Topology
	ts             map[string][]string
	refresh        = make(chan bool)
	tunnelSyncDone = make(chan bool)
	tplv           = make(map[string]*LocalVlan)
	tplvCache      = make(map[string]int)

	lv            = "/localvlan/"
	vlanVpcMin    = 2
	vlanVpcMax    = 4000
	vlanSharedMin = 4002
	vlanSharedMax = 4094

	torconf = "http://10.216.25.51:8081"
)

type LocalVlan struct {
	Shared, Unshared *BitMap
}

type VlanMapping struct {
	VlanId int    `json:"vlan_id"`
	NetId  string `json:"netid"`
	Host   string `json:"host"`
	Tor    string `json:"tor"`
}

func getId(tor, netid string, shared bool) (id string) {
	if shared {
		id = tor + netid + "-1"
	} else {
		id = tor + netid + "-0"
	}
	return
}

func getTunnelIds(tor string) []int {
	ids := []int{}
	tunnels := make([]*SapiTorTunnels, 0)
	SelectAllTunnelByTor(tor, &tunnels)

	for _, tunnel := range tunnels {
		ids = append(ids, tunnel.TunnelId)
	}
	return ids
}

func getVsis(tor string) []int {
	ids := []int{}
	vsis := make([]*SapiTorVsis, 0)
	SelectAllVsiByTor(tor, &vsis)

	for _, vsi := range vsis {
		ids = append(ids, vsi.Vxlan)
	}
	return ids
}

func GetTors() []string {
	ret := []string{}
	tors := make([]*SapiTor, 0)
	SelectAllTors(&tors)

	for _, tor := range tors {
		ret = append(ret, tor.TorIp)
	}
	return ret
}

func allocateVlanId(tor string, vid *uint32, shared bool) bool {
	//allocate a new local vlan id.
	if shared {
		return tplv[tor].Shared.GetUnusedBit(vid)
	} else {
		return tplv[tor].Unshared.GetUnusedBit(vid)
	}
}

func releaseVlanId(tor string, vid uint32, shared bool) {
	if shared {
		tplv[tor].Shared.UnsetBit(vid)
	} else {
		tplv[tor].Unshared.UnsetBit(vid)
	}
}

func deleteSva(netid, tor string, vlanId int, shared bool) {
	sva := new(SapiVlanAllocations)
	sva.NetworkId = netid
	sva.TorIp = tor
	sva.VlanId = vlanId
	sva.Allocated = true
	sva.Shared = shared
	sva.delete()
}

func addNewSva(netid, tor string, vlanId int, shared bool) {
	sva := new(SapiVlanAllocations)
	sva.NetworkId = netid
	sva.TorIp = tor
	sva.VlanId = vlanId
	sva.Allocated = true
	sva.Shared = shared
	sva.insert()
}

func addNewVsi(tor string, vxlan int) {
	vsi := new(SapiTorVsis)
	vsi.TorIp = tor
	vsi.Vxlan = vxlan
	vsi.insert()
}

func deleteVsi(tor string, vxlan int) {
	vsi := new(SapiTorVsis)
	vsi.TorIp = tor
	vsi.Vxlan = vxlan
	vsi.delete()
}

func addNewPvm(portid, netid, tor string, vlanId, index int) {
	pvm := new(SapiPortVlanMapping)
	pvm.NetworkId = netid
	pvm.TorIp = tor
	pvm.VlanId = vlanId
	pvm.Index = index
	pvm.insert()
}

func addNewTunnel(tor, dst string, id int) {
	tunnel := new(SapiTorTunnels)
	tunnel.TorIp = tor
	tunnel.TunnelId = id
	tunnel.DstAddr = dst
	tunnel.insert()
}

func selectUptor(h string) string {
	for torIp, hosts := range ts {
		for _, host := range hosts {
			if h == host {
				return torIp
				break
			}
		}
	}
	return ""
}

func selectIndex(h, tor string) (index int) {
	indexs, _ := tp[tor]
	for _, item := range indexs {
		if h == item.Host {
			index, _ = strconv.Atoi(item.Index)
			break
		}
	}
	return index
}

func goTunnelSync(torType, tor, src, dst string) {
	tunnel_id := struct {
		TunnelId int `json:"tunnel_id"`
	}{}
	client := NewHttpAgent()
	_, bodystr, _ := client.Post(torconf + "/tunnel").ReqData(
		struct {
			Type string `json:"type"`
			Mgr  string `json:"mgr"`
			Src  string `json:"src"`
			Dst  string `json:"dst"`
		}{
			Type: torType,
			Mgr:  tor,
			Src:  src,
			Dst:  dst},
	).Issue()
	if err := json.Unmarshal([]byte(bodystr), &tunnel_id); err != nil {
		Log().WithFields(logrus.Fields{
			"Error": err,
			"Body":  bodystr,
		}).Error("registerTor: /tunnel response error")
	}
	addNewTunnel(tor, dst, tunnel_id.TunnelId)
	Log().WithFields(logrus.Fields{
		"Switch":     tor,
		"Source":     src,
		"Destnation": dst,
	}).Info("registerTor: New vxlan tunnel")
}

func registerTor(rw http.ResponseWriter, r *http.Request) {
	var (
		data struct {
			Type string `json:"switch_type"`
			Mgr  string `json:"mgr"`
			Src  string `json:"tunnel_src"`
		}
		sapiTor  = new(SapiTor)
		sapiTors = make([]*SapiTor, 0)
	)

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		fmt.Println(err)
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}
	if data.Type == "" || data.Mgr == "" || data.Src == "" {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}
	if err := SelectAllTors(&sapiTors); err != nil {
		HttpError(rw, "Database Error", err, http.StatusInternalServerError)
		return
	}

	sapiTor.TorIp = data.Mgr
	sapiTor.TunnelSrcIp = data.Src
	sapiTor.Type = data.Type
	Log().Info(fmt.Sprintf("registerTor: Tor %s", sapiTor))
	sapiTor.insert()

	//need refresh topology
	tors = append(tors, sapiTor.TorIp)
	InitInmemoryData([]string{sapiTor.TorIp})
	refresh <- true

	//TODO: tunnel sync
	go func() {
		client := NewHttpAgent()
		Log().Info(fmt.Sprintf("registerTor: make tunnel sync request %s", torconf+"/tsync"))
		_, bodystr, _ := client.Post(torconf + "/tsync").ReqData(
			struct {
				Src        string `json:"src"`
				TunnelType string `json:"tunnel_type"`
			}{
				Src:        sapiTor.TunnelSrcIp,
				TunnelType: "vxlan",
			},
		).Issue()
		Log().Info(fmt.Sprintf("registerTor: request got response %s", bodystr))

		respj, _ := sjson.NewFromReader(strings.NewReader(bodystr))
		tunnels, _ := respj.CheckGet("tunnels")
		all, _ := tunnels.Array()
		for _, tunnel := range all {
			if m, ok := tunnel.(map[string]interface{}); ok {
				ip, _ := m["ip_address"].(string)
				if ip != sapiTor.TunnelSrcIp {
					//new tunnel with existing tunnel src ip.
					goTunnelSync(sapiTor.Type, sapiTor.TorIp, sapiTor.TunnelSrcIp, ip)
				}
			}
		}

		for _, tor := range sapiTors {
			//new tunnel form existing tunnel src ip to new tunnel src ip.
			goTunnelSync(tor.Type, tor.TorIp, tor.TunnelSrcIp, sapiTor.TunnelSrcIp)
		}
		Log().Info(fmt.Sprintf("registerTor: tunnel sync done, send signal to channel tunnelSyncDone"))
		tunnelSyncDone <- true
	}()
	rw.Write([]byte("OK"))
}

func makeLocalvlanMap(rw http.ResponseWriter, r *http.Request) {
	var (
		data struct {
			NetId  string `json:"netid"`
			Host   string `json:"host"`
			PortId string `json:"portid"`
		}
		vlanmapping VlanMapping
		upTor       string
		id          string
		message     string
		vid         uint32
	)

	rw.Header().Set("Content/Type", "application/json")
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}
	if data.NetId == "" || data.Host == "" || data.PortId == "" {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}
	upTor = selectUptor(data.Host)
	if upTor == "" {
		HttpError(rw, "Host not found", nil, http.StatusNotFound)
		return
	}
	net := new(SapiProvisionedNets)
	has, _ := net.search(data.NetId)
	if !has {
		HttpError(rw, "Net not found", nil, http.StatusNotFound)
		return
	}
	has, _ = new(SapiProvisionedPorts).search(data.PortId)
	if !has {
		HttpError(rw, "Port not found", nil, http.StatusNotFound)
		return
	}

	id = getId(upTor, data.NetId, net.Shared)
	vlanId, ok := tplvCache[id]
	if !ok {
		if !allocateVlanId(upTor, &vid, net.Shared) {
			HttpError(rw, "No avaliable id to allocate", nil, http.StatusBadRequest)
			return
		}
		//add corrsponding records in database
		vlanId = int(vid)
		tplvCache[id] = vlanId
		addNewSva(data.NetId, upTor, vlanId, net.Shared)
		addNewVsi(upTor, net.SegmentationId)
		Log().WithFields(logrus.Fields{
			"Tor":   upTor,
			"Vlan":  vlanId,
			"Vxlan": net.SegmentationId,
		}).Info("makeLocalvlanMap: New SapiVlanAllocations.")
		Log().WithFields(logrus.Fields{
			"Tor": upTor,
			"Vsi": net.SegmentationId,
		}).Info("makeLocalvlanMap: New SapiTorVsis.")
	}

	index := selectIndex(data.Host, upTor)
	tunnel_ids := getTunnelIds(upTor)
	addNewPvm(data.PortId, data.NetId, upTor, vlanId, index)
	Log().WithFields(logrus.Fields{
		"Tor":   upTor,
		"Vxlan": net.SegmentationId,
		"Vlan":  vlanId,
		"Port":  data.PortId,
		"Index": index,
	}).Info("makeLocalvlanMap: New SapiPortVlanMapping.")

	//config tor
	go func() {
		client := NewHttpAgent()
		Log().WithFields(logrus.Fields{
			"Tor":     upTor,
			"Vxlan":   net.SegmentationId,
			"Vlan":    vlanId,
			"Tunnels": tunnel_ids,
			"Index":   index,
		}).Info("makeLocalvlanMap: request torconf /vlan2vxlan")
		client.Post(torconf + "/vlan2vxlan").ReqData(
			struct {
				Type      string `json:"type"`
				Vlan      int    `json:"vlan"`
				Vxlan     int    `json:"vxlan"`
				TunnelIds []int  `json:"tunnel_ids"`
				Mgr       string `json:"mgr"`
				Index     int    `json:"index"`
			}{
				Type:      "h3c",
				Vlan:      vlanId,
				Vxlan:     net.SegmentationId,
				TunnelIds: tunnel_ids,
				Mgr:       upTor,
				Index:     index},
		).Issue()
	}()

	message = "OK"
	vlanmapping.VlanId = vlanId
	vlanmapping.Tor = upTor
	vlanmapping.NetId = data.NetId
	vlanmapping.Host = data.Host

	ret, _ := json.MarshalIndent(
		struct {
			Vm      VlanMapping `json:"vlanmapping"`
			Message string      `json:"message"`
		}{
			Vm:      vlanmapping,
			Message: message,
		}, "", "    ")
	rw.Write(ret)
}

func deleteLocalvlanMap(rw http.ResponseWriter, r *http.Request) {
	var pvm = new(SapiPortVlanMapping)
	var net = new(SapiProvisionedNets)
	var id string
	var onlyIndex = true
	portId, _ := mux.Vars(r)["id"]

	if has, _ := pvm.search(portId); !has {
		HttpError(rw, "Not found", nil, http.StatusNotFound)
		return
	}

	if pvm.count() == 1 {
		onlyIndex = false
		net.search(pvm.NetworkId)
		id = getId(pvm.TorIp, pvm.NetworkId, net.Shared)
		//release this vlan id
		releaseVlanId(pvm.TorIp, uint32(pvm.VlanId), net.Shared)
		//delete key in tplvCache
		delete(tplvCache, id)
		//delete corrsponding record in database
		deleteSva(pvm.NetworkId, pvm.TorIp, pvm.VlanId, net.Shared)
		deleteVsi(pvm.TorIp, net.SegmentationId)
		Log().WithFields(logrus.Fields{
			"Tor":   pvm.TorIp,
			"Vxlan": net.SegmentationId,
			"Vlan":  pvm.VlanId,
		}).Info("deleteLocalvlanMap: release SapiVlanAllocations.")
		Log().WithFields(logrus.Fields{
			"Tor": pvm.TorIp,
			"Vsi": net.SegmentationId,
		}).Info("deleteLocalvlanMap: release SapiTorVsis.")
	}
	//delete port mapping record
	vlanId := pvm.VlanId
	index := pvm.Index
	upTor := pvm.TorIp
	pvm.delete()

	go func() {
		client := NewHttpAgent()
		Log().WithFields(logrus.Fields{
			"Tor":       pvm.TorIp,
			"Vxlan":     net.SegmentationId,
			"Vlan":      vlanId,
			"Index":     index,
			"OnlyIndex": onlyIndex,
		}).Info("deleteLocalvlanMap: request torconf /vlan2vxlan.")
		client.Delete(torconf + "/vlan2vxlan").ReqData(
			struct {
				Type  string `json:"type"`
				Vlan  int    `json:"vlan"`
				Vxlan int    `json:"vxlan"`
				Mgr   string `json:"mgr"`
				Index int    `json:"index"`
				Only  bool   `json:"delete_on_index"`
			}{
				Type:  "h3c",
				Vlan:  vlanId,
				Vxlan: net.SegmentationId,
				Mgr:   upTor,
				Index: index,
				Only:  onlyIndex},
		).Issue()
	}()

	rw.Write([]byte("OK"))
}

func getTopology(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content/Type", "application/json")

	ret, _ := json.MarshalIndent(
		struct {
			Topology       map[string][]*Topology `json:"topology"`
			TopologySimple map[string][]string    `json:"topology_simple"`
		}{
			Topology:       tp,
			TopologySimple: ts}, "", "  ")
	w.Write(ret)
}

func emptyTopology(w http.ResponseWriter, r *http.Request) {
	refresh <- true
	w.Write([]byte("Ready to refresh."))
}

//init in memory data, eg: tplvCache and initialized data in database
func InitInmemoryData(inputs []string) {
	for _, input := range inputs {
		tors = append(tors, input)
	}
	for _, tor := range tors {
		tplv[tor] = &LocalVlan{}
		tplv[tor].Shared = NewBitmap(uint32(vlanSharedMin), uint32(vlanSharedMax))
		tplv[tor].Unshared = NewBitmap(uint32(vlanVpcMin), uint32(vlanVpcMax))
	}

	everyAlloctions := make([]*SapiVlanAllocations, 0)
	if err := SelectAllVlanAlloctions(&everyAlloctions); err != nil {
		Log().Error(fmt.Sprintf("%s", err))
		os.Exit(0)
	}
	for _, alloction := range everyAlloctions {
		if alloction.Shared {
			tplv[alloction.TorIp].Shared.Setbit(uint32(alloction.VlanId))
			tplvCache[alloction.TorIp+alloction.NetworkId+"-1"] = alloction.VlanId
		} else {
			tplv[alloction.TorIp].Unshared.Setbit(uint32(alloction.VlanId))
			tplvCache[alloction.TorIp+alloction.NetworkId+"-0"] = alloction.VlanId
		}
	}
}

//run periodic updata for topology
func GoTopology(done chan struct{}) {
	tp, ts = GetTopology(tors, "public", time.Second*10)
	go func() {
		Seconds := time.NewTimer(time.Second * 30)
		for {
			select {
			case <-Seconds.C:
				tp, ts = GetTopology(tors, "public", time.Second*10)
				Seconds.Reset(time.Second * 30)
			case <-refresh:
				Log().WithFields(logrus.Fields{
					"Time": time.Now(),
				}).Info("updateTopology: refresh received")
				tp, ts = GetTopology(tors, "public", time.Second*10)
			case <-done:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-tunnelSyncDone:
				//ensure all vxlan associated with all tunnels
				Log().Info("tunnelSyncDone: receive signal from master, gogo")
				tors := make([]*SapiTor, 0)
				if err := SelectAllTors(&tors); err != nil {
					Log().WithFields(logrus.Fields{
						"Error": err,
					}).Info("tunnelSyncDone: SelectAllTors error.")
					continue
				}
				for _, tor := range tors {
					ids := getTunnelIds(tor.TorIp)
					vsis := getVsis(tor.TorIp)
					Log().WithFields(logrus.Fields{
						"Switch":    tor.TorIp,
						"Vsis":      vsis,
						"TunnelIds": ids,
					}).Info("tunnelSyncDone: sync switch vxlan with tunnels")
					client := NewHttpAgent()
					client.Post(torconf + "/ensure").ReqData(
						struct {
							Type    string `json:"type"`
							Mgr     string `json:"mgr"`
							Vxlans  []int  `json:"vxlans"`
							Tunnels []int  `json:"tunnels"`
						}{
							Type:    tor.Type,
							Vxlans:  vsis,
							Mgr:     tor.TorIp,
							Tunnels: ids},
					).Issue()
				}
			case <-done:
				return
			}
		}
	}()
}

func init() {
	Regist(MakeApiEndpoints("GET", "/topology", http.HandlerFunc(getTopology)))
	Regist(MakeApiEndpoints("GET", "/refresh", http.HandlerFunc(emptyTopology)))
	Regist(MakeApiEndpoints("POST", lv, http.HandlerFunc(makeLocalvlanMap)))
	Regist(MakeApiEndpoints("DELETE", lv+"{id}", http.HandlerFunc(deleteLocalvlanMap)))
	Regist(MakeApiEndpoints("POST", "/tsync", http.HandlerFunc(registerTor)))
}
