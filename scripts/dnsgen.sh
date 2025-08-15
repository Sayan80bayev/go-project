#!/usr/bin/env bash

# Detect OS
OS="$(uname -s)"

# Function to get primary IPv4
get_ip() {
  if [[ "$OS" == "Darwin" ]]; then
    # macOS - get IP of primary interface (usually en0 for Wi-Fi)
    ip=$(ifconfig en0 | awk '/inet / {print $2}' | head -n 1)
  else
    # Linux - get IP of primary interface (excluding 127.0.0.1)
    ip=$(hostname -I | awk '{print $1}')
  fi
  echo "$ip"
}

# Get current IP
IP=$(get_ip)

if [[ -z "$IP" ]]; then
  echo "❌ Could not detect IP address. Check your network connection."
  exit 1
fi

echo "📡 Detected IP: $IP"

# Pick correct dnsmasq config path
if [[ "$OS" == "Darwin" ]]; then
  CONFIG_PATH="/opt/homebrew/etc/dnsmasq.conf"
else
  CONFIG_PATH="/etc/dnsmasq.conf"
fi

# Generate dnsmasq.conf
sudo tee "$CONFIG_PATH" > /dev/null <<EOF
# Map go-project-macos.dev to your current IP
address=/go-project-macos.dev/$IP

# Listen for DNS queries on localhost and your IP
listen-address=127.0.0.1
listen-address=$IP

# Don't use /etc/resolv.conf for upstream lookups
no-resolv

# Use public DNS for all other requests
server=8.8.8.8
server=1.1.1.1
EOF

echo "✅ Generated $CONFIG_PATH with IP $IP"

# Restart dnsmasq
if [[ "$OS" == "Darwin" ]]; then
  sudo brew services restart dnsmasq
else
  sudo systemctl restart dnsmasq
fi

echo "🚀 dnsmasq restarted. Now point your DNS to:"
echo "  - On this machine: 127.0.0.1"
echo "  - On other devices: $IP"