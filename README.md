# IPTV Low-Latency Streaming Server

[English](#english) | [한국어](#한국어)

---

## English

A high-performance, low-latency video streaming server implemented in Go. It captures video from local devices (camera, screen, etc.) using FFmpeg and broadcasts it to multiple web clients via WebSockets using the MPEG-TS format.

### 🚀 Features

- **Low Latency**: Optimized FFmpeg flags and active buffer management in the browser ensure sub-second delay.
- **Cross-Platform**: Supports Windows (DirectShow) and Linux (V4L2).
- **Security**: Two-step ticket-based authentication system.
- **Real-time Stats**: Live FPS and bitrate monitoring in the web interface.
- **Robustness**: Automatic MPEG-TS sync-byte detection and backpressure handling.

### 🛠 Prerequisites

- **FFmpeg**: Must be installed and available in your system's `PATH`.

### 🚀 Getting Started (Binary)

1. **Download**: Obtain the compiled binary for your OS.
2. **Environment Setup**: Create a `.env` file in the same directory as the binary.
   ```bash
   DEVICE=Integrated Camera
   WIDTH=1280
   HEIGHT=720
   FPS=30
   CODEC=mjpeg
   API_KEY=your-secure-key
   PORT=3000
   ```
3. **Run**:
   ```bash
   ./iptv  # Linux/macOS
   iptv.exe # Windows
   ```

### ⚙️ Configuration

| Variable | Description | Example (Windows) | Example (Linux) |
| :--- | :--- | :--- | :--- |
| `DEVICE` | Capture device name | `Integrated Camera` | `/dev/video0` |
| `WIDTH` | Video width | `1280` | `1280` |
| `HEIGHT` | Video height | `720` | `720` |
| `FPS` | Frames per second | `30` | `30` |
| `CODEC` | FFmpeg encoder | `mjpeg` | `mjpeg` |
| `API_KEY` | Secret for authentication | `your-key` | `your-key` |
| `PORT` | Server port | `3000` | `3000` |

---

## 한국어

Go로 구현된 고성능 저지연 비디오 스트리밍 서버입니다. FFmpeg를 사용하여 로컬 장치(카메라, 화면 등)에서 영상을 캡처하고, MPEG-TS 형식을 통해 WebSocket으로 여러 웹 클라이언트에 방송합니다.

### 🚀 주요 기능

- **저지연(Low Latency)**: 최적화된 FFmpeg 플래그와 브라우저의 능동적 버퍼 관리를 통해 1초 미만의 지연 시간을 보장합니다.
- **교차 플랫폼**: Windows (DirectShow) 및 Linux (V4L2)를 지원합니다.
- **보안**: 2단계 티켓 기반 인증 시스템을 적용했습니다.
- **실시간 통계**: 웹 인터페이스에서 실시간 FPS 및 비트레이트 모니터링이 가능합니다.
- **견고성**: 자동 MPEG-TS 동기화 바이트 감지 및 백프레셔 처리 기능을 포함합니다.

### 🛠 필수 요구 사항

- **FFmpeg**: 시스템에 설치되어 있어야 하며 `PATH`에 등록되어 있어야 합니다.

### 🚀 시작하기 (바이너리 실행)

1. **다운로드**: 운영체제에 맞는 빌드된 바이너리 파일을 다운로드합니다.
2. **환경 설정**: 바이너리와 같은 폴더에 `.env` 파일을 생성하고 설정을 입력합니다.
   ```env
   DEVICE=Integrated Camera
   WIDTH=1280
   HEIGHT=720
   FPS=30
   CODEC=mjpeg
   API_KEY=your-secure-key
   PORT=3000
   ```
3. **실행**:
   ```bash
   ./iptv  # Linux/macOS
   iptv.exe # Windows
   ```

### ⚙️ 상세 설정

| 변수명 | 설명 | Windows 예시 | Linux 예시 |
| :--- | :--- | :--- | :--- |
| `DEVICE` | 캡처 장치 이름 | `Integrated Camera` | `/dev/video0` |
| `WIDTH` | 비디오 너비 | `1280` | `1280` |
| `HEIGHT` | 비디오 높이 | `720` | `720` |
| `FPS` | 초당 프레임 수 | `30` | `30` |
| `CODEC` | FFmpeg 인코더 | `mjpeg` | `mjpeg` |
| `API_KEY` | 인증용 비밀키 | `your-key` | `your-key` |
| `PORT` | 서버 포트 | `3000` | `3000` |

### 🔒 보안 아키텍처

1. **인증 요청**: 클라이언트가 `x-api-key` 헤더를 포함하여 `/auth`에 `POST` 요청을 보냅니다.
2. **티켓 발행**: 서버는 키를 검증하고 16바이트 무작위 일회용 티켓을 반환합니다.
3. **WS 연결**: 클라이언트는 `/ws?ticket=...`로 연결하며, 서버는 티켓 확인 즉시 해당 티켓을 폐기합니다.
