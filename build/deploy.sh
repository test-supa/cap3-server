#!/bin/bash
# Deploy Chameleon C2 Server
set -e

echo "[+] Creating directories..."
mkdir -p /opt/chameleon-c2
mkdir -p /opt/chameleon-c2/data

echo "[+] Copying files..."
cp chameleon-c2 /opt/chameleon-c2/
cp config.yaml /opt/chameleon-c2/
chmod +x /opt/chameleon-c2/chameleon-c2

echo "[+] Creating data directory..."
touch /opt/chameleon-c2/data/.gitkeep

echo "[+] Installing systemd service..."
cp systemd/chameleon-c2.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable chameleon-c2

echo "[+] Installing nginx config..."
cp nginx/chameleon-c2.conf /etc/nginx/sites-available/
ln -sf /etc/nginx/sites-available/chameleon-c2.conf /etc/nginx/sites-enabled/
nginx -t && systemctl reload nginx

echo "[+] Starting service..."
systemctl restart chameleon-c2

echo "[+] Checking status..."
sleep 2
systemctl status chameleon-c2 --no-pager

echo ""
echo "[+] Deploy complete!"
echo "    API:  https://208414.landvps.online/api/health"
echo "    WS:   wss://208414.landvps.online/ws"
echo "    Logs: journalctl -u chameleon-c2 -f"
