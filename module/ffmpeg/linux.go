//go:build linux

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
		"-f", "v4l2",
		"-list_devices", "true",
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
	// Example: [video4linux2,v4l2 @ 0x5580a1329c40] /dev/video0: USB Camera (046d:0825)
	re := regexp.MustCompile(`/dev/video\d+`)

	var devices []types.Device

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.Contains(line, "/dev/video") {
			continue
		}

		match := re.FindString(line)
		if match != "" {
			devices = append(devices, types.Device{
				Name: match,
			})
		}
	}

	return devices, nil
}

func Format(device types.Device) ([]types.Format, error) {
	// Linux (v4l2) does not provide a structured list_options like dshow.
	// We use list_formats to get resolutions.
	cmd := exec.Command(
		"ffmpeg",
		"-f", "v4l2",
		"-list_formats", "all",
		"-i", device.Name,
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

	// Example line: [video4linux2,v4l2 @ 0x...] Raw       :     yuyv422 :           YUYV 4:2:2 : 640x480 160x120 ...
	resRe := regexp.MustCompile(`(\d+)x(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.Contains(line, "Raw") && !strings.Contains(line, "Compressed") {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}

		inputFormat := strings.TrimSpace(parts[2])
		resPart := parts[3]

		matches := resRe.FindAllStringSubmatch(resPart, -1)
		for _, match := range matches {
			width, _ := strconv.Atoi(match[1])
			height, _ := strconv.Atoi(match[2])

			formats = append(formats, types.Format{
				InputFormat: inputFormat,
				Width:       width,
				Height:      height,
				Fps:         30, // v4l2 list_formats doesn't easily show FPS, defaulting to 30
			})
		}
	}

	return formats, nil
}

func MakeExecH264(device types.Device, format types.Format, codec string) *exec.Cmd {
	return exec.Command("ffmpeg",
		"-f", "v4l2", // Linux의 비디오 장치 프레임워크
		"-input_format", format.InputFormat,
		"-video_size", fmt.Sprintf("%dx%d", format.Width, format.Height),
		"-framerate", fmt.Sprintf("%g", format.Fps),
		"-i", device.Name, // Linux는 보통 "/dev/video0" 형태의 경로를 사용합니다
		"-vcodec", codec,
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-g", "30",
		"-fflags", "nobuffer",
		"-flags", "low_delay",
		"-flush_packets", "1",
		"-f", "mpegts",
		"-loglevel", "quiet",
		"-",
	)
}
