package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func resizeFromURLHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("[DEBUG] Hit resize from URL ...")

	sum := getSum(fmt.Sprintf("%s", req.URL))

	if lock.locked(sum) {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	lock.lock(sum)
	defer lock.unlock(sum)

	inBuffer, err := downloadSource(mux.Vars(req)["source"])
	if err != nil {
		log.Printf("[ERROR] Download source error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	processBuffer(w, req, inBuffer)
}

func downloadSource(sourceURL string) (*bytes.Buffer, error) {
	resp, err := http.Get(sourceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	res := &bytes.Buffer{}
	_, err = io.Copy(res, resp.Body)
	return res, err
}
