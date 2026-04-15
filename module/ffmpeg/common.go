package ffmpeg

import (
	"bytes"
	"homecam/module/types"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"strconv"
)

func CaptureFrame(device types.Device, format types.Format, codec string, ch chan []byte, bufPacketCount int) {
	cmd := MakeExecH264(device, format, codec)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	defer cmd.Process.Kill()

	// MPEG-TS 패킷 크기에 맞춰 버퍼 정렬 (188 * N)
	packetSize := 188
	numPackets := bufPacketCount
	if numPackets < 1 {
		numPackets = 10 // 최소 10개 패킷 단위
	}
	alignSize := packetSize * numPackets

	for {
		// 1. 매 루프마다 첫 바이트가 0x47(Sync Byte)인지 확인
		syncBuf := make([]byte, 1)
		if _, err := io.ReadFull(stdout, syncBuf); err != nil {
			break
		}

		if syncBuf[0] != 0x47 {
			// 싱크가 깨졌다면 0x47이 나올 때까지 한 바이트씩 버리며 찾음
			continue
		}

		// 2. 싱크 바이트를 찾았으므로 나머지 데이터 읽기
		// (이미 1바이트를 읽었으므로 alignSize - 1 만큼 더 읽음)
		payload := make([]byte, alignSize-1)
		if _, err := io.ReadFull(stdout, payload); err != nil {
			break
		}

		// 3. 전체 패킷 조합 (0x47 + 나머지)
		fullData := append([]byte{0x47}, payload...)

		// 4. 채널 전송 (중요: 지연 방지 로직)
		select {
		case ch <- fullData:
			// 정상 전송
		default:
			// 채널이 꽉 찼다면 현재 데이터를 버리거나 채널의 오래된 데이터를 비움
			// 실시간 스트리밍에서는 '현재' 데이터가 가장 중요함
			select {
			case <-ch: // 가장 오래된 데이터 하나 제거
			default:
			}
			ch <- fullData
		}
	}
}

func LoadDeviceAndForamtFromEnv() (*types.Device, *types.Format, string, error) {
	device := &types.Device{Name: os.Getenv("DEVICE")}

	width, err := strconv.Atoi(os.Getenv("WIDTH"))
	if err != nil {
		return nil, nil, "", err
	}
	height, err := strconv.Atoi(os.Getenv("HEIGHT"))
	if err != nil {
		return nil, nil, "", err
	}
	fps, err := strconv.ParseFloat(os.Getenv("FPS"), 64)
	if err != nil {
		return nil, nil, "", err
	}
	format := &types.Format{
		InputFormat: os.Getenv("INPUT_FORMAT"),
		Width:       width,
		Height:      height,
		Fps:         fps,
	}

	codec := os.Getenv("CODEC")
	if codec == "" {
		codec = "libx264"
	}

	return device, format, codec, nil
}

func BufToPng(buf []byte, width int, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// RGB → RGBA 변환
	j := 0
	for i := 0; i < len(buf); i += 3 {
		img.Pix[j+0] = buf[i+2] // R
		img.Pix[j+1] = buf[i+1] // G
		img.Pix[j+2] = buf[i+0] // B
		img.Pix[j+3] = 255
		j += 4
	}

	return img
}

func EncodePng(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
