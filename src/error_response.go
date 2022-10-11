package main

type ErrorResponse struct {
	Response
	Message string `json:"message"`
}
