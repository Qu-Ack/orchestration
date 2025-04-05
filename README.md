# Orchestration

This project is responsible for handling orchestration of containers. Uses Traefik as a reverse proxy to route to the correct container based on the subdomain.


## Code Base Structure


  ┌───────────────────────┐<br/>
  │     Transport Layer   │<br/>
  │ handles the incomin   │<br/>
  │   requests            │<br/>
  └───────────┬───────────┘<br/>
              │             <br/>
  ┌───────────▼───────────┐ <br/>
  │   Data Layer          │<br/>
  │ Manages data,         │<br/>
  │ create entities       │<br/>
  └───────────┬───────────┘<br/>
              │             <br/>
  ┌───────────▼───────────┐<br/>
  │  Logic Layer          │<br/>
  │   business logic      │<br/>
  └───────────┬───────────┘<br/>
              │             <br/>
  ┌───────────▼───────────┐ <br/>
  │ Persistence layer     │ <br/>
  └───────────────────────┘<br/>



## Steps To Set Up Locally


1. run the docker-compose.yml with `docker compose up -d` to run the traefik service
2. run the command 

`docker run -d   --name my-postgres   --network db-network   -e POSTGRES_USER=postgres   -e POSTGRES_PASSWORD=postgres   -e POSTGRES_DB=postgres   -v postgres-data:/var/lib/postgresql/data   -p 5433:5432   postgres:latest`

to run the containerized postgres-database

3.the datbase connection string `postgresql://postgres:postgres@my-postgres:5432/{YOUR DATABASE NAME}` 

