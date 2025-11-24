# Homebrew Formula for tf-safe
# This file should be submitted to homebrew-core or maintained in a custom tap

class TfSafe < Formula
  desc "Terraform state file protection and backup tool"
  homepage "https://github.com/BIRhrt/tf-safe"
  version "1.1.0"
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-darwin-arm64.tar.gz"
      sha256 "8011fd9cf6e68c881b832c0f1465bcd0c42a634d4e28536928f991fe80847d0b"
    else
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-darwin-amd64.tar.gz"
      sha256 "c231e268aadd68ba630f777d57a40c03f139cc7bcd096c4bdeeaf553f6b39b38"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-linux-arm64.tar.gz"
      sha256 "aee60cff593449bf2104c2c9042454589c963ea46210fcf7e306cf48d1d3b6b8"
    else
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-linux-amd64.tar.gz"
      sha256 "ea25dead3188784207a4d08b5de352b30c8d70a743eb7f999a445478518b3533"
    end
  end

  depends_on "terraform" => :recommended

  def install
    bin.install "tf-safe-#{OS.kernel_name.downcase}-#{Hardware::CPU.arch}" => "tf-safe"
    
    # Generate shell completions
    generate_completions_from_executable(bin/"tf-safe", "completion")
  end

  test do
    system "#{bin}/tf-safe", "--version"
    
    # Test basic functionality
    system "#{bin}/tf-safe", "--help"
    
    # Test init command (should create config file)
    system "#{bin}/tf-safe", "init", "--dry-run"
  end
end