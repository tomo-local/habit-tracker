class HabitTracker < Formula
  desc "CLI tool for tracking daily habits using Google Calendar"
  homepage "https://github.com/tomo-local/habit-tracker"
  version "0.0.1-beta.1"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/tomo-local/habit-tracker/releases/download/v#{version}/habit-tracker-darwin-arm64"
      sha256 "264948e41c001317ef26bc690d5938268a55a35ae08b1f5e57f59288ae61a6f4"
    end
    on_intel do
      url "https://github.com/tomo-local/habit-tracker/releases/download/v#{version}/habit-tracker-darwin-amd64"
      sha256 "b3dc0faa1d58dcdb3035f9d086c72a763710eeabcc38e7aa0242df664f5ee031"
    end
  end

  on_linux do
    url "https://github.com/tomo-local/habit-tracker/releases/download/v#{version}/habit-tracker-linux-amd64"
    sha256 "b682a4c6db91c5c3b7c389b66fd3c7a29fe95e684a4cd8df540bc20074a3edc4"
  end

  def install
    binary_name = Dir["habit-tracker-*"].first
    bin.install binary_name => "habit"
  end

  test do
    system "#{bin}/habit", "-h"
  end
end
