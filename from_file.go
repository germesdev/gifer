package main

import (
	"bytes"
	"io"
	"mime/multipart"

	"log"
	"net/http"
)

func resizeFromFileHandler(w http.ResponseWriter, req *http.Request) {
	log.Println("[DEBUG] Hit resize from FILE ...")

	if err := req.ParseMultipartForm(50 * MB); nil != err {
		log.Printf("[ERROR] while parse: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer func() {
		err := req.MultipartForm.RemoveAll()
		if err != nil {
			log.Printf("[ERROR] Cant delete multipart error %s\n", err)
		}
	}()

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			log.Printf("Income file len: %d", hdr.Size)

			var (
				err       error
				infile    multipart.File
				outbuffer bytes.Buffer
			)

			if infile, err = hdr.Open(); err != nil {
				log.Printf("[ERROR] Handle open error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer infile.Close()

			_, err = io.Copy(&outbuffer, infile)

			if err != nil {
				log.Printf("[ERROR] Create Read Input error %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			processBuffer(w, req, &outbuffer)
			return
		}
	}
}
