class Agentry < Formula
  desc "Minimal, performant AI-Agent runtime"
  homepage "https://github.com/marcodenic/agentry"
  version "0.1.1"

  on_macos do
    url "https://github.com/marcodenic/agentry/releases/download/v#{version}/agentry-darwin-amd64"
    sha256 "PUT_SHA256_HERE"
  end

  on_linux do
    url "https://github.com/marcodenic/agentry/releases/download/v#{version}/agentry-linux-amd64"
    sha256 "PUT_SHA256_HERE"
  end

  def install
    bin.install "agentry"
  end
end
