package main

type Role string

type A2ARequest struct {
	JsonRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Id      any    `json:"id"`
	Params  Params `json:"params"`
}

type A2AResponseSuccess struct {
	JsonRPC string  `json:"jsonrpc"`
	Id      any     `json:"id"`
	Result  Message `json:"result"`
}

type A2AResponseError struct {
	JsonRPC string       `json:"jsonrpc"`
	Id      any          `json:"id"`
	Error   JsonRPCError `json:"error"`
}

// Message represents communication between users and agents
type Message struct {
	Id    string `json:"messageId"`
	Role  Role   `json:"role"` // "user" or "agent"
	Parts []Part `json:"parts"`
	Kind  string `json:"kind"`
}

type DataPart struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

type Part struct {
	Kind string     `json:"kind"`
	Text string     `json:"text,omitempty"`
	Data []DataPart `json:"data,omitempty"`
}

type Params struct {
	Message Message `json:"message"`
}

type JsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
