package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/codegangsta/negroni"
	"github.com/wangtaoX/sapi"
	"github.com/wangtaoX/sapi/middleware"
)

func main() {
	done := make(chan struct{})
	defer close(done)
	var usage = func() {
		fmt.Println("Usage of sapi server:")
		flag.PrintDefaults()
	}

	db := flag.String("db", "127.0.0.1", "default database for sapi")
	logFile := flag.String("log", "/var/log/sapi/sapi.log", "log file for sapi")
	flag.Parse()

	if err := sapi.InitLog(*logFile); err != nil {
		fmt.Printf("Init log error: %s\n\n", err)
		usage()
		os.Exit(1)
	}
	if err := sapi.InitDb("sapi", "sapi", *db, "sapi", done); err != nil {
		fmt.Printf("Init database error: %s\n\n", err)
		usage()
		os.Exit(1)
	}

	sapi.InitInmemoryData(sapi.GetTors())
	sapi.GoTopology(done)
	r := sapi.Router()
	server := negroni.New(
		middleware.NewBasicAuth(),
		middleware.NewLog("sapi"))

	server.UseHandler(r)
	server.Run(":8080")
}
