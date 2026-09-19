class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "2.0.2"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.2/umaru_2.0.2_darwin_arm64.tar.gz"
      sha256 "d490e569f75276cdf3eb2dbc83e98f09eafa44246b8945da0427fdf37276572a"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.2/umaru_2.0.2_darwin_amd64.tar.gz"
      sha256 "57ccafd255f1469d1b555f0def801fc6d3cc1e1dbd7af22892238704fc72e0cc"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.2/umaru_2.0.2_linux_arm64.tar.gz"
      sha256 "8a3ccaa3c84e8a77c1d2150d8db92a001f7bfb1da783ab2730975f6bef32b9e6"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/v2.0.2/umaru_2.0.2_linux_amd64.tar.gz"
      sha256 "eaaa7f60d7a89e2bd48a48104054a0fd13fd40abba8652ec78d12f334d083026"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{bin}/umaru version")
  end
end
