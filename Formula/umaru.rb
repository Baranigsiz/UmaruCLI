class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "2.0.3"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.3/umaru_2.0.3_darwin_arm64.tar.gz"
      sha256 "f5893524333dabe74b3d8308c536811905f8613b28ce38bb77a0dc6087498830"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.3/umaru_2.0.3_darwin_amd64.tar.gz"
      sha256 "1caf116115226d7db759919f046e90f05b7c80e4b5f5b9ecc22bc39cad0c6841"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.3/umaru_2.0.3_linux_arm64.tar.gz"
      sha256 "bbba017253b1bd758cc9721dff1bbf217059040fd2255c0b7a7625c5da9ffad2"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.3/umaru_2.0.3_linux_amd64.tar.gz"
      sha256 "a43f37f3fdb69812541a731cad9eeae1e1db6c590fef18962ac3a8c1e6d389d8"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{bin}/umaru version")
  end
end
