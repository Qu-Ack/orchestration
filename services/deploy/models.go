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
	FilePath       string
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
