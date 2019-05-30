package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func convert(src *bytes.Buffer, format string, dimensions string) (*bytes.Buffer, error) {
	inputfile, err := ioutil.TempFile("", "*")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Create Input error %v", err)
	}
	defer os.Remove(inputfile.Name())

	_, err = io.Copy(inputfile, src)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Copy inputfile error %v", err)
	}

	outfile, err := ioutil.TempFile("", "res*")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Create Outfile error %v", err)
	}
	defer os.Remove(outfile.Name())

	cmd := exec.Command("ffmpeg",
		"-an", // disable audio
		"-y",  // overwrite
		// "-trans_color", "ffffff", // TODO read from input
		"-i", inputfile.Name(), // set input
		"-vf", dimensions,
		// "-pix_fmt", "yuv420p",
		// "-movflags", "frag_keyframe",
		"-movflags", "faststart",
		// "-qmin", "10", // the minimum quantizer (default 4, range 0–63), lower - better quality --- VP9 only
		// "-qmax", "42", // the maximum quantizer (default 63, range qmin–63) higher - lower quality --- VP9 only
		// By default the CRF value can be from 4–63, and 10 is a good starting point. Lower values mean better quality.
		// "-preset", "medium", // quality preset
		// "-maxrate", "1M",
		// "-minrate", "800K",
		"-qmin", "0",
		"-qmax", "50",
		"-crf", "25",
		// "-profile:v", "baseline", // https://trac.ffmpeg.org/wiki/Encode/H.264 - compatibility level
		// "-level", "4.0", // ^^^
		"-b:v", "1M",
		"-f", format,
		outfile.Name(),
	)
	var (
		outbuffer bytes.Buffer
		out       bytes.Buffer
		errout    bytes.Buffer
	)
	cmd.Stdout = &out

	cmd.Stderr = &errout

	err = cmd.Run()
	if err != nil {
		log.Printf("[ERROR] FFmpeg command : %v, %v, %v\n", err, out.String(), errout.String())
		return nil, err
	}

	_, err = io.Copy(&outbuffer, outfile)
	if err != nil {
		return nil, err
	}

	return &outbuffer, nil
}
