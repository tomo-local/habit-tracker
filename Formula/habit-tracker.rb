class HabitTracker < Formula
  desc "CLI tool for tracking daily habits using Google Calendar"
  homepage "https://github.com/tomo-local/habit-tracker"
  version "0.0.1-beta.2"
  license "MIT"

  on_macos do
    on_arm do
      url "https://github.com/tomo-local/habit-tracker/releases/download/v0.0.1-beta.2/habit-tracker-darwin-arm64"
      sha256 "307e3226be608e2e7895f357a751b72096dc644b8f6c77470063bbeb5b744ee9"
    end
    on_intel do
      url "https://github.com/tomo-local/habit-tracker/releases/download/v0.0.1-beta.2/habit-tracker-darwin-amd64"
      sha256 "0764a2836ab38e3ebe736e4037d2016c20d076f6d55745c130787d5fc3dd4b71"
    end
  end

  on_linux do
    url "https://github.com/tomo-local/habit-tracker/releases/download/v0.0.1-beta.2/habit-tracker-linux-amd64"
    sha256 "bf84d4b1d3319e06843b951798a60e6e24854e3d34642f5563e2b15130589120"
  end

  def install
    binary_name = Dir["habit-tracker-*"].first
    bin.install binary_name => "habit"
  end

  test do
    system "#{bin}/habit", "-h"
  end
end
