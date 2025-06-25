package deploy

import (
	"bytes"
	"fmt"
	"text/template"
)

const NodeDockerFileTemplate = `
FROM node:22-alpine
WORKDIR /app
COPY . ./
{{range .EnvVars}}
ENV {{.Key}}="{{.Value}}"
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

const ViteReactDockerFileTemplate = `
# Use Node.js base image
FROM node:22-alpine

# Set working directory
WORKDIR /app

# Copy package.json and lock file
COPY ./package*.json ./

# Set environment variables
{{range .EnvVars}}
ENV {{.Key}}="{{.Value}}"
{{end}}

# Install dependencies
RUN npm install --prefer-offline --no-audit --progress=false

# Copy the rest of the project files
COPY . ./

# Build the Vite app
RUN npm run build

# Install a lightweight HTTP server for static files
RUN npm install -g serve

# Expose the port Vite will run on
EXPOSE {{.Port}}

# Start the app using serve
CMD ["serve", "-s", "dist", "-l", "{{.Port}}"]
`

var ViteReactFile = template.Must(template.New("").Parse(ViteReactDockerFileTemplate))
var NodeDockerFile = template.Must(template.New("").Parse(NodeDockerFileTemplate))
var NextDockerFile = template.Must(template.New("").Parse(NextjsDockerFileTemplate))
var GoDockerFile = template.Must(template.New("").Parse(GoDockerFileTemplate))

func ExecuteNodeTemplate(data DockerTemplateData) string {

	for i := 0; i < len(data.EnvVars); i++ {
		fmt.Println(data.EnvVars[i].Value)
	}

	buf := bytes.Buffer{}
	if err := NodeDockerFile.Execute(&buf, data); err != nil {
		fmt.Println("Error", err)
		return ""
	}
	return buf.String()
}

func ExecuteViteReactTemplate(data DockerTemplateData) string {
	buf := bytes.Buffer{}
	if err := ViteReactFile.Execute(&buf, data); err != nil {
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
