# IPTV Project Codebase Analysis

## 1. Project Overview
This project is a low-latency video streaming server implemented in Go. It captures video from a local device using FFmpeg and broadcasts it to multiple web clients via WebSockets using the MPEG-TS format.

## 2. Architecture

### Backend (Go)
- **Main Entry (`main.go`)**: Orchestrates the application. It loads configuration from `.env`, initializes the FFmpeg capture process, and starts the web server.
- **FFmpeg Module (`module/ffmpeg/`)**:
    - **Platform-Specific Commands**: `linux.go` and `windows.go` define how FFmpeg is invoked on each OS. On Windows, it uses DirectShow (`dshow`).
    - **MPEG-TS Parsing (`common.go`)**: Reads the stdout of the FFmpeg subprocess. It uses a sync-byte detection mechanism (looking for `0x47`) to ensure it sends complete MPEG-TS packets to the distribution channel.
    - **Latency Optimization**: Uses FFmpeg flags like `-tune zerolatency`, `-preset ultrafast`, and `-fflags nobuffer`.
- **Server Module (`module/server/`)**:
    - **Gin Web Framework**: Serves a static `index.html` and provides a `/ws` WebSocket endpoint.
    - **WebSocket Broadcasting**: Uses `BroadcastStream` to pull data from the capture channel and push it to all connected WebSocket clients.
    - **Security**: Implements a simple API key check (`x-api-key` header) during the WebSocket handshake.
- **Types (`module/types/`)**: Shared data structures like `Device` and `Format`.

### Frontend (HTML/JS)
- **`index.html`**:
    - Uses **mpegts.js** for playing the MPEG-TS stream in the browser via Media Source Extensions (MSE).
    - **Latency Management**: Periodically checks the video buffer. If the delay exceeds 1 second, it jumps the `currentTime` forward to minimize lag.
    - **Statistics**: Calculates and displays real-time FPS and bitrate using `mpegts.Events.STATISTICS_INFO` and the browser's `VideoPlaybackQuality` API.

## 3. Key Components and Logic
- **`CaptureFrame` in `module/ffmpeg/common.go`**:
    - This is the heart of the data pipeline. It manages a buffer of 188-byte packets (MPEG-TS standard).
    - It uses a channel (`chan []byte`) with a buffer size (default 100) to decouple capture from distribution.
    - If the channel is full, it drops the oldest packet to maintain low latency.
- **WebSocket Handshake**:
    - The server checks for an `x-api-key` header against the `API_KEY` environment variable.

## 4. Configuration
The application is configured via environment variables (usually in a `.env` file):
- `DEVICE`: The name of the capture device.
- `WIDTH`, `HEIGHT`, `FPS`: Video parameters.
- `API_KEY`: Secret for WebSocket authentication.
- `PORT`: Server port (default 3000).

## 5. Potential Improvements
- **Dynamic Device Selection**: Currently, the device name is hardcoded in `.env`. A discovery API could be added.
- **Improved Security**: Moving beyond a simple API key in the header, perhaps using JWT or more robust session management.
- **Transcoding Options**: Adding support for different bitrates or resolutions on the fly.
