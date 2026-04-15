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
    - **Security**: Implements a two-step authentication system. 
    1.  `POST /auth`: Client sends an `x-api-key` header matching the `API_KEY` environment variable. The server returns a one-time 16-byte random ticket.
    2.  `GET /ws?ticket=...`: Client connects via WebSocket using the issued ticket. The ticket is immediately invalidated upon use.
- **Types (`module/types/`)**: Shared data structures like `Device` and `Format` used across `ffmpeg` and `main.go`.

### Frontend (HTML/JS)
- **`index.html`**:
    - **Authentication UI**: A password-style input for the `API_KEY` that triggers the `/auth` request.
    - **Streaming Player**: Uses **mpegts.js** for playing the MPEG-TS stream in the browser via Media Source Extensions (MSE).
    - **Latency Management**: Periodically (every 3 seconds) checks the video buffer. If the delay (`buffered.end - currentTime`) exceeds 1 second, it jumps the `currentTime` to `end - 0.2` to maintain real-time playback.
    - **Statistics**: Calculates and displays real-time FPS and bitrate using `mpegts.Events.STATISTICS_INFO` and the browser's `VideoPlaybackQuality` API.

## 3. Key Components and Logic
- **`CaptureFrame` in `module/ffmpeg/common.go`**:
    - This is the heart of the data pipeline. It manages a buffer of 188-byte packets (MPEG-TS standard).
    - It uses a channel (`chan []byte`) with a buffer size (default 100) to decouple capture from distribution.
    - **Sync Byte Detection**: To handle potential stream corruption, it looks for the MPEG-TS sync byte (`0x47`). If it doesn't find it, it scans the stream until it finds the next one.
    - **Latency Handling**: If the distribution channel is full, it drops the oldest packet before pushing the new one. This ensures that only the latest frames are queued for broadcasting.
- **`BroadcastStream` in `module/server/server.go`**:
    - Continuously pulls data from the channel and iterates over all connected WebSocket clients to push the binary data.
    - Automatically removes clients if a write operation fails.

## 4. Configuration
The application is configured via environment variables (usually in a `.env` file):
- `DEVICE`: The name of the capture device (e.g., `video=Integrated Camera` on Windows, `/dev/video0` on Linux).
- `WIDTH`, `HEIGHT`, `FPS`: Video parameters.
- `CODEC`: The FFmpeg encoder codec (e.g., `libx264`, `h264_nvenc`).
- `API_KEY`: Secret for authentication.
- `PORT`: Server port (default 3000).
- `ORIGIN`: Optional. If set, restricts WebSocket connections to the specified origin.

## 5. Build and Deployment
- **`build.yml` (.github/workflows)**: Automates building for multiple platforms (Windows, Linux, Darwin) on push/PR to the `main` branch.
- **Environment Setup**: Requires FFmpeg to be installed and available in the system's `PATH`.
- **Operating Systems**: 
    - **Windows**: Uses `dshow` for device capture.
    - **Linux**: Uses `v4l2` for device capture.
