class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "2.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_darwin_arm64.tar.gz"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_darwin_amd64.tar.gz"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_linux_arm64.tar.gz"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_linux_amd64.tar.gz"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{bin}/umaru version")
  end
end
