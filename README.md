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
