package main

import (
    "log"

	"github.com/panjf2000/gnet/v2"
	"github.com/kxrxh/ddss/internal/db"
	"github.com/kxrxh/ddss/internal/server"
)

func main() {
    dgraph, err := db.NewDgraphClient("dgraph-alpha:9080")
    if err != nil {
        log.Fatalf("Failed to create Dgraph client: %v", err)
    }

    srv := server.NewServer("tcp://0.0.0.0:9000", dgraph)
    log.Fatal(gnet.Run(srv, srv.GetAddress(), gnet.WithMulticore(true)))
}