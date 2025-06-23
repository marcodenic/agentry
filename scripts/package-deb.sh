#!/usr/bin/env bash
set -euo pipefail

VERSION="$1"
ARCH="amd64"
BUILD_DIR="dist/deb"
INSTALL_DIR="$BUILD_DIR/usr/local/bin"

mkdir -p "$INSTALL_DIR" "$BUILD_DIR/DEBIAN"
cp "dist/agentry-linux-$ARCH" "$INSTALL_DIR/agentry"
chmod 755 "$INSTALL_DIR/agentry"

cat > "$BUILD_DIR/DEBIAN/control" <<EOF2
Package: agentry
Version: $VERSION
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: Agentry Team <dev@none>
Description: Minimal, performant AI-Agent runtime
EOF2

dpkg-deb --build "$BUILD_DIR" "dist/agentry_${VERSION}_${ARCH}.deb"
printf "Debian package created at dist/agentry_${VERSION}_${ARCH}.deb\n"
