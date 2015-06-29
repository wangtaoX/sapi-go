package sapi

import (
	"encoding/json"
	"errors"
	"net/http"

	sjson "github.com/bitly/go-simplejson"
	"github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var (
	port        = "/port"
	ErrorNoPort = errors.New("Key port not founded")
)

func (this *SapiProvisionedPorts) Get(rw http.ResponseWriter, r *http.Request) {
	var has bool
	var err error
	var sapiPort = &SapiProvisionedPorts{}
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	has, err = sapiPort.search(id)
	if !has {
		HttpError(rw, "Not Found", nil, http.StatusNotFound)
		return
	}
	if err != nil {
		HttpError(rw, "Database Error", err, http.StatusInternalServerError)
		return
	}

	ret, _ := json.MarshalIndent(struct {
		Port *SapiProvisionedPorts `json:"port"`
	}{
		Port: sapiPort,
	}, "", "    ")

	rw.Write([]byte(ret))
}

func (this *SapiProvisionedPorts) Create(rw http.ResponseWriter, r *http.Request) {
	var err error

	sapiPort, err := getPort(r)
	if err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	if err = sapiPort.insert(); err != nil {
		if mysqlError, ok := err.(*mysql.MySQLError); ok {
			if mysqlError.Number != 1062 {
				HttpError(rw, "Database Error", err, http.StatusInternalServerError)
				return
			}
		}
	}

	rw.Write([]byte("OK"))
}

func (this *SapiProvisionedPorts) Update(rw http.ResponseWriter, r *http.Request) {
	var err error
	var sapiPort *SapiProvisionedPorts
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	sapiPort, err = getPort(r)
	if err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}
	if has, _ := new(SapiProvisionedPorts).search(id); !has {
		HttpError(rw, "Not Found", nil, http.StatusNotFound)
		return
	}

	_, err = sapiPort.update(id)
	if err != nil {
		HttpError(rw, "Database Error", err, http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("OK"))
}

func (this *SapiProvisionedPorts) Delete(rw http.ResponseWriter, r *http.Request) {
	var count int64
	var err error
	var sapiPort = &SapiProvisionedPorts{}
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	count, err = sapiPort.delete(id)
	if count <= 0 {
		HttpError(rw, "Not Found", nil, http.StatusNotFound)
		return
	}
	if err != nil {
		HttpError(rw, "Database Error", err, http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("OK"))
}

func (this *SapiProvisionedPorts) Mapper() ApiEndpoints {
	endpoints := make(ApiEndpoints)

	for _, methods := range Methods {
		endpoints[methods] = make(map[string]http.Handler)
		switch methods {
		case GET:
			endpoints[methods][port+"/{id}"] = http.HandlerFunc(this.Get)
		case DELETE:
			endpoints[methods][port+"/{id}"] = http.HandlerFunc(this.Delete)
		case UPDATE:
			endpoints[methods][port+"/{id}"] = http.HandlerFunc(this.Update)
		case POST:
			endpoints[methods][port+"/"] = http.HandlerFunc(this.Create)
		}
	}
	return endpoints
}

func (this *SapiProvisionedPorts) insert() error {
	_, err := DB().Insert(this)
	return err
}

func (this *SapiProvisionedPorts) search(id string) (bool, error) {
	this.PortId = id
	has, err := DB().Get(this)
	return has, err
}

func (this *SapiProvisionedPorts) delete(id string) (int64, error) {
	this.PortId = id
	c, err := DB().Delete(this)
	return c, err
}

func (this *SapiProvisionedPorts) update(id string) (int64, error) {
	this.PortId = id
	affected, err := DB().AllCols().Where("port_id=?", id).Update(this)
	return affected, err
}

func (this *SapiProvisionedPorts) truncate() error {
	_, err := DB().Where("port_id!=?", this.PortId).Delete(this)
	return err
}

func getFixips(b []byte) (ip, subnetId string, err error) {
	var fixips struct {
		Ips []map[string]string `json:"fixed_ips"`
	}

	if err := json.Unmarshal(b, &fixips); err != nil {
		return "", "", err
	}

	if len(fixips.Ips) >= 1 {
		return fixips.Ips[0]["ip_address"], fixips.Ips[0]["subnet_id"], nil
	}
	return "", "", nil
}

func getPort(r *http.Request) (*SapiProvisionedPorts, error) {
	var port *sjson.Json
	var sapiPort = &SapiProvisionedPorts{}

	post, err := sjson.NewFromReader(r.Body)
	if err != nil {
		return nil, err
	}

	port, ok := post.CheckGet("port")
	if !ok {
		return nil, ErrorNoPort
	}

	b, err := port.Encode()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, sapiPort); err != nil {
		return nil, err
	}
	sapiPort.IpAddress, sapiPort.SubnetId, err = getFixips(b)
	if err != nil {
		return nil, err
	}

	return sapiPort, nil
}

func init() {
	Regist(new(SapiProvisionedPorts))
}
