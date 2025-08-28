package deploy

type DEP int

const (
	DEP_DOCKER_COMPOSE DEP = iota
	DEP_DOCKER_FILE
	DEP_NEXT
	DEP_REACT
	DEP_GO
	DEP_NODE
	DEP_NEXT_PRISMA
)

type Status int

const (
	STATUS_PENDING = iota
	STATUS_FAILED
	STATUS_SUCCESS
)

type SERVICE int

const (
	SER_NEXT SERVICE = iota
	SER_NEXT_PRISMA
	SER_NODE
	SER_GO
	SER_DOCKER
	SER_POSTGRES
	SER_REDIS
	SER_NGINX
)

type Service struct {
	CodeRepo       string   `json:"code_repo"`
	CodeRepoBranch string   `json:"code_repo_branch"`
	Name           string   `json:"name"`
	ServiceType    SERVICE  `json:"service_type"`
	Port           int      `json:"port"`
	Domain         string   `json:"domain"`
	EnvVars        []EnvVar `json:"envs"`
	image          string
	containerId    string
	logFilePath    string
}

type Deployment_New struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Type     DEP       `json:"type"`
	Status   Status    `json:"status"`
	Services []Service `json:"services"`
}

type EnvVar_New struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type User_Deployment_Request struct {
}
