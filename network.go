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
	network    = "/network"
	ErrorNoNet = errors.New("Key network not found")
)

func (this *SapiProvisionedNets) Get(rw http.ResponseWriter, r *http.Request) {
	var has bool
	var err error
	var sapiNet = &SapiProvisionedNets{}
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	has, err = sapiNet.search(id)
	if !has {
		HttpError(rw, "Not Found", nil, http.StatusNotFound)
		return
	}
	if err != nil {
		HttpError(rw, "Database Error", err, http.StatusInternalServerError)
		return
	}

	ret, _ := json.MarshalIndent(struct {
		Network *SapiProvisionedNets `json:"network"`
	}{
		Network: sapiNet,
	}, "", "    ")

	rw.Write([]byte(ret))
}

func (this *SapiProvisionedNets) Create(rw http.ResponseWriter, r *http.Request) {
	var err error

	sapiNet, err := getNet(r)
	if err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	if err = sapiNet.insert(); err != nil {
		if mysqlError, ok := err.(*mysql.MySQLError); ok {
			if mysqlError.Number != 1062 {
				HttpError(rw, "Database Error", err, http.StatusInternalServerError)
				return
			}
		}
	}

	rw.Write([]byte("OK"))
}

func (this *SapiProvisionedNets) Update(rw http.ResponseWriter, r *http.Request) {
	var err error
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	sapiNet, err := getNet(r)
	if err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	affected, err := sapiNet.update(id)
	if affected <= 0 {
		HttpError(rw, "Not Found", nil, http.StatusNotFound)
		return
	}
	if err != nil {
		HttpError(rw, "Database Error", err, http.StatusInternalServerError)
		return
	}

	rw.Write([]byte("OK"))
}

func (this *SapiProvisionedNets) Delete(rw http.ResponseWriter, r *http.Request) {
	var count int64
	var err error
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}
	sapiNet := &SapiProvisionedNets{}
	count, err = sapiNet.delete(id)
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

func (this *SapiProvisionedNets) Mapper() ApiEndpoints {
	endpoints := make(ApiEndpoints)

	for _, methods := range Methods {
		endpoints[methods] = make(map[string]http.Handler)
		switch methods {
		case GET:
			endpoints[methods][network+"/{id}"] = http.HandlerFunc(this.Get)
		case DELETE:
			endpoints[methods][network+"/{id}"] = http.HandlerFunc(this.Delete)
		case UPDATE:
			endpoints[methods][network+"/{id}"] = http.HandlerFunc(this.Update)
		case POST:
			endpoints[methods][network+"/"] = http.HandlerFunc(this.Create)
		}
	}
	return endpoints
}

func (this *SapiProvisionedNets) insert() error {
	_, err := DB().Insert(this)
	return err
}

func (this *SapiProvisionedNets) search(id string) (bool, error) {
	this.NetworkId = id
	has, err := DB().Get(this)
	return has, err
}

func (this *SapiProvisionedNets) delete(id string) (int64, error) {
	this.NetworkId = id
	c, err := DB().Delete(this)
	return c, err
}

func (this *SapiProvisionedNets) update(id string) (int64, error) {
	this.NetworkId = id
	affected, err := DB().AllCols().Where("network_id=?", id).Update(this)
	return affected, err
}

func (this *SapiProvisionedNets) truncate() error {
	_, err := DB().Where("network_id!=?", this.NetworkId).Delete(this)
	return err
}

func getNet(r *http.Request) (*SapiProvisionedNets, error) {
	var network *sjson.Json
	var sapiNet = &SapiProvisionedNets{}

	post, err := sjson.NewFromReader(r.Body)
	if err != nil {
		return nil, err
	}
	network, ok := post.CheckGet("network")
	if !ok {
		return nil, ErrorNoNet
	}
	b, err := network.Encode()
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(b, sapiNet); err != nil {
		return nil, err
	}

	return sapiNet, nil
}

func init() {
	Regist(new(SapiProvisionedNets))
}
