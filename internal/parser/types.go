package parser

type ConnectTimeout struct {
	seconds int32 `yaml:"seconds"`
}

type Endpoint struct {
	Address string `yaml:"address"`
	Port    uint32 `yaml:"port"`
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

type InitConfig struct {
	Adapter       string `yaml:"adapter"`
	Target        string `yaml:"target"`
	TargetAdapter string `yaml:"target_adapter"`
}

type TestDiscoveryRequest struct {
	VersionInfo   string   `default:"" yaml:"version_info"`
	ResourceNames []string `default:[]string yaml:"resource_names"`
	TypeURL       string   `default:"" yaml:"type_url"`
	ResponseNonce string   `default:"" yaml:"response_nonce"`
	// Node          EnvoyNode `yaml:"node"`
}
