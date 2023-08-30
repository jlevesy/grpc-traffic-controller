package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type bootstrapConfig struct {
	XDSServers []xdsServer `json:"xds_servers"`
	Node       node        `json:"node"`
}

type xdsServer struct {
	URI      string   `json:"server_uri"`
	Features []string `json:"server_features"`
	Creds    []cred   `json:"channel_creds"`
}

type cred struct {
	Type string `json:"type"`
}

type node struct {
	ID       string   `json:"id"`
	Locality Locality `json:"locality,omitempty"`
}

type Locality struct {
	Zone string `json:"zone"`
}

func main() {
	var (
		out       string
		serverURI string
		nodeID    string
		zone      string
	)

	flag.StringVar(&out, "out", "./bootstrap.json", "path to write the generated config")
	flag.StringVar(&serverURI, "server-uri", "", "uri of the xds server")
	flag.StringVar(&nodeID, "node-id", "", "id of the node")
	flag.StringVar(&zone, "zone", "", "current zone we're running on")
	flag.Parse()

	if serverURI == "" {
		log.Fatal("please provide a server-uri")
	}

	if nodeID == "" {
		nodeID, _ = os.Hostname()
	}

	cfg := bootstrapConfig{
		XDSServers: []xdsServer{
			{
				URI:      serverURI,
				Features: []string{"xds_v3"},
				Creds:    []cred{{Type: "insecure"}},
			},
		},
		Node: node{
			ID: nodeID,
			Locality: Locality{
				Zone: zone,
			},
		},
	}

	output, err := os.Create(out)
	if err != nil {
		log.Fatal(err)
	}
	defer output.Close()

	if err := json.NewEncoder(output).Encode(&cfg); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully wrote configuration at path:", out)
}
