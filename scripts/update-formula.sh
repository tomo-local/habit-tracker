#!/usr/bin/env bash
set -euo pipefail

# Regenerates Formula/habit-tracker.rb for the given tag using the
# release assets already uploaded to GitHub Releases.
#
# Usage: scripts/update-formula.sh v1.2.3

TAG="${1:?usage: update-formula.sh <tag>}"
VERSION="${TAG#v}"
REPO="tomo-local/habit-tracker"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

gh release download "$TAG" --repo "$REPO" --dir "$WORKDIR" \
  --pattern "habit-tracker-darwin-amd64" \
  --pattern "habit-tracker-darwin-arm64" \
  --pattern "habit-tracker-linux-amd64" \
  --clobber

sha_darwin_amd64="$(shasum -a 256 "$WORKDIR/habit-tracker-darwin-amd64" | cut -d' ' -f1)"
sha_darwin_arm64="$(shasum -a 256 "$WORKDIR/habit-tracker-darwin-arm64" | cut -d' ' -f1)"
sha_linux_amd64="$(shasum -a 256 "$WORKDIR/habit-tracker-linux-amd64" | cut -d' ' -f1)"

cat > Formula/habit-tracker.rb <<EOF
class HabitTracker < Formula
  desc "CLI tool for tracking daily habits using Google Calendar"
  homepage "https://github.com/${REPO}"
  version "${VERSION}"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/${REPO}/releases/download/${TAG}/habit-tracker-darwin-arm64"
      sha256 "${sha_darwin_arm64}"
    end
    on_intel do
      url "https://github.com/${REPO}/releases/download/${TAG}/habit-tracker-darwin-amd64"
      sha256 "${sha_darwin_amd64}"
    end
  end

  on_linux do
    url "https://github.com/${REPO}/releases/download/${TAG}/habit-tracker-linux-amd64"
    sha256 "${sha_linux_amd64}"
  end

  def install
    binary_name = Dir["habit-tracker-*"].first
    bin.install binary_name => "habit"
  end

  test do
    system "#{bin}/habit", "-h"
  end
end
EOF

echo "Formula/habit-tracker.rb updated for ${TAG}"
