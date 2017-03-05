package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/malnick/holler"
)

var (
	VERSION  = "unset"
	REVISION = "unset"

	versionFlag = flag.Bool("version", false, "Print holler version")
)

func main() {
	flag.Parse()
	if *versionFlag {
		fmt.Printf("Holler Proxy\n    Version: %s\n    Revision: %s\n", VERSION, REVISION)
		os.Exit(0)
	}

	myHoller, err := holler.New()
	if err != nil {
		panic(err)
	}

	myHoller.Start()
}
