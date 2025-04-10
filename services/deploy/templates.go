package deploy

import (
	"bytes"
	"fmt"
	"html/template"
)

const NodeDockerFileTemplate = `
FROM node:22-alpine
WORKDIR /app
COPY . ./
{{range .EnvVars}}
ENV {{.Key}}={{.Value}}
{{end}}
RUN npm install
EXPOSE {{.Port}}
CMD ["node", "./src/index.js"]
`

const GoDockerFileTemplate = `
FROM golang:1.24-bookworm
WORKDIR /app
COPY . ./
{{range .EnvVars}}
ENV {{.Key}}={{.Value}}
{{end}}
RUN go build -o app ./cmd
EXPOSE {{.Port}}
CMD ["./app"]
`

const NextjsDockerFileTemplate = `
FROM node:22-alpine
WORKDIR /app
# Copy package files
COPY ./package*.json ./
{{range .EnvVars}}
ENV {{.Key}}="{{.Value}}"
{{end}}
# Install dependencies
RUN npm install --prefer-offline --no-audit --progress=false
# Copy project files
COPY . ./
# Generate Prisma client
RUN npx prisma generate
# Build application
RUN npm run build
# Expose the Next.js port
EXPOSE {{.Port}}
# Start the application
CMD ["npm", "start"]`

var NodeDockerFile = template.Must(template.New("").Parse(NodeDockerFileTemplate))
var NextDockerFile = template.Must(template.New("").Parse(NextjsDockerFileTemplate))
var GoDockerFile = template.Must(template.New("").Parse(GoDockerFileTemplate))

func ExecuteNodeTemplate(data DockerTemplateData) string {
	buf := bytes.Buffer{}
	if err := NodeDockerFile.Execute(&buf, data); err != nil {
		fmt.Println("Error", err)
		return ""
	}
	return buf.String()
}

func ExecuteGoTemplate(data DockerTemplateData) string {
	buf := bytes.Buffer{}
	if err := GoDockerFile.Execute(&buf, data); err != nil {
		fmt.Println("Error", err)
		return ""
	}
	return buf.String()
}

func ExecuteNextTemplate(data DockerTemplateData) string {
	buf := bytes.Buffer{}
	if err := NextDockerFile.Execute(&buf, data); err != nil {
		fmt.Println("Error", err)
		return ""
	}
	return buf.String()
}
