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

	outfile, err := ioutil.TempFile("", "res*."+format)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Create Outfile error %v", err)
	}
	defer os.Remove(outfile.Name())

	args := []string{
		"-an", // disable audio
		"-y",  // overwrite
		// "-trans_color", "ffffff", // TODO read from input
		"-i", inputfile.Name(), // set input
		"-vf", dimensions,
		"-movflags", "faststart",
	}

	switch format {
	case "webm":
		args = append(args, []string{
			"-pix_fmt", "yuva420p",
			"-auto-alt-ref", "0",
			"-qmin", "8", // the minimum quantizer (default 4, range 0–63), lower - better quality --- VP9 only
			"-qmax", "63", // the maximum quantizer (default 63, range qmin–63) higher - lower quality --- VP9 only
			"-crf", "25", // By default the CRF value can be from 4–63, and 10 is a good starting point. Lower values mean better quality.
			"-maxrate", "500k",
			"-minrate", "250K",
			"-c:v", "libvpx",
			"-b:v", "500k",
			"-f", format,
		}...)
	case "mp4":
		args = append(args, []string{
			"-pix_fmt", "yuv420p",
			"-preset", "medium", // quality preset
			"-maxrate", "500k",
			"-minrate", "250K",
			"-profile:v", "baseline", // https://trac.ffmpeg.org/wiki/Encode/H.264 - compatibility level
			"-level", "3.1", // ^^^
			"-crf", "25", // enable constant bitrate(0-51) lower - better
			"-c:v", "libx264",
			"-refs", "2",
			"-f", format,
		}...)
	case "jpg":
		args = append(args, []string{
			"-vframes", "1",
			"-f", "image2",
		}...)
	case "webp":
		args = append(args, []string{
			"-pix_fmt", "yuv420p",
			"-c:v", "libwebp",
			"-lossless", "0", // enable lossles. 1 - enable
			"-compression_level", "4", // Higher values give better quality for a given size. default - 4
			"-q:v", "25",
			"-loop", "0",
			// "-qscale", "75", // For lossy encoding, this controls image quality, 0 to 100
		}...)
	}

	args = append(args, outfile.Name())

	cmd := exec.Command("ffmpeg", args...)
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
