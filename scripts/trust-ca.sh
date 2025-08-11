#!/bin/bash
set -e

CERT_PATH="../nginx/certs/ca.crt"

if [ ! -f "$CERT_PATH" ]; then
  echo "CA certificate not found at $CERT_PATH"
  exit 1
fi

OS="$(uname -s)"

case "$OS" in
  Darwin)
    echo "Installing CA certificate on macOS..."

    # Requires sudo
    sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain "$CERT_PATH"

    echo "CA installed. You may need to restart your browser or system for changes to take effect."
    ;;

  Linux)
    echo "Installing CA certificate on Linux..."

    # Detect distro to choose CA directory
    if [ -d /usr/local/share/ca-certificates ]; then
      # Debian/Ubuntu
      sudo cp "$CERT_PATH" /usr/local/share/ca-certificates/go-project-ca.crt
      sudo update-ca-certificates
      echo "CA installed for Debian/Ubuntu."
    elif [ -d /etc/pki/ca-trust/source/anchors ]; then
      # RHEL/CentOS/Fedora
      sudo cp "$CERT_PATH" /etc/pki/ca-trust/source/anchors/go-project-ca.crt
      sudo update-ca-trust extract
      echo "CA installed for RHEL/CentOS/Fedora."
    else
      echo "Unsupported Linux distro or CA path not found. Please install the CA cert manually."
      exit 1
    fi
    ;;

  *)
    echo "Unsupported OS: $OS. Please trust the CA manually."
    exit 1
    ;;
esac