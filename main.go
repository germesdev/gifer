package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"github.com/cloudfoundry/bytefmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	MB       = 1 << 20
	Minutes5 = time.Minute * 5
)

type Cfg struct {
	WEBM_CRF  string `env:"WEBM_CRF" envDefault:"25"`
	WEBM_QMIN string `env:"WEBM_QMIN" envDefault:"8"`
}

var (
	config Cfg
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	err := env.Parse(&config)
	if err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}
	r := mux.NewRouter()

	log.Printf("Start gifer server on %s", port)

	r.SkipClean(true)

	r.HandleFunc("/healthz", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
	})).Methods("GET")

	r.HandleFunc(`/unsafe/{dimension:\d+x\d+}/{filters:filters:\w{3,}\(.*\)}/{source:.*}`, resizeFromURLHandler).Methods("GET")
	r.HandleFunc(`/unsafe/{dimension:\d+x\d+}/{filters:filters:\w{3,}\(.*\)}/{source:.*}`, resizeFromFileHandler).Methods("POST")
	r.HandleFunc(`/unsafe/{dimension:\d+x\d+}/{filters:filters:\w{3,}\(.*\)}`, resizeFromFileHandler).Methods("POST")
	r.HandleFunc("/version", versionHandler())
	srv := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: Minutes5, // big file
		ReadTimeout:  Minutes5, // big file
	}

	go func() {
		log.Info(srv.ListenAndServe())
	}()
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown // Block until signal received
	log.Info("Gifer prepares for shutdown ...")

	ctx, cancel := context.WithTimeout(context.Background(), Minutes5)
	defer cancel()
	srv.Shutdown(ctx)

	log.Info("Gifer goes shutdown ...")
}

func parseParams(req *http.Request) (string, string, error) {
	var (
		dimension string
		format    string
		err       error
	)
	dimension = parseDimension(mux.Vars(req)["dimension"])
	if format, err = parseFormat(mux.Vars(req)["filters"]); err != nil {
		log.Errorf("Bad format: %s %v", err, mux.Vars(req))
		return "", "", err
	}
	return dimension, format, nil
}

func parseFormat(format string) (string, error) {
	switch format {
	case "filters:gifv(mp4)":
		return "mp4", nil
	case "filters:gifv(webm)":
		return "webm", nil
	case "filters:gifv(webp)":
		return "webp", nil
	case "filters:gifv(gif)":
		return "gif", nil
	case "filters:format(jpg)", "filters:format(jpeg)":
		return "jpg", nil
	default:
		return "", fmt.Errorf("bad format")
	}
}

func parseDimension(dim string) string {
	dimensions := strings.Split(dim, "x")
	widht, height := dimensions[0], dimensions[1]
	w, _ := strconv.ParseInt(widht, 10, 64)
	h, _ := strconv.ParseInt(height, 10, 64)
	var dimension string
	switch {
	case w > 0 && h == 0:
		dimension = fmt.Sprintf("scale=trunc(%d/2)*2:-2", w)
	case h > 0 && w == 0:
		dimension = fmt.Sprintf("scale=-2:trunc(%d/2)*2", h)
	case w > 0 && h > 0:
		dimension = fmt.Sprintf("scale=w=%v:h=%v:force_original_aspect_ratio=increase,crop=%v:%v", w, h, w, h)
	default:
		dimension = "scale=trunc(iw/2)*2:trunc(ih/2)*2"
	}
	return dimension
}

func versionHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Trace("Hit version")
		cmd := exec.Command("ffmpeg", "--help")

		var out bytes.Buffer
		cmd.Stdout = &out

		var errout bytes.Buffer
		cmd.Stderr = &errout

		err := cmd.Run()

		if err != nil {
			log.Errorf("FFmpeg output: %v, %v, %v\n", err, out.String(), errout.String())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Errorf("FFmpeg output: %v", out.String())

		w.WriteHeader(http.StatusOK)
	})
}

func processBuffer(w http.ResponseWriter, req *http.Request, inBuffer *bytes.Buffer) {
	dimension, format, err := parseParams(req)
	if err != nil {
		log.Errorf("parse params error %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Tracef("Process file -> Dimension: %s", dimension)
	sourceSize := inBuffer.Len()

	var xfilename, contentType string
	switch format {
	case "webm", "mp4":
		xfilename = "video." + format
		contentType = "video/" + format
	case "jpg", "webp", "gif":
		xfilename = "image." + format
		contentType = "image/" + format
	}

	resBuffer, err := convert(inBuffer, format, dimension)
	if err != nil {
		log.Errorf("Convert error %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	imageLen := resBuffer.Len()

	log.WithFields(log.Fields{
		"before": bytefmt.ByteSize(uint64(sourceSize)),
		"after":  bytefmt.ByteSize(uint64(imageLen)),
		"format": format,
		"coff":   float64(uint((float64(imageLen)/float64(sourceSize))*100)) / 100,
	}).Info("Convert result")

	w.Header().Set("X-Filename", xfilename)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(imageLen))
	w.Header().Set("Accept-Ranges", "bytes")

	http.ServeContent(w, req, xfilename, time.Time{}, bytes.NewReader(resBuffer.Bytes()))
}
