# Homebrew Formula for tf-safe
# This file should be submitted to homebrew-core or maintained in a custom tap

class TfSafe < Formula
  desc "Terraform state file protection and backup tool"
  homepage "https://github.com/BIRhrt/tf-safe"
  version "1.0.0"
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-darwin-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_ARM64"
    else
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-darwin-amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_AMD64"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-linux-arm64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_ARM64"
    else
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-linux-amd64.tar.gz"
      sha256 "REPLACE_WITH_ACTUAL_SHA256_FOR_LINUX_AMD64"
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