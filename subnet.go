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
	subnet        = "/subnet"
	ErrorNoSubnet = errors.New("Key subnet not founded")
)

func (this *SapiProvisionedSubnets) Get(rw http.ResponseWriter, r *http.Request) {
	var has bool
	var err error
	var sapiSubnet = &SapiProvisionedSubnets{}
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	has, err = sapiSubnet.serach(id)
	if !has {
		HttpError(rw, "Not Found", nil, http.StatusNotFound)
		return
	}
	if err != nil {
		HttpError(rw, "Database Error", err, http.StatusInternalServerError)
		return
	}

	ret, _ := json.MarshalIndent(struct {
		Subnet *SapiProvisionedSubnets `json:"subnet"`
	}{
		Subnet: sapiSubnet,
	}, "", "    ")

	rw.Write([]byte(ret))
}

func (this *SapiProvisionedSubnets) Create(rw http.ResponseWriter, r *http.Request) {
	var err error

	sapiSubnet, err := getSubnet(r)
	if err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	if err = sapiSubnet.insert(); err != nil {
		if mysqlError, ok := err.(*mysql.MySQLError); ok {
			if mysqlError.Number != 1062 {
				HttpError(rw, "Database Error", err, http.StatusInternalServerError)
				return
			}
		}
	}

	rw.Write([]byte("OK"))
}

func (this *SapiProvisionedSubnets) Update(rw http.ResponseWriter, r *http.Request) {
	var err error
	var sapiSubnet *SapiProvisionedSubnets
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	sapiSubnet, err = getSubnet(r)
	if err != nil {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	affected, err := sapiSubnet.update(id)
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

func (this *SapiProvisionedSubnets) Delete(rw http.ResponseWriter, r *http.Request) {
	var count int64
	var err error
	var sapiSubnet = &SapiProvisionedSubnets{}
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok {
		HttpError(rw, "Bad Request", nil, http.StatusBadRequest)
		return
	}

	count, err = sapiSubnet.delete(id)
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

func (this *SapiProvisionedSubnets) Mapper() ApiEndpoints {
	endpoints := make(ApiEndpoints)

	for _, methods := range Methods {
		endpoints[methods] = make(map[string]http.Handler)
		switch methods {
		case GET:
			endpoints[methods][subnet+"/{id}"] = http.HandlerFunc(this.Get)
		case DELETE:
			endpoints[methods][subnet+"/{id}"] = http.HandlerFunc(this.Delete)
		case UPDATE:
			endpoints[methods][subnet+"/{id}"] = http.HandlerFunc(this.Update)
		case POST:
			endpoints[methods][subnet+"/"] = http.HandlerFunc(this.Create)
		}
	}
	return endpoints
}

func (this *SapiProvisionedSubnets) insert() error {
	_, err := DB().Insert(this)
	return err
}

func (this *SapiProvisionedSubnets) serach(id string) (bool, error) {
	this.SubnetId = id
	has, err := DB().Get(this)
	return has, err
}

func (this *SapiProvisionedSubnets) delete(id string) (int64, error) {
	this.SubnetId = id
	c, err := DB().Delete(this)
	return c, err
}

func (this *SapiProvisionedSubnets) update(id string) (int64, error) {
	this.SubnetId = id
	affected, err := DB().AllCols().Where("subnet_id=?", id).Update(this)
	return affected, err
}

func (this *SapiProvisionedSubnets) truncate() error {
	_, err := DB().Where("subnet_id!=?", this.SubnetId).Delete(this)
	return err
}

func getSubnet(r *http.Request) (*SapiProvisionedSubnets, error) {
	var port *sjson.Json
	var sapiSubnet = &SapiProvisionedSubnets{}

	post, err := sjson.NewFromReader(r.Body)
	if err != nil {
		return nil, err
	}

	port, ok := post.CheckGet("subnet")
	if !ok {
		return nil, ErrorNoPort
	}

	b, err := port.Encode()
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, sapiSubnet); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return sapiSubnet, nil
}

func init() {
	Regist(new(SapiProvisionedSubnets))
}
