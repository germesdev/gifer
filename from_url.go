package main

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func resizeFromURLHandler(w http.ResponseWriter, req *http.Request) {
	log.Trace("Hit resize from URL ...")

	inBuffer, err := downloadSource(mux.Vars(req)["source"])
	if err != nil {
		log.Errorf("Download source error: %s", err)
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
