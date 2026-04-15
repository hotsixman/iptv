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
	args := []string{
		"-f", "dshow",
	}

	// 입력 포맷이 지정되어 있다면 추가 (예: mjpeg, yuyv422 등)
	if format.InputFormat != "" {
		args = append(args, "-pixel_format", format.InputFormat)
	}

	args = append(args,
		"-video_size", fmt.Sprintf("%dx%d", format.Width, format.Height),
		"-framerate", fmt.Sprintf("%g", format.Fps),
		"-i", "video="+device.Name,
		"-vcodec", codec,
	)

	// 코덱별 최적 프리셋 설정
	if strings.Contains(codec, "nvenc") {
		args = append(args,
			"-preset", "p1", // 최저 지연/최고 속도
			"-rc", "vbr",
			"-cq", "28", // 화질과 압축률 균형
			"-delay", "0",
		)
	} else if strings.Contains(codec, "qsv") {
		args = append(args, "-preset", "veryfast")
	} else {
		args = append(args, "-preset", "ultrafast", "-tune", "zerolatency")
	}

	args = append(args,
		"-g", "30", // 키프레임 간격
		"-fflags", "nobuffer",
		"-flags", "low_delay",
		"-flush_packets", "1",
		"-f", "mpegts",
		"-loglevel", "quiet",
		"-",
	)

	return exec.Command("ffmpeg", args...)
}
