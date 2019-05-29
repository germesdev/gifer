package main

import (
	"bytes"
	"io/ioutil"
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

	for _, fheaders := range req.MultipartForm.File {
		for _, hdr := range fheaders {
			log.Printf("Income file len: %d", hdr.Size)

			var err error
			var infile multipart.File

			if infile, err = hdr.Open(); err != nil {
				log.Printf("[ERROR] Handle open error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer infile.Close()

			inbytes, err := ioutil.ReadAll(infile)
			if err != nil {
				log.Printf("[ERROR] Create Read Input error %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			processBuffer(w, req, bytes.NewBuffer(inbytes))
			return
		}
	}
}
