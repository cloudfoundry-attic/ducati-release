package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type Handler struct{}

func (h *Handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, "/proxy/") {
		resp.Write([]byte("hello, this is proxy"))
		return
	}

	destination := strings.TrimPrefix(req.URL.Path, "/proxy/")
	destination = "http://" + destination

	getResp, err := http.Get(destination)
	if err != nil {
		log.Fatal("request failed")
	}
	defer getResp.Body.Close()

	readBytes, err := ioutil.ReadAll(getResp.Body)
	if err != nil {
		log.Fatal("failed reading response")
	}

	resp.Write(readBytes)
}

func main() {
	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		log.Fatal("missing required env var PORT")
	}

	handler := &Handler{}
	http.ListenAndServe("0.0.0.0:"+listenPort, handler)

}
