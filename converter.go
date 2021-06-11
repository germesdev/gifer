package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

func convert(src *bytes.Buffer, format string, dimensions string) (*bytes.Buffer, error) {
	hash := sha256.New()
	hash.Write(src.Bytes())
	hash.Write([]byte(format + dimensions))
	sum := hash.Sum(nil)
	cksum := fmt.Sprintf("%x", sum)

	convertQueue.ResultsLock.Lock()
	results, exists := convertQueue.ResultsQueue[cksum]

	if !exists {
		cv := ConvertQueue{
			Results: make(chan ConvertResult, 100),
			Waiting: 1,
		}
		convertQueue.ResultsQueue[cksum] = &cv
		results = &cv
		go func() {
			b, e := convertExec(src, format, dimensions)
			for {
				bufferCopy := bytes.NewBuffer(b.Bytes())
				cv.Results <- ConvertResult{Data: bufferCopy, Error: e}
				remaining := atomic.AddInt32(&cv.Waiting, -1)
				if remaining == 0 {
					delete(convertQueue.ResultsQueue, cksum)
					break
				}
			}
		}()
	} else {
		atomic.AddInt32(&results.Waiting, 1)
	}
	convertQueue.ResultsLock.Unlock()

	res := <-results.Results
	return res.Data, res.Error
}

func convertExec(src *bytes.Buffer, format string, dimensions string) (*bytes.Buffer, error) {
	inputfile, err := ioutil.TempFile("", "*")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Create Input error %v", err)
	}
	defer func() {
		err := os.Remove(inputfile.Name())
		if err != nil {
			log.Errorf("Cant delete inputfile %s error %s", inputfile.Name(), err)
		}
	}()

	_, err = io.Copy(inputfile, src)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Copy inputfile error %v", err)
	}

	if format == "gif" {
		return convertToGif(src, dimensions, inputfile)
	}

	outfile, err := ioutil.TempFile("", "res*."+format)
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Create Outfile error %v", err)
	}
	defer func() {
		err := os.Remove(outfile.Name())
		if err != nil {
			log.Errorf("Cant delete outfile %s error %s\n", outfile.Name(), err)
		}
	}()

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
			"-qmin", config.WEBM_QMIN, // the minimum quantizer (default 4, range 0–63), lower - better quality
			"-qmax", "63", // the maximum quantizer (default 63, range qmin–63) higher - lower quality
			"-crf", config.WEBM_CRF, // By default the CRF value can be from 4–63, and 10 is a good starting point. Lower values mean better quality.
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
			"-compression_level", "6", // Higher values give better quality for a given size. default - 4
			"-q:v", "75",
			"-t", "00:00:05",
			"-loop", "0",
			// "-qscale", "75", // For lossy encoding, this controls image quality, 0 to 100
		}...)
	case "gif":
		args = append(args, []string{
			// "-pix_fmt", "yuv420p",
			// "-c:v", "libwebp",
			// "-lossless", "0", // enable lossles. 1 - enable
			// "-compression_level", "4", // Higher values give better quality for a given size. default - 4
			// "-q:v", "25",
			// "-loop", "0",
			"-vframes", "1",
			"-f", "gif",
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
		log.Errorf("FFmpeg command : %v, %v, %v\n", err, out.String(), errout.String())
		return nil, err
	}

	_, err = io.Copy(&outbuffer, outfile)
	if err != nil {
		return nil, err
	}

	return &outbuffer, nil
}

func convertToGif(src *bytes.Buffer, dimensions string, inputfile *os.File) (*bytes.Buffer, error) {
	outfile, err := ioutil.TempFile("", "res*.gif")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Create Outfile error %v", err)
	}
	defer os.Remove(outfile.Name())

	palette, err := ioutil.TempFile("", "palette*.png")
	if err != nil {
		return nil, fmt.Errorf("[ERROR] Create Outfile error %v", err)
	}
	defer os.Remove(palette.Name())

	filters := "fps=15," + dimensions //+ ":flags=lanczos"

	paletteArgs := []string{
		"-i", inputfile.Name(),
		"-vf", filters + ",palettegen",
		"-vframes", "1",
		"-y", palette.Name(),
	}

	paletteCmd := exec.Command("ffmpeg", paletteArgs...)
	var errOut bytes.Buffer
	paletteCmd.Stderr = &errOut

	err = paletteCmd.Run()
	if err != nil {
		log.Errorf("FFmpeg command : %v, %v\n", err, errOut.String())
		return nil, err
	}

	convertArgs := []string{
		"-i", inputfile.Name(),
		"-i", palette.Name(),
		"-vframes", "1",
		"-lavfi", filters + " [x]; [x][1:v] paletteuse",
		"-y", outfile.Name(),
	}

	convertCmd := exec.Command("ffmpeg", convertArgs...)
	errOut = bytes.Buffer{}
	convertCmd.Stderr = &errOut

	err = convertCmd.Run()
	if err != nil {
		log.Errorf("FFmpeg command : %v, %v\n", err, errOut.String())
		return nil, err
	}

	var outbuffer bytes.Buffer
	_, err = io.Copy(&outbuffer, outfile)
	if err != nil {
		return nil, err
	}

	return &outbuffer, nil
}
