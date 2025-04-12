package deploy

const (
	node   = iota
	next   = iota
	react  = iota
	golang = iota
)

type DockerTemplateData struct {
	RepoIdentifier string
	Port           int
	EnvVars        []EnvVar
}

type Deployment struct {
	ID          string
	SubDomain   string
	CloneUrl    string
	Branch      string
	RepoName    string
	ProjectPath string
	ProjectType int
	EnvVars     []EnvVar
	Port        int
}

type EnvVar struct {
	Key   string
	Value string
}

type ContainerStats struct {
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage int64   `json:"memoryUsage"`
	MemoryLimit int64   `json:"memoryLimit"`
	NetworkRx   int64   `json:"networkRx"`
	NetworkTx   int64   `json:"networkTx"`
	Status      string  `json:"status"`
}
