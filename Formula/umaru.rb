class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "2.0.1"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.1/umaru_2.0.1_darwin_arm64.tar.gz"
      sha256 "f31cbca65fdb9fe7f7ac23c5438ef062bd82400a486433560a78fa5d17192cee"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.1/umaru_2.0.1_darwin_amd64.tar.gz"
      sha256 "7c9a446e52b4ce1f6207662abb4a5dcfd83a80c2dce34613ef889772bf26e4ce"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.1/umaru_2.0.1_linux_arm64.tar.gz"
      sha256 "abf27fccb4653878fa65172485954531e5e8747d9f78eeec1127c0f742c6ff78"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.1/umaru_2.0.1_linux_amd64.tar.gz"
      sha256 "c629247b0569845170de7aafbf2aa2a3e8e98a2944c486ea3ad3ba23dce80c0c"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{bin}/umaru version")
  end
end
