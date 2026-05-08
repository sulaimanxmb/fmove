class Fmove < Formula
  desc "Native Go + FFmpeg Concatenation & Space Recovery tool"
  homepage "https://github.com/sulaimaneksambi/fmove"
  url "https://github.com/sulaimaneksambi/fmove.git",
      tag:      "v1.0.0",
      revision: "UPDATE_WITH_COMMIT_HASH"
  license "MIT"

  depends_on "go" => :build
  depends_on "pkg-config" => :build
  depends_on "ffmpeg"

  def install
    # Set CGO flags so the Go compiler can find Homebrew's FFmpeg libraries
    ENV["PKG_CONFIG_PATH"] = "#{HOMEBREW_PREFIX}/lib/pkgconfig"
    ENV["CGO_CFLAGS"] = "-I#{HOMEBREW_PREFIX}/include"
    ENV["CGO_LDFLAGS"] = "-L#{HOMEBREW_PREFIX}/lib -lavformat -lavcodec -lavutil"

    system "go", "build", *std_go_args(ldflags: "-s -w"), "."
  end

  test do
    system "#{bin}/fmove", "--animation"
  end
end
