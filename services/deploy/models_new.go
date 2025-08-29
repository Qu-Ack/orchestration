package deploy

type DEP int

const (
	DEP_DOCKER_COMPOSE DEP = iota
	DEP_DOCKER_FILE
	DEP_NEXT
	DEP_VITE_HTML
	DEP_VITE_REACT
	DEP_REACT
	DEP_GO
	DEP_NODE
	DEP_NEXT_PRISMA
)

type Status int

const (
	STATUS_PENDING Status = iota
	STATUS_FAILED
	STATUS_SUCCESS
)

type User_Deployment int

const (
	USER_DEP_DOCKER_COMPOSE User_Deployment = iota
	USER_DEP_DOCKER
	USER_DEP_GO
	USER_DEP_NEXT
	USER_DEP_PRISMA_NEXT
	USER_DEP_VITE_HTML
	USER_DEP_VITE_REACT
	USER_DEP_NODE
)

func (u User_Deployment) String() string {
	switch u {
	case USER_DEP_DOCKER_COMPOSE:
		return "docker_compose"
	case USER_DEP_DOCKER:
		return "docker"
	case USER_DEP_GO:
		return "go"
	case USER_DEP_NEXT:
		return "next"
	case USER_DEP_PRISMA_NEXT:
		return "prisma_next"
	case USER_DEP_VITE_HTML:
		return "vite_html"
	case USER_DEP_VITE_REACT:
		return "vite_react"
	case USER_DEP_NODE:
		return "node"
	default:
		return "unknown"
	}
}

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
	SER_VITE_HTML
	SER_VITE_REACT
)

type Service struct {
	ServiceID      string   `json:"service_id"`
	CodeRepo       string   `json:"code_repo"`
	CodeRepoBranch string   `json:"code_repo_branch"`
	ClonePath      string   `json:"clone_path"`
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
	UserID   string    `json:"userid"`
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Type     DEP       `json:"type"`
	Status   Status    `json:"status"`
	Services []Service `json:"services"`
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type User_Deployment_Request struct {
	UserID     string          `json:"userid"`
	Repo       string          `json:"repo"`
	RepoBranch string          `json:"branch"`
	Type       User_Deployment `json:"type"`
	FilePath   string          `json:"file_path"`
	Name       string          `json:"name"`
}

type User_Validation_response struct {
	Status     string          `json:"status"`
	Type       string          `json:"type"`
	Identified map[string]any  `json:"identified"`
	Deployment *Deployment_New `json:"deployment"`
}
