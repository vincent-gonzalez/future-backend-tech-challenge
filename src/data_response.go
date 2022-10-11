package main

type DataResponse struct {
	Response
	Data interface{} `json:"data"`
}
