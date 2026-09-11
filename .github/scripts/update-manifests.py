#!/usr/bin/env python3
import os
import re
import sys
import json

def main():
    tag = os.environ.get("GITHUB_REF_NAME", "")
    if not tag:
        if len(sys.argv) > 1:
            tag = sys.argv[1]
        else:
            print("Error: GITHUB_REF_NAME environment variable not set")
            sys.exit(1)

    version = tag.lstrip("v")
    checksum_file = "dist/checksums.txt"

    if not os.path.exists(checksum_file):
        # Fallback if running outside goreleaser default dist folder
        if os.path.exists("checksums.txt"):
            checksum_file = "checksums.txt"
        else:
            print(f"Warning: Checksum file not found at {checksum_file}, skipping hash updates")
            checksum_file = None

    hashes = {}
    if checksum_file:
        with open(checksum_file, "r", encoding="utf-8") as f:
            for line in f:
                parts = line.strip().split()
                if len(parts) >= 2:
                    hashes[parts[1]] = parts[0]
        print(f"Loaded {len(hashes)} checksums from {checksum_file}")

    # 1. Update bucket/umaru.json (Scoop)
    win_amd64 = f"umaru_{version}_windows_amd64.zip"
    win_arm64 = f"umaru_{version}_windows_arm64.zip"
    if os.path.exists("bucket/umaru.json"):
        with open("bucket/umaru.json", "r", encoding="utf-8") as f:
            manifest = json.load(f)
        
        manifest["version"] = version
        manifest["architecture"]["64bit"]["url"] = f"https://github.com/Baranigsiz/UmaruCLI/releases/download/{tag}/{win_amd64}"
        if win_amd64 in hashes:
            manifest["architecture"]["64bit"]["hash"] = hashes[win_amd64]

        manifest["architecture"]["arm64"]["url"] = f"https://github.com/Baranigsiz/UmaruCLI/releases/download/{tag}/{win_arm64}"
        if win_arm64 in hashes:
            manifest["architecture"]["arm64"]["hash"] = hashes[win_arm64]

        with open("bucket/umaru.json", "w", encoding="utf-8") as f:
            json.dump(manifest, f, indent=2)
            f.write("\n")
        print("Successfully updated bucket/umaru.json")

    # 2. Update Formula/umaru.rb (Homebrew)
    darwin_arm64 = f"umaru_{version}_darwin_arm64.tar.gz"
    darwin_amd64 = f"umaru_{version}_darwin_amd64.tar.gz"
    linux_arm64 = f"umaru_{version}_linux_arm64.tar.gz"
    linux_amd64 = f"umaru_{version}_linux_amd64.tar.gz"

    if os.path.exists("Formula/umaru.rb"):
        with open("Formula/umaru.rb", "r", encoding="utf-8") as f:
            old_formula = f.read()

        def extract_sha(arch_suffix, text):
            m = re.search(rf'umaru_[^"]+_{re.escape(arch_suffix)}".*?\n\s+sha256 "([a-f0-9]+)"', text, re.DOTALL)
            return m.group(1) if m else ""

        sha_darwin_arm64 = hashes.get(darwin_arm64) or extract_sha("darwin_arm64.tar.gz", old_formula)
        sha_darwin_amd64 = hashes.get(darwin_amd64) or extract_sha("darwin_amd64.tar.gz", old_formula)
        sha_linux_arm64 = hashes.get(linux_arm64) or extract_sha("linux_arm64.tar.gz", old_formula)
        sha_linux_amd64 = hashes.get(linux_amd64) or extract_sha("linux_amd64.tar.gz", old_formula)

        if not all([sha_darwin_arm64, sha_darwin_amd64, sha_linux_arm64, sha_linux_amd64]):
            print("Warning: One or more Homebrew SHA-256 hashes could not be resolved from checksums or existing formula")

        formula_content = f"""class Umaru < Formula
  desc "Blazing-fast CLI to scaffold modern fullstack, backend, frontend & CLI starters"
  homepage "https://github.com/Baranigsiz/UmaruCLI"
  version "{version}"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/{tag}/{darwin_arm64}"
      sha256 "{sha_darwin_arm64}"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/{tag}/{darwin_amd64}"
      sha256 "{sha_darwin_amd64}"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/{tag}/{linux_arm64}"
      sha256 "{sha_linux_arm64}"
    else
      url "https://github.com/Baranigsiz/UmaruCLI/releases/download/{tag}/{linux_amd64}"
      sha256 "{sha_linux_amd64}"
    end
  end

  def install
    bin.install "umaru"
  end

  test do
    assert_match "Umaru CLI", shell_output("#{{bin}}/umaru version")
  end
end
"""
        with open("Formula/umaru.rb", "w", encoding="utf-8") as f:
            f.write(formula_content)
        print("Successfully updated Formula/umaru.rb")

    # 3. Update install.sh and install.ps1 fallback versions
    if os.path.exists("install.sh"):
        with open("install.sh", "r", encoding="utf-8") as f:
            content = f.read()
        content = re.sub(r'LATEST_TAG="v[0-9\.]+"', f'LATEST_TAG="{tag}"', content)
        with open("install.sh", "w", encoding="utf-8") as f:
            f.write(content)
        print("Successfully updated install.sh fallback tag")

    if os.path.exists("install.ps1"):
        with open("install.ps1", "r", encoding="utf-8") as f:
            content = f.read()
        content = re.sub(r'falling back to v[0-9\.]+"', f'falling back to {tag}"', content)
        content = re.sub(r'return "v[0-9\.]+"', f'return "{tag}"', content)
        with open("install.ps1", "w", encoding="utf-8") as f:
            f.write(content)
        print("Successfully updated install.ps1 fallback tag")

    # 4. Update README.md
    if os.path.exists("README.md"):
        with open("README.md", "r", encoding="utf-8") as f:
            readme = f.read()
        readme = re.sub(r'Latest: \[v[0-9\.]+\]\(https://github.com/Baranigsiz/UmaruCLI/releases/tag/v[0-9\.]+\)',
                        f'Latest: [{tag}](https://github.com/Baranigsiz/UmaruCLI/releases/tag/{tag})', readme)
        readme = re.sub(r'umaru_[0-9\.]+_windows_amd64\.zip', f'umaru_{version}_windows_amd64.zip', readme)
        readme = re.sub(r'umaru_[0-9\.]+_windows_arm64\.zip', f'umaru_{version}_windows_arm64.zip', readme)
        readme = re.sub(r'umaru_[0-9\.]+_darwin_arm64\.tar\.gz', f'umaru_{version}_darwin_arm64.tar.gz', readme)
        readme = re.sub(r'umaru_[0-9\.]+_darwin_amd64\.tar\.gz', f'umaru_{version}_darwin_amd64.tar.gz', readme)
        readme = re.sub(r'umaru_[0-9\.]+_linux_amd64\.tar\.gz', f'umaru_{version}_linux_amd64.tar.gz', readme)
        readme = re.sub(r'umaru_[0-9\.]+_linux_arm64\.tar\.gz', f'umaru_{version}_linux_arm64.tar.gz', readme)
        readme = re.sub(r'/releases/download/v[0-9\.]+/', f'/releases/download/{tag}/', readme)
        with open("README.md", "w", encoding="utf-8") as f:
            f.write(readme)
        print("Successfully updated README.md links")

if __name__ == "__main__":
    main()
