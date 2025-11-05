package main

// Role indicates the sender type ("user" or "agent").
type Role string

// A2ARequest represents an incoming JSON-RPC request.
type A2ARequest struct {
	JsonRPC string `json:"jsonrpc"` // Must be "2.0"
	Method  string `json:"method"`  // Method name
	Id      any    `json:"id"`      // Request ID
	Params  Params `json:"params"`  // Message payload
}

// A2AResponseSuccess is a successful JSON-RPC response.
type A2AResponseSuccess struct {
	JsonRPC string   `json:"jsonrpc"`
	Id      any      `json:"id"`
	Result  Messages `json:"result"`
}

// A2AResponseError is a JSON-RPC error response.
type A2AResponseError struct {
	JsonRPC string       `json:"jsonrpc"`
	Id      any          `json:"id"`
	Error   JsonRPCError `json:"error"`
}

type Result struct {
	Id        any        `json:"id"`
	ContextId any        `json:"contextId"`
	Status    Status     `json:"status"`
	Artifacts []Artifact `json:"artifacts"`
	History   any        `json:"history"`
	Kind      string     `json:"kind"`
}

type Status struct {
	State     string  `json:"state"`
	TimeStamp string  `json:"timestamp"`
	Message   Message `json:"message"`
}

type Artifact struct {
	ArtifactId string             `json:"artifactId"`
	Name       string             `json:"name"`
	Parts      []ArtifactDataPart `json:"parts"`
}

// Message represents a message between a user and agent.
type Message struct {
	Id    string `json:"messageId"`
	Role  Role   `json:"role"`
	Parts []Part `json:"parts"`
	Kind  string `json:"kind"`
}

// DataPart holds individual data items inside a Part.
type DataPart struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

type ArtifactDataPart struct {
	Kind string `json:"kind"`
	Data any    `json:"data"`
}

// Part represents a message component (text or structured data).
type Part struct {
	Kind string     `json:"kind"`
	Text string     `json:"text,omitempty"`
	Data []DataPart `json:"data,omitempty"`
}

// Params wraps the message in an A2A request.
type Params struct {
	Message Message `json:"message"`
}

// JsonRPCError defines a JSON-RPC error object.
type JsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
