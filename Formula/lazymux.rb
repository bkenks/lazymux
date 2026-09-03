class Lazymux < Formula
  desc "TUI repo manager that also serves its repo inventory over MCP"
  homepage "https://fj.ktbcloud.com/bkenks/lazymux"
  version "1.1.1"
  license "GPL-3.0-or-later"

  on_macos do
    on_arm do
      url "https://fj.ktbcloud.com/bkenks/lazymux/releases/download/1.1.1/lazymux-1.1.1-darwin-arm64"
      sha256 "f8d6935710dcc7e4432276e3b23da9c8c489de954f7c7a51414b34cb33d7d0bb"
    end
    on_intel do
      url "https://fj.ktbcloud.com/bkenks/lazymux/releases/download/1.1.1/lazymux-1.1.1-darwin-amd64"
      sha256 "ba516a4380b0aa0331fd365674697de68eaff116c85f52c08e824a1e1011809a"
    end
  end

  on_linux do
    on_arm do
      url "https://fj.ktbcloud.com/bkenks/lazymux/releases/download/1.1.1/lazymux-1.1.1-linux-arm64"
      sha256 "438f072ffb7fbce5dc5457546ea14fcd618e79d7a58869fa535041e3284e1e43"
    end
    on_intel do
      url "https://fj.ktbcloud.com/bkenks/lazymux/releases/download/1.1.1/lazymux-1.1.1-linux-amd64"
      sha256 "f69cf3759cda43c48807dc6aeb7fdc535155cec04f66e201450d82d3b44454e2"
    end
  end

  def install
    bin.install Dir["lazymux-*"].first => "lazymux"
  end

  service do
    run [opt_bin/"lazymux", "mcp", "serve"]
    keep_alive true
    working_dir Dir.home
    log_path var/"log/lazymux-mcp.log"
    error_log_path var/"log/lazymux-mcp.log"
  end

  test do
    assert_match "lazymux #{version}", shell_output("#{bin}/lazymux --version")
  end
end
