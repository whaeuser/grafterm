#!/bin/bash

# Install Go for macOS
set -e

echo "=== Go Installation für macOS ==="

# Prüfe ob Go bereits installiert ist
if command -v go &> /dev/null; then
    echo "Go ist bereits installiert:"
    go version
    echo ""
    echo "Möchtest du Go neu installieren? (y/N)"
    read -r answer
    if [[ "$answer" != "y" && "$answer" != "Y" ]]; then
        echo "Installation abgebrochen."
        exit 0
    fi
fi

# Bestimme die Architektur
ARCH=$(uname -m)
if [[ "$ARCH" == "x86_64" ]]; then
    GO_ARCH="amd64"
elif [[ "$ARCH" == "arm64" ]]; then
    GO_ARCH="arm64"
else
    echo "Nicht unterstützte Architektur: $ARCH"
    exit 1
fi

echo "Erkannte Architektur: $GO_ARCH"

# Download URL (aktuellste stabile Version)
GO_VERSION="1.21.5"  # Aktuelle stabile Version
DOWNLOAD_URL="https://go.dev/dl/go${GO_VERSION}.darwin-${GO_ARCH}.tar.gz"

echo "Herunterladen von Go $GO_VERSION für macOS..."
echo "URL: $DOWNLOAD_URL"

# Download mit curl oder wget
if command -v curl &> /dev/null; then
    curl -L -O "$DOWNLOAD_URL"
elif command -v wget &> /dev/null; then
    wget "$DOWNLOAD_URL"
else
    echo "Weder curl noch wget gefunden. Bitte installiere eines dieser Tools."
    exit 1
fi

# Entpacken nach /usr/local
echo "Entpacken nach /usr/local..."
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go${GO_VERSION}.darwin-${GO_ARCH}.tar.gz"

# Aufräumen
rm -f "go${GO_VERSION}.darwin-${GO_ARCH}.tar.gz"

# PATH setzen
echo "Setze PATH in ~/.zshrc oder ~/.bash_profile..."

# Shell erkennen
SHELL_NAME=$(basename "$SHELL")
if [[ "$SHELL_NAME" == "zsh" ]]; then
    PROFILE_FILE="$HOME/.zshrc"
elif [[ "$SHELL_NAME" == "bash" ]]; then
    PROFILE_FILE="$HOME/.bash_profile"
else
    PROFILE_FILE="$HOME/.profile"
fi

# PATH hinzufügen, wenn nicht vorhanden
if ! grep -q "/usr/local/go/bin" "$PROFILE_FILE" 2>/dev/null; then
    echo "" >> "$PROFILE_FILE"
    echo "# Go PATH" >> "$PROFILE_FILE"
    echo 'export PATH=$PATH:/usr/local/go/bin' >> "$PROFILE_FILE"
    echo "PATH wurde zu $PROFILE_FILE hinzugefügt"
fi

# Aktuelles Terminal aktualisieren
export PATH=$PATH:/usr/local/go/bin

echo ""
echo "=== Installation abgeschlossen! ==="
echo "Starte ein neues Terminal oder führe aus:"
echo "source $PROFILE_FILE"
echo ""
echo "Prüfe Installation:"
go version
echo ""
echo "Jetzt kannst du grafterm bauen mit:"
echo "./build.sh"