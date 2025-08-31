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
	ServiceID       string   `json:"service_id"`
	CodeRepo        string   `json:"code_repo"`
	CodeRepoBranch  string   `json:"code_repo_branch"`
	ClonePath       string   `json:"clone_path"`
	Name            string   `json:"name"`
	ServiceType     SERVICE  `json:"service_type"`
	ServiceUserType string   `json:"service_user_type"`
	Port            int      `json:"port"`
	Domain          string   `json:"domain"`
	EnvVars         []EnvVar `json:"envs"`
	image           string
	containerId     string
	logFilePath     string
}

var serviceTypeMap = map[SERVICE]string{
	SER_NEXT:        "next",
	SER_NEXT_PRISMA: "next_prisma",
	SER_NODE:        "node",
	SER_GO:          "go",
	SER_DOCKER:      "docker",
	SER_POSTGRES:    "postgres",
	SER_REDIS:       "redis",
	SER_NGINX:       "nginx",
	SER_VITE_HTML:   "vite_html",
	SER_VITE_REACT:  "vite_react",
}

type Deployment_New struct {
	UserID   string
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Type     DEP       `json:"type"`
	UserType string    `json:"user_type"`
	Status   Status    `json:"status"`
	Services []Service `json:"services"`
}

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var deploymentTypeMapServer = map[DEP]string{
	DEP_DOCKER_COMPOSE: "docker_compose",
	DEP_DOCKER_FILE:    "docker_file",
	DEP_GO:             "golang",
	DEP_NEXT:           "next",
	DEP_NEXT_PRISMA:    "next_prisma",
	DEP_VITE_HTML:      "vite_html",
	DEP_VITE_REACT:     "vite_react",
	DEP_NODE:           "node",
}

var deploymentTypeMap = map[string]User_Deployment{
	"docker_compose": USER_DEP_DOCKER_COMPOSE,
	"docker_file":    USER_DEP_DOCKER,
	"golang":         USER_DEP_GO,
	"next":           USER_DEP_NEXT,
	"next_prisma":    USER_DEP_PRISMA_NEXT,
	"vite_html":      USER_DEP_VITE_HTML,
	"vite_react":     USER_DEP_VITE_REACT,
	"node":           USER_DEP_NODE,
}

type User_Deployment_Request struct {
	UserID     string
	Repo       string `json:"repo"`
	RepoBranch string `json:"branch"`
	Type       string `json:"type"`
	Name       string `json:"name"`
}

type User_Validation_response struct {
	Status     string          `json:"status"`
	Type       string          `json:"type"`
	Identified map[string]any  `json:"identified"`
	Deployment *Deployment_New `json:"deployment"`
}
