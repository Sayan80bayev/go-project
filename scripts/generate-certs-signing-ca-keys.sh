#!/bin/bash
set -euo pipefail

CERT_DIR="../nginx/certs"
DOMAIN="${DOMAIN:-go-project-keycloak}"
LOCAL_DNS="localhost"
LOCAL_IP="127.0.0.1"

CA_CRT="$CERT_DIR/ca.crt"
CA_KEY="$CERT_DIR/ca.key"

if [ ! -f "$CA_CRT" ] || [ ! -f "$CA_KEY" ]; then
  echo "Missing CA files. Expected:"
  echo "  $CA_CRT"
  echo "  $CA_KEY"
  exit 1
fi

mkdir -p "$CERT_DIR"

# --- OS & IP detection ---
OS="$(uname -s)"

get_ip() {
  case "$OS" in
    Darwin)
      IP="$(ipconfig getifaddr en0 2>/dev/null || true)"
      if [ -z "${IP:-}" ]; then
        DEF_IFACE="$(route -n get default 2>/dev/null | awk '/interface:/{print $2}')"
        [ -n "${DEF_IFACE:-}" ] && IP="$(ipconfig getifaddr "$DEF_IFACE" 2>/dev/null || true)"
      fi
      [ -z "${IP:-}" ] && IP="$(ifconfig | awk '/inet / && $2!="127.0.0.1"{print $2; exit}')"
      ;;
    Linux)
      DEF_IFACE="$(ip route get 1.1.1.1 2>/dev/null | awk '/dev/{for(i=1;i<=NF;i++) if($i=="dev") print $(i+1)}' | head -1)"
      IP="$(ip -4 addr show "$DEF_IFACE" 2>/dev/null | awk '/inet /{print $2}' | cut -d/ -f1 | head -1)"
      ;;
    *)
      echo "Unsupported OS: $OS" >&2
      exit 1
      ;;
  esac
  echo "${IP:-}"
}

IP="$(get_ip)"
if [ -z "$IP" ]; then
  echo "Failed to detect IP address!" >&2
  exit 1
fi

echo "Using IP: $IP"
echo "Domain: $DOMAIN"

# --- Server key & CSR ---
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

# --- Server cert extensions ---
cat > "$CERT_DIR/server_ext.cnf" <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1   = $DOMAIN
DNS.2   = $LOCAL_DNS
IP.1    = $IP
IP.2    = $LOCAL_IP
EOF

# --- Sign with existing CA ---
openssl x509 -req -in "$CERT_DIR/server.csr" \
  -CA "$CA_CRT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$CERT_DIR/server.crt" -days 365 -sha256 -extfile "$CERT_DIR/server_ext.cnf"

# --- Cleanup ---
rm -f "$CERT_DIR/server.csr" "$CERT_DIR/server.cnf" "$CERT_DIR/server_ext.cnf" "$CERT_DIR/ca.srl" || true

echo "Done. Files in $CERT_DIR:"
ls -l "$CERT_DIR"