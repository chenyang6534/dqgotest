// testclient project testclient.go

package main

import (
	"dq/log"
	"dq/app"
	//"os"
	//"os/signal"
	//"io"
//	"fmt"
//	"net"
	//"dq/rpc"
	//"time"
	//"net/rpc/jsonrpc"
	
)

func main() {

	app := new(app.DefaultApp)
	app.Run()
	log.Info("dq over!")
}
