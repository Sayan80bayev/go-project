#!/usr/bin/env bash
#This script has been made to recognize domain of another laptop which is running as a dns server

set -e

# === CHANGE THIS MANUALLY WHEN MACOS IP CHANGES ===
IP="192.168.0.42"   # Your IP on the network / hotspot
IFACE="wlan0"           # Your active Linux network interface

echo "📡 Setting DNS to use Mac at $IP for .dev domains"

# Point systemd-resolved to macOS for DNS
sudo resolvectl dns "$IFACE" "$IP"

# Disable DNSSEC (in case Mac's dnsmasq doesn't support it)
sudo resolvectl dnssec "$IFACE" no

# Route all .dev domain lookups to this DNS
sudo resolvectl domain "$IFACE" "~dev"

echo "✅ DNS updated for .dev via $IP"