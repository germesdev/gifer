package main

import (
	"bytes"
	"io"
	"mime/multipart"

	"net/http"

	log "github.com/sirupsen/logrus"
)

func resizeFromFileHandler(w http.ResponseWriter, req *http.Request) {
	log.Trace("Hit resize from FILE")

	if err := req.ParseMultipartForm(50 * MB); nil != err {
		log.Error("while parse", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		err := req.MultipartForm.RemoveAll()
		if err != nil {
			log.Error("Cant delete multipart error", err)
		}
	}()

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			log.Tracef("Income file len: %d", hdr.Size)

			var (
				err       error
				infile    multipart.File
				outbuffer bytes.Buffer
			)

			if infile, err = hdr.Open(); err != nil {
				log.Errorf("Handle open error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				continue
			}
			defer infile.Close()

			_, err = io.Copy(&outbuffer, infile)

			if err != nil {
				log.Errorf("Create Read Input error %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				continue
			}

			processBuffer(w, req, &outbuffer)
			return
		}
	}
}
