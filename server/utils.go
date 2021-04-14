package main

import (
	"fmt"
	"net/http"
)

func SendResponse(resp http.ResponseWriter, errorCode int, msg string) {
	ret := fmt.Sprintf("{\"code\" : %d, \"message\": \"%s\"}", errorCode, msg)
	resp.Write([]byte(ret))
}
