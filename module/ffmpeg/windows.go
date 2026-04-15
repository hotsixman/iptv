//go:build windows

package ffmpeg

import (
	"bufio"
	"fmt"
	"homecam/module/types"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func Device() ([]types.Device, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-list_devices", "true",
		"-f", "dshow",
		"-i", "dummy",
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stderr)
	re := regexp.MustCompile(`"(.+?)"`)

	var devices []types.Device

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "[dshow") || !strings.HasSuffix(line, "(video)") {
			continue
		}

		match := re.FindStringSubmatch(line)
		if len(match) > 1 {
			devices = append(devices, types.Device{
				Name: match[1],
			})
		}
	}

	return devices, nil
}

func Format(device types.Device) ([]types.Format, error) {
	cmd := exec.Command(
		"ffmpeg",
		"-f", "dshow",
		"-list_options", "true",
		"-i", "video="+device.Name,
	)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stderr)
	var formats []types.Format
	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "[dshow") {
			continue
		}

		var format types.Format
		valid := false
		for _, phrase := range strings.Split(line, " ") {
			if phrase == "" {
				continue
			}

			if strings.HasPrefix(phrase, "pixel_format=") {
				valid = true
				format.InputFormat = phrase[13:]
			} else if strings.HasPrefix(phrase, "s=") {
				numbers := strings.Split(phrase[2:], "x")

				width, err := strconv.Atoi(numbers[0])
				if err != nil {
					return nil, err
				}
				height, err := strconv.Atoi(numbers[1])
				if err != nil {
					return nil, err
				}

				format.Width = width
				format.Height = height
			} else if strings.HasPrefix(phrase, "fps=") {
				fps, err := strconv.ParseFloat(phrase[4:], 64)
				if err != nil {
					return nil, err
				}

				format.Fps = fps
			}
		}

		if valid {
			formats = append(formats, format)
		}
	}

	return formats, nil
}

/*
func MakeExecRaw(device types.Device, format types.Format) *exec.Cmd {
	return exec.Command("ffmpeg",
		"-f", "dshow",
		"-video_size", fmt.Sprintf("%dx%d", format.Width, format.Height),
		"-framerate", fmt.Sprintf("%f", format.Fps),
		//"-input_format", format.Codec,
		"-i", "video="+device.Name,
		"-pix_fmt", "bgr24",
		"-f", "rawvideo",
		"-loglevel", "quiet",
		"-",
	)
}
*/

func MakeExecH264(device types.Device, format types.Format, codec string) *exec.Cmd {
	return exec.Command("ffmpeg",
		"-f", "dshow",
		"-video_size", fmt.Sprintf("%dx%d", format.Width, format.Height),
		"-framerate", fmt.Sprintf("%g", format.Fps),
		"-i", "video="+device.Name,
		"-vcodec", codec,
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-g", "30", // 키프레임 간격 단축
		"-fflags", "nobuffer", // 버퍼링 방지
		"-flags", "low_delay", // 저지연 모드
		"-flush_packets", "1", // 패킷 즉시 플러시
		"-f", "mpegts",
		"-loglevel", "quiet",
		"-",
	)
}
