package sapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/codegangsta/negroni"
)

var (
	nets  = make([]*SapiProvisionedNets, 5)
	ports = make([]*SapiProvisionedPorts, 9)
	m     *negroni.Negroni
)

type Resp struct {
	Vm      VlanMapping `json:"vlanmapping"`
	Message string      `json:"message"`
}

func getresp(rec *httptest.ResponseRecorder) (*Resp, error) {
	resp := new(Resp)
	decoder := json.NewDecoder(rec.Body)
	if err := decoder.Decode(resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func NewNet(id string, sid int, shared bool, s *SapiProvisionedNets) {
	s.NetworkId = id
	s.TenantId = "faker"
	s.SegmentationType = "vxlan"
	s.SegmentationId = sid
	s.Shared = shared
	s.AdminStateUp = true
}

func init() {
	var shared bool

	InitDb("sapi", "sapi", "10.216.25.57", "sapi")
	Truncate([]string{"sapi_provisioned_nets",
		"sapi_provisioned_ports",
		"sapi_port_vlan_mapping",
		"sapi_vlan_allocations"})
	ts = map[string][]string{
		"tor1": []string{"compute1", "compute2"},
		"tor2": []string{"compute3", "compute4"},
	}

	for i := 1; i <= 4; i++ {
		if i%2 == 0 {
			shared = true
		} else {
			shared = false
		}
		nets[i] = new(SapiProvisionedNets)
		NewNet("network"+strconv.Itoa(i), i, shared, nets[i])
		nets[i].insert()
	}
	for i := 1; i <= 8; i++ {
		ports[i] = new(SapiProvisionedPorts)
		ports[i].PortId = "port" + strconv.Itoa(i)
		ports[i].insert()
	}

	InitInmemoryData([]string{"tor1", "tor2"})
	m = negroni.New()
	Regist(MakeApiEndpoints("POST", lv, http.HandlerFunc(makeLocalvlanMap)))
	Regist(MakeApiEndpoints("DELETE", lv+"{id}", http.HandlerFunc(deleteLocalvlanMap)))
	m.UseHandler(gRouter)
}

func TestNoSuchPort(t *testing.T) {
	r, _ := http.NewRequest("POST", "/localvlan/", strings.NewReader(`{"portid":"port9","netid":"network1","host":"compute1"}`))
	r.Header.Set("Content/Type", "application/json")
	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	if recorder.Code != 404 {
		t.Error("Response not 404")
	}
}

func TestHavePort(t *testing.T) {
	r, _ := http.NewRequest("POST", "/localvlan/", strings.NewReader(`{"portid":"port1","netid":"network1","host":"compute1"}`))
	r.Header.Set("Content/Type", "application/json")
	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	if recorder.Code != 200 {
		t.Error("Response not 200")
	}

	resp, err := getresp(recorder)
	if err != nil {
		t.Errorf("resp decode error %s", err)
	}
	if resp.Vm.VlanId != 2 {
		t.Errorf("Expected vlan id %d, but got %d", 2, resp.Vm.VlanId)
	}
}

func TestDelete(t *testing.T) {
	r, _ := http.NewRequest("DELETE", "/localvlan/port1", nil)
	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	if recorder.Code != 200 {
		t.Error("Response not 200")
	}
}

func TestUnsharedVlan(t *testing.T) {
	r, _ := http.NewRequest("POST", "/localvlan/", strings.NewReader(`{"portid":"port1","netid":"network1","host":"compute1"}`))
	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	r, _ = http.NewRequest("POST", "/localvlan/", strings.NewReader(`{"portid":"port2","netid":"network1","host":"compute2"}`))
	recorder = httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	resp, err := getresp(recorder)
	if err != nil {
		t.Errorf("resp decode error %s", err)
	}
	if resp.Vm.VlanId != 2 {
		t.Errorf("Expected vlan id %d, but got %d", 2, resp.Vm.VlanId)
	}

	r, _ = http.NewRequest("POST", "/localvlan/", strings.NewReader(`{"portid":"port5","netid":"network3","host":"compute1"}`))
	recorder = httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	resp, err = getresp(recorder)
	if err != nil {
		t.Errorf("resp decode error %s", err)
	}
	if resp.Vm.VlanId != 3 {
		t.Errorf("Expected vlan id %d, but got %d", 3, resp.Vm.VlanId)
	}
}

func TestSharedVlan(t *testing.T) {
	r, _ := http.NewRequest("POST", "/localvlan/", strings.NewReader(`{"portid":"port3","netid":"network2","host":"compute1"}`))
	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	resp, err := getresp(recorder)
	if err != nil {
		t.Errorf("resp decode error %s", err)
	}
	if resp.Vm.VlanId != 4002 {
		t.Errorf("Expected vlan id %d, but got %d", 2, resp.Vm.VlanId)
	}

	r, _ = http.NewRequest("DELETE", "/localvlan/port5", nil)
	recorder = httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	if recorder.Code != 200 {
		t.Error("Response not 200")
	}
}

func TestReAllocateVlan(t *testing.T) {
	r, _ := http.NewRequest("POST", "/localvlan/", strings.NewReader(`{"portid":"port5","netid":"network3","host":"compute1"}`))
	recorder := httptest.NewRecorder()
	m.ServeHTTP(recorder, r)

	resp, err := getresp(recorder)
	if err != nil {
		t.Errorf("resp decode error %s", err)
	}
	if resp.Vm.VlanId != 3 {
		t.Errorf("Expected vlan id %d, but got %d", 3, resp.Vm.VlanId)
	}
}

func TestClean(t *testing.T) {
	Truncate([]string{"sapi_provisioned_nets",
		"sapi_provisioned_ports",
		"sapi_port_vlan_mapping",
		"sapi_vlan_allocations"})
}
