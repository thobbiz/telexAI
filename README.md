## Telex History AI

telexAI is a History Teaching AI agent built with Google Gemini, Go and the Gin Web Framework . It is integrated with a Telex AI coworker using A2A protocol and serves as the core engine of the worker.

## Features:
- Give randoms historical facts.
- Handles custom user request

## Stack and Tools:
- Golang
- Google gemini
- Gin Web Framework
- net/http

## Get Started
### Prerequisites
- G0 1.21+
- Internet Connection
- Modules:
  ```bash
  go get github.com/gin-gonic/gin

## Usage
- Start the server:
  ```bash
  go run main.go
- Send a A2A request:
  ```bash
  curl http://localhost:8080/history_agent
