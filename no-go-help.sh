#!/bin/bash

# Alternative: Online Go Compiler für einfache Tests
set -e

echo "=== Go Compiler Online Services ==="
echo ""
echo "Da Go nicht lokal installiert ist, kannst du:"
echo ""
echo "1. Go lokal installieren (empfohlen):"
echo "   ./install-go.sh"
echo ""
echo "2. Online Go Compiler verwenden:"
echo "   - https://go.dev/play/ (für kleine Tests)"
echo "   - https://replit.com/@replit/Go-Template"
echo "   - https://www.onlinegdb.com/online_go_compiler"
echo ""
echo "3. Go mit Homebrew installieren:"
echo "   brew install go"
echo ""
echo "4. Manuelle Installation:"
echo "   - Download von https://golang.org/dl/"
echo "   - Entpacken nach /usr/local"
echo "   - PATH setzen: export PATH=\$PATH:/usr/local/go/bin"
echo ""
echo "5. Docker verwenden (wenn installiert):"
echo "   docker run -v \$(pwd):/src -w /src golang:latest go build ./cmd/grafterm"
echo ""
echo "Für die Entwicklung wird lokale Go-Installation empfohlen."
echo "Möchtest du Go jetzt installieren? (y/N)"
read -r answer

if [[ "$answer" == "y" || "$answer" == "Y" ]]; then
    echo "Starte Installation..."
    ./install-go.sh
else
    echo "Ok. Hier sind die Optionen zum Testen ohne lokale Go-Installation:"
    
    # Prüfe ob Docker verfügbar ist
    if command -v docker &> /dev/null; then
        echo ""
        echo "✅ Docker gefunden - kann verwendet werden:"
        echo "Baue mit Docker:"
        echo "docker run -v \$(pwd):/src -w /src golang:latest go build -o bin/grafterm ./cmd/grafterm"
        echo ""
        echo "Tests mit Docker:"
        echo "docker run -v \$(pwd):/src -w /src golang:latest go test ./..."
    fi
    
    # Prüfe ob Homebrew verfügbar ist
    if command -v brew &> /dev/null; then
        echo ""
        echo "✅ Homebrew gefunden - installiere Go mit:"
        echo "brew install go"
    fi
    
    echo ""
    echo "Oder führe die Installation aus:"
    echo "./install-go.sh"
fi