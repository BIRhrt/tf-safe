# Homebrew Formula for tf-safe
# This file should be submitted to homebrew-core or maintained in a custom tap

class TfSafe < Formula
  desc "Terraform state file protection and backup tool"
  homepage "https://github.com/BIRhrt/tf-safe"
  version "1.0.0"
  
  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-darwin-arm64.tar.gz"
      sha256 "b5df27f1b9e87d5bf49525cfa293b207b5caed082e5aa0772b1975224116ef16"
    else
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-darwin-amd64.tar.gz"
      sha256 "77258995a3b835f45c3f04d4f18780bfce340c2a0475a65182f92c809ded589a"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-linux-arm64.tar.gz"
      sha256 "d8923310099a127b5e181659f0c43f4d45e364b740f65b76394fba18c86d4bf1"
    else
      url "https://github.com/BIRhrt/tf-safe/releases/download/v#{version}/tf-safe-linux-amd64.tar.gz"
      sha256 "eed72faf397159b65126d796d68de827eb1b93207edffc5ee353383f00cc6cab"
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