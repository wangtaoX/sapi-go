package main

import (
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	neutron "github.com/wangtaoX/sapi"
	"os"
)

func main() {
	var usage = func() {
		fmt.Printf("Usage of dbsync:")
		flag.PrintDefaults()
	}
	create := flag.Bool("c", false, "create tables.")
	deleted := flag.Bool("d", false, "delete tables.")
	user := flag.String("u", "root", "database user name.")
	pass := flag.String("p", "root", "database user password.")
	host := flag.String("h", "localhost", "database hostname or ip address.")
	db := flag.String("db", "neutron", "sapi database name.")
	flag.Parse()

	address := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", *user, *pass, *host, *db)
	engine, err := xorm.NewEngine("mysql", address)
	if err != nil {
		fmt.Println(err)
		usage()
		os.Exit(0)
	}
	err = engine.Ping()
	if err != nil {
		fmt.Println(err)
		usage()
		os.Exit(0)
	}
	engine.ShowWarn = false

	if *deleted {
		engine.DropTables(
			new(neutron.SapiProvisionedNets),
			new(neutron.SapiProvisionedSubnets),
			new(neutron.SapiProvisionedPorts),
			new(neutron.SapiTorTunnels),
			new(neutron.SapiPortVlanMapping),
			new(neutron.SapiTor),
			new(neutron.SapiTorTunnels),
			new(neutron.SapiTorVsis),
			new(neutron.SapiVlanAllocations))
	}

	if *create {
		engine.Sync2(
			new(neutron.SapiProvisionedNets),
			new(neutron.SapiProvisionedSubnets),
			new(neutron.SapiProvisionedPorts),
			new(neutron.SapiTorTunnels),
			new(neutron.SapiPortVlanMapping),
			new(neutron.SapiTor),
			new(neutron.SapiTorTunnels),
			new(neutron.SapiTorVsis),
			new(neutron.SapiVlanAllocations))
	}
}
