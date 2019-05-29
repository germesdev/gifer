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
		// "-movflags", "faststart",
		// "-qmin", "10", // the minimum quantizer (default 4, range 0–63), lower - better quality --- VP9 only
		// "-qmax", "42", // the maximum quantizer (default 63, range qmin–63) higher - lower quality --- VP9 only
		// "-crf", "23", // enable constant bitrate(0-51) lower - better
		// "-preset", "medium", // quality preset
		// "-maxrate", "500k", // max bitrate. higher - better
		// "-profile:v", "baseline", // https://trac.ffmpeg.org/wiki/Encode/H.264 - compatibility level
		// "-level", "4.0", // ^^^
		"-f", format,
		outfile.Name(),
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	var errout bytes.Buffer
	cmd.Stderr = &errout

	err = cmd.Run()
	if err != nil {
		log.Printf("[ERROR] FFmpeg command : %v, %v, %v\n", err, out.String(), errout.String())
		return nil, err
	}

	output, err := ioutil.ReadAll(outfile)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(output), nil
}
