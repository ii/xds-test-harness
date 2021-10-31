package parser

type ConnectTimeout struct {
	Seconds int64 `yaml:"seconds"`
}

type Endpoint struct {
	Name    string `yaml:"name"`
	Cluster string `yaml:"cluster"`
	Address string `yaml:"address"`
}

type Cluster struct {
	Name           string         `yaml:"name"`
	Endpoints      []Endpoint     `yaml:"endpoints"`
	ConnectTimeout ConnectTimeout `yaml:"connect_timeout"`
}

type Route struct {
	Name         string   `yaml:"name"`
	Prefix       string   `yaml:"prefix"`
	ClusterNames []string `yaml:"clusters"`
}

type Listener struct {
	Name    string  `yaml:"name"`
	Address string  `yaml:"address"`
	Port    uint32  `yaml:"port"`
	Routes  []Route `yaml:"routes"`
}

type Runtime struct {
	Name string `yaml:"name"`
}

type Secret struct {
	Name string `yaml:"name"`
}

type Resources struct {
	Endpoints []Endpoint `yaml:"endpoints"`
	Clusters  []Cluster  `yaml:"clusters"`
	Routes    []Route    `yaml:"routes"`
	Listeners []Listener `yaml:"listeners"`
	Runtimes  []Runtime  `yaml:"runtimes"`
	Secrets   []Secret   `yaml:"secret"`
}

type Snapshot struct {
	Node      string    `yaml:"node"`
	Version   string    `yaml:"version"`
	Resources Resources `yaml:"resources"`
}

type DiscoveryResponse struct {
	VersionInfo string    `default:"" yaml:"version_info"`
	TypeURL     string    `yaml:"type_url"`
	Resources   []Cluster `yaml:"resources"` //hack for now, should be any type of resource
	Nonce       string    `default:"" yaml:"nonce"`
}

type SimpleResponse struct {
	// Only the info used for validating our tests
	Version   string
	Resources []string
	Nonce     string
}
