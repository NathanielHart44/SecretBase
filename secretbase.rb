class Secretbase < Formula
    desc "Manage environment variables for projects across different environments."
    homepage "https://github.com/NathanielHart44/homebrew-secret-base"
    url "https://github.com/NathanielHart44/homebrew-secret-base/releases/download/alpha/SecretBase.v0.0.1-alpha"
    sha256 "3eec12f2890fca90d28b8b2264dd5e69e2b772c225f8972e399d6cf2ad270c5e"
    license "MIT"
    version "0.0.1-alpha"
  
    def install
      bin.install "SecretBase"
    end
  
    test do
      assert_match "SecretBase version 0.0.1-alpha", shell_output("#{bin}/SecretBase --version")
    end
  end