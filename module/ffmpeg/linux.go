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
	args := []string{
		"-f", "v4l2",
	}

	// 입력 포맷이 지정되어 있다면 추가 (mjpeg, yuyv422 등)
	if format.InputFormat != "" {
		args = append(args, "-input_format", format.InputFormat)
	}

	args = append(args,
		"-video_size", fmt.Sprintf("%dx%d", format.Width, format.Height),
		"-framerate", fmt.Sprintf("%g", format.Fps),
		"-i", device.Name,
		"-vcodec", codec,
	)

	// 코덱별 최적 프리셋 설정
	if strings.Contains(codec, "nvenc") {
		args = append(args,
			"-preset", "p1", // 최저 지연
			"-rc", "vbr",
			"-cq", "28",
			"-delay", "0",
		)
	} else if strings.Contains(codec, "vaapi") {
		args = append(args, "-compression_level", "1")
	} else if strings.Contains(codec, "qsv") {
		args = append(args, "-preset", "veryfast")
	} else if strings.Contains(codec, "v4l2m2m") || strings.Contains(codec, "omx") {
		// ARM 하드웨어 가속 (v4l2m2m, omx)는 일반적인 -preset 옵션을 지원하지 않음
		// 추가 설정이 필요하다면 여기에 작성 (보통 기본값으로도 충분히 빠름)
		args = append(args, "-num_capture_buffers", "16")
	} else {
		// 소프트웨어 인코더 (libx264 등)
		args = append(args, "-preset", "ultrafast", "-tune", "zerolatency")
	}

	args = append(args,
		"-g", "30",
		"-fflags", "nobuffer",
		"-flags", "low_delay",
		"-flush_packets", "1",
		"-f", "mpegts",
		"-loglevel", "quiet",
		"-",
	)

	return exec.Command("ffmpeg", args...)
}
