package sapi

import (
	"encoding/json"
	"errors"
	"net/http"

	sjson "github.com/bitly/go-simplejson"
)

var (
	sync                 = "/sync/"
	ErrorNoSinaOpenstack = errors.New("Key sina_openstack not found")
)

func Sync(rw http.ResponseWriter, r *http.Request) {
	if err := handleSync(r); err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	rw.Write([]byte("OK"))
}

func handleSync(r *http.Request) (err error) {
	post, err := sjson.NewFromReader(r.Body)
	if err != nil {
		return
	}
	openstack, ok := post.CheckGet("sina_openstack")
	if !ok {
		return ErrorNoSinaOpenstack
	}

	if err := syncNet(openstack); err != nil {
		return err
	}
	if err := syncSubnet(openstack); err != nil {
		return err
	}
	if err := syncPort(openstack); err != nil {
		return err
	}

	return nil
}

func syncSubnet(data *sjson.Json) (err error) {
	var sapiSubnets []*SapiProvisionedSubnets

	subnet, ok := data.CheckGet("subnet")
	if !ok {
		return ErrorNoSubnet
	}
	bytes, err := subnet.Encode()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, &sapiSubnets); err != nil {
		return err
	}

	new(SapiProvisionedSubnets).truncate()
	for _, subnet := range sapiSubnets {
		if err := subnet.insert(); err != nil {
			return err
		}
	}
	return nil
}

func syncNet(data *sjson.Json) (err error) {
	var sapiNets []*SapiProvisionedNets

	network, ok := data.CheckGet("network")
	if !ok {
		return ErrorNoNet
	}
	bytes, err := network.Encode()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, &sapiNets); err != nil {
		return err
	}

	new(SapiProvisionedNets).truncate()
	for _, net := range sapiNets {
		if err := net.insert(); err != nil {
			return err
		}
	}
	return nil
}

func syncPort(data *sjson.Json) (err error) {
	var sapiPorts []*SapiProvisionedPorts
	var fixips []*struct {
		Ips []map[string]string `json:"fixed_ips"`
	}

	port, ok := data.CheckGet("port")
	if !ok {
		return ErrorNoNet
	}
	bytes, err := port.Encode()
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, &sapiPorts); err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, &fixips); err != nil {
		return err
	}

	new(SapiProvisionedPorts).truncate()
	for index, port := range sapiPorts {
		//fmt.Println(index, *port, fixips[index].Ips[0]["ip_address"], fixips[index].Ips[0]["subnet_id"])
		if len(fixips[index].Ips) >= 1 {
			port.IpAddress = fixips[index].Ips[0]["ip_address"]
			port.SubnetId = fixips[index].Ips[0]["subnet_id"]
		}
		if err := port.insert(); err != nil {
			return err
		}
	}

	return nil
}

func init() {
	Regist(MakeApiEndpoints("POST", sync, http.HandlerFunc(Sync)))
}
