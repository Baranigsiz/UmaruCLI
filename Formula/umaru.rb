class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "2.0.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_darwin_arm64.tar.gz"
      sha256 "7f2ac4c303cccc7cf281e2961261058b4424ddab92f2a836c7c6b8f2d5b7cf8f"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_darwin_amd64.tar.gz"
      sha256 "2b722b64f36c91c9959d57547e3d5d498221d64c9da75e58bb9b48319d145cec"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_linux_arm64.tar.gz"
      sha256 "f6c3fa2ff1a853db2a489a2044756d7cd5f5394df21af40f5c94fe5d347bed8b"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.0/umaru_2.0.0_linux_amd64.tar.gz"
      sha256 "fd4d060024dd9787ba3c559b323e4502a38d388fd1e5fa34e8b1fc6829e21248"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{bin}/umaru version")
  end
end
