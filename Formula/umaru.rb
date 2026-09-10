class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "1.9.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_darwin_arm64.tar.gz"
      sha256 "cd651ef60f91880ede6a84e83be5fdeb1f6675676932023443dd47f038d6aae2"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_darwin_amd64.tar.gz"
      sha256 "d0bac69bde6a197a366a8de36d9928306e40e054232fe627f9cf853d24b2f67e"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_linux_arm64.tar.gz"
      sha256 "f00c1412ac56eaeb29c684b476475c851ae32a806e9371a5454a6bc8d6f7265f"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v1.9.0/umaru_1.9.0_linux_amd64.tar.gz"
      sha256 "a2a4875dc6985759becaf13272c6551a81c958f4e722f0ff5efdf425a86bf96f"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{bin}/umaru version")
  end
end
