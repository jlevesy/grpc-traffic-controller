package bootstrap

type BootstrapConfig struct {
	XDSServers []XDSServer `json:"xds_servers"`
	Node       Node        `json:"node"`
}

type XDSServer struct {
	URI      string   `json:"server_uri"`
	Features []string `json:"server_features"`
	Creds    []Cred   `json:"channel_creds"`
}

type Cred struct {
	Type string `json:"type"`
}

type Node struct {
	ID       string   `json:"id"`
	Locality Locality `json:"locality,omitempty"`
}

type Locality struct {
	Zone string `json:"zone"`
}
