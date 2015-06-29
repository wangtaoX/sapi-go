package sapi

type SapiProvisionedNets struct {
	NetworkId        string `json:"id" xorm:"pk varchar(36)"`
	TenantId         string `json:"tenant_id"`
	SegmentationType string `json:"provider:network_type" xorm:"varchar(36)"`
	SegmentationId   int    `json:"provider:segmentation_id"`
	AdminStateUp     bool   `json:"admin_state_up"`
	Shared           bool   `json:"shared"`
}

type SapiProvisionedSubnets struct {
	SubnetId   string `json:"id" xorm:"pk varchar(36)"`
	TenantId   string `json:"tenant_id"`
	NetworkId  string `json:"network_id" xorm:"varchar(36)"`
	Shared     bool   `json:"shared"`
	EnableDhcp bool   `json:"enable_dhcp"`
}

type SapiProvisionedPorts struct {
	PortId        string `json:"id" xorm:"pk varchar(36)"`
	TenantId      string `json:"tenant_id"`
	NetworkId     string `json:"network_id" xorm:"varchar(36)"`
	SubnetId      string `json:"subnet_id" xorm:"varchar(36)"`
	DeviceId      string `json:"device_id"`
	DeviceOwner   string `json:"device_owner" xorm:"varchar(40)"`
	Status        string `json:"status" xorm:"varchar(40)"`
	AdminStateUp  bool   `json:"admin_state_up"`
	BindingHostId string `json:"binding_host_id" xorm:"varchar(40)"`
	IpAddress     string `json:"ip_address"`
	MacAddress    string `json:"mac_address"`
}

type SapiPortVlanMapping struct {
	Id        int    `xorm:"pk autoincr"`
	NetworkId string `xorm:"varchar(36)"`
	TorIp     string `xorm:"varchar(45)"`
	VlanId    int
	Index     int
}

func (this *SapiPortVlanMapping) insert() error {
	_, err := DB().Insert(this)
	return err
}

func (this *SapiPortVlanMapping) search(id string) (bool, error) {
	this.NetworkId = id
	has, err := DB().Get(this)
	return has, err
}

func (this *SapiPortVlanMapping) count() int64 {
	total, _ := DB().Where("network_id=? AND tor_ip=?", this.NetworkId, this.TorIp).Count(new(SapiPortVlanMapping))
	return total
}

func (this *SapiPortVlanMapping) delete() (int64, error) {
	c, err := DB().Delete(this)
	return c, err
}

type SapiTor struct {
	TorIp       string `xorm:"pk varchar(45)"`
	TunnelSrcIp string `xorm:"varchar(45)"`
	Type        string `xorm:"varchar(45)"`
}

func (this *SapiTor) insert() error {
	_, err := DB().Insert(this)
	return err
}

type SapiTorTunnels struct {
	Id       int    `xorm:"pk autoincr"`
	TorIp    string `xorm:"varchar(45)"`
	TunnelId int
	DstAddr  string `xorm:"varchar(45)"`
}

func (this *SapiTorTunnels) insert() error {
	_, err := DB().Insert(this)
	return err
}

func (this *SapiTorTunnels) delete() (int64, error) {
	c, err := DB().Delete(this)
	return c, err
}

type SapiTorVsis struct {
	Id    int    `xorm:"pk autoincr"`
	TorIp string `xorm:"varchar(45)"`
	Vxlan int
}

func (this *SapiTorVsis) insert() error {
	_, err := DB().Insert(this)
	return err
}

func (this *SapiTorVsis) delete() (int64, error) {
	c, err := DB().Delete(this)
	return c, err
}

type SapiVlanAllocations struct {
	Id        int `xorm:"pk autoincr"`
	NetworkId string
	TorIp     string `xorm:"varchar(45)"`
	VlanId    int
	Allocated bool
	Shared    bool
}

func (this *SapiVlanAllocations) insert() error {
	_, err := DB().Insert(this)
	return err
}

func (this *SapiVlanAllocations) delete() (int64, error) {
	c, err := DB().Delete(this)
	return c, err
}

func SelectAllVlanAlloctions(every *[]*SapiVlanAllocations) error {
	if err := DB().Find(every); err != nil {
		return err
	}
	return nil
}

func SelectAllTors(every *[]*SapiTor) error {
	if err := DB().Find(every); err != nil {
		return err
	}
	return nil
}

func SelectAllVsiByTor(torIp string, every *[]*SapiTorVsis) error {
	if err := DB().Where("tor_ip=?", torIp).Find(every); err != nil {
		return err
	}
	return nil
}

func SelectAllTunnelByTor(torIp string, every *[]*SapiTorTunnels) error {
	if err := DB().Where("tor_ip=?", torIp).Find(every); err != nil {
		return err
	}
	return nil
}
