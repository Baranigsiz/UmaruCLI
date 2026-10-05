class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "2.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.1.0/umaru_2.1.0_darwin_arm64.tar.gz"
      sha256 "086af2d3382c543df56ba15d21b6aed671cd3f6407b763d5f21267a276a2f50a"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.1.0/umaru_2.1.0_darwin_amd64.tar.gz"
      sha256 "0345aeecd0085f1b7db255dc2a2f26204c8eba3cc76c651f167a6e66d57ea53a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.1.0/umaru_2.1.0_linux_arm64.tar.gz"
      sha256 "a12849a7a332dad0b8aee2c07f76501c0211388de9e6e75e65fdec4164f6fb77"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.1.0/umaru_2.1.0_linux_amd64.tar.gz"
      sha256 "a3570eefacb1b57a6f0073aed7dedb0a1873bcf046c491cec159981c77cf636e"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{bin}/umaru version")
  end
end
