# Orchestration

This project is responsible for handling orchestration of containers. Uses Traefik as a reverse proxy to route to the correct container based on the subdomain.


## Code Base Structure


  ┌───────────────────────┐
  │     Transport Layer   │
  │ handles the incomin   │
  │   requests            │
  └───────────┬───────────┘
              │
  ┌───────────▼───────────┐
  │   Data Layer          │
  │ Manages data,         │
  │ create entities       │
  └───────────┬───────────┘
              │
  ┌───────────▼───────────┐
  │  Logic Layer          │
  │   business logic      │
  └───────────┬───────────┘
              │
  ┌───────────▼───────────┐
  │ Persistence layer     │
  └───────────────────────┘
