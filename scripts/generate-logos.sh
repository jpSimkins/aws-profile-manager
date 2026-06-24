#!/bin/bash
# Generate PNG logo variants from Icon.svg (single source of truth)
# Note: Using PNG because Fyne cannot render complex SVG paths (1800+ bezier curves)

set -e

ICON_PNG="Icon.png"
LOGO_DIR="internal/bundled/assets"

# Ensure the assets directory exists
mkdir -p "$LOGO_DIR"

# Logo dimensions for GUI header (height constrained to 50px)
# Icon aspect ratio is ~1.42:1, so width = 50 * 1.42 ≈ 71px
LOGO_WIDTH=71
LOGO_HEIGHT=50

# Generate at 2x resolution for crisp display on high-DPI screens
PNG_WIDTH=$((LOGO_WIDTH * 2))
PNG_HEIGHT=$((LOGO_HEIGHT * 2))

# Check dependencies
if command -v magick >/dev/null 2>&1; then
  IM=magick
elif command -v convert >/dev/null 2>&1; then
  IM=convert
else
  echo "ERROR: ImageMagick not found - cannot generate PNG logos"
  echo "Install with: sudo apt-get install imagemagick"
  exit 1
fi

if [ ! -f "$ICON_PNG" ]; then
  echo "ERROR: $ICON_PNG not found - run 'make icon' first"
  exit 1
fi

# Generate light mode logo from Icon.png (preserves original colors)
$IM "$ICON_PNG" \
  -resize ${PNG_WIDTH}x${PNG_HEIGHT} \
  -background none \
  -gravity center \
  -extent ${PNG_WIDTH}x${PNG_HEIGHT} \
  -transparent white \
  "$LOGO_DIR/logo.png"

echo "✓ Generated $LOGO_DIR/logo.png (${PNG_WIDTH}x${PNG_HEIGHT} @ 2x, display: ${LOGO_WIDTH}x${LOGO_HEIGHT})"

# Generate dark mode logo with brightened colors for visibility on dark backgrounds
# Brighten blues and oranges for better contrast
$IM "$ICON_PNG" \
  -resize ${PNG_WIDTH}x${PNG_HEIGHT} \
  -background none \
  -gravity center \
  -extent ${PNG_WIDTH}x${PNG_HEIGHT} \
  -transparent white \
  -modulate 120,110,100 \
  "$LOGO_DIR/logo-dark-mode.png"

echo "✓ Generated $LOGO_DIR/logo-dark-mode.png (${PNG_WIDTH}x${PNG_HEIGHT} @ 2x, display: ${LOGO_WIDTH}x${LOGO_HEIGHT})"
echo "  (Brightened by 20% for dark mode visibility)"
