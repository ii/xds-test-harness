package parser

type EnvoyConfig struct {
	Name string `yaml:"name"`
	Spec `yaml:"spec"`
}

type Spec struct {
	Listeners []Listener `yaml:"listeners"`
	Clusters  []Cluster  `yaml:"clusters"`
}

type Listener struct {
	Name    string  `yaml:"name"`
	Address string  `yaml:"address"`
	Port    uint32  `yaml:"port"`
	Routes  []Route `yaml:"routes"`
}

type Route struct {
	Name         string   `yaml:"name"`
	Prefix       string   `yaml:"prefix"`
	ClusterNames []string `yaml:"clusters"`
}

type Cluster struct {
	Name      string     `yaml:"name"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Address string `yaml:"address"`
	Port    uint32 `yaml:"port"`
}

type EnvoyNode struct {
	ID string `yaml:"id"`
}

type TestDiscoveryRequest struct {
	VersionInfo string `default:"" yaml:"version_info"`
	ResourceNames []string `default:[]string yaml:"resource_names"`
	TypeURL string `default:"" yaml:"type_url"`
	ResponseNonce string `default:"" yaml:"response_nonce"`
	Node EnvoyNode `yaml:"node"`
}
