#!/bin/bash
set -e

CERT_DIR="../nginx/certs"
DOMAIN="go-project-macos.dev"
LOCAL_DNS="localhost"
LOCAL_IP="127.0.0.1"

# Detect OS
OS="$(uname -s)"

# Get primary IP dynamically per OS
get_ip() {
  case "$OS" in
    Darwin)  # macOS
      # Get IP for active Wi-Fi (en0) or primary interface
      IP=$(ipconfig getifaddr en0 2>/dev/null || echo "")
      if [ -z "$IP" ]; then
        # fallback: get first non-loopback IPv4 from ifconfig
        IP=$(ifconfig | awk '/inet / && $2 != "127.0.0.1" {print $2; exit}')
      fi
      echo "$IP"
      ;;
    Linux)
      # get IP of default route interface
      IFACE=$(ip route get 1 | awk '{for(i=1;i<=NF;i++) if($i=="dev") print $(i+1)}')
      IP=$(ip -4 addr show "$IFACE" | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1)
      echo "$IP"
      ;;
    CYGWIN*|MINGW32*|MSYS*|MINGW*)
      # Windows (via powershell)
      powershell.exe -Command "(Get-NetIPAddress -AddressFamily IPv4 -InterfaceAlias 'Wi-Fi').IPAddress" | tr -d '\r'
      ;;
    *)
      echo "Unsupported OS: $OS" >&2
      exit 1
      ;;
  esac
}

IP=$(get_ip)
if [ -z "$IP" ]; then
  echo "Failed to detect IP address!" >&2
  exit 1
fi

mkdir -p "$CERT_DIR"

echo "Using IP: $IP"

echo "Generating CA key and cert..."
openssl genrsa -out "$CERT_DIR/ca.key" 4096
openssl req -x509 -new -nodes -key "$CERT_DIR/ca.key" -sha256 -days 3650 \
  -subj "/C=US/ST=Local/L=Dev/O=GoProject CA/CN=GoProject Root CA" \
  -out "$CERT_DIR/ca.crt"

echo "Generating server key and CSR..."
openssl genrsa -out "$CERT_DIR/server.key" 2048

cat > "$CERT_DIR/server.cnf" <<EOF
[req]
default_bits       = 2048
prompt             = no
default_md         = sha256
req_extensions     = req_ext
distinguished_name = dn

[dn]
C  = US
ST = Local
L  = Dev
O  = GoProject
CN = $DOMAIN

[req_ext]
subjectAltName = @alt_names

[alt_names]
DNS.1   = $DOMAIN
DNS.2   = $LOCAL_DNS
IP.1    = $IP
IP.2    = $LOCAL_IP
EOF

openssl req -new -key "$CERT_DIR/server.key" -out "$CERT_DIR/server.csr" -config "$CERT_DIR/server.cnf"

echo "Signing server certificate with CA..."
cat > "$CERT_DIR/server_ext.cnf" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1   = $DOMAIN
DNS.2   = $LOCAL_DNS
IP.1    = $IP
IP.2    = $LOCAL_IP
EOF

openssl x509 -req -in "$CERT_DIR/server.csr" -CA "$CERT_DIR/ca.crt" -CAkey "$CERT_DIR/ca.key" -CAcreateserial \
  -out "$CERT_DIR/server.crt" -days 365 -sha256 -extfile "$CERT_DIR/server_ext.cnf"

echo "Cleaning up CSR and config files..."
rm "$CERT_DIR/server.csr" "$CERT_DIR/server.cnf" "$CERT_DIR/server_ext.cnf" "$CERT_DIR/ca.srl"

echo "Done! Certificates saved in $CERT_DIR:"
ls -l "$CERT_DIR"