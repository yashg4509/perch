# frozen_string_literal: true

class Perch < Formula
  desc "Terminal UI for exploring multi-service deployment stacks"
  homepage "https://github.com/yashg4509/perch"
  license "MIT"
  head "https://github.com/yashg4509/perch.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/perch"
  end

  test do
    assert_match "perch", shell_output("#{bin}/perch --help")
  end
end
