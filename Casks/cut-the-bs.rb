cask "cut-the-bs" do
  version "0.1.4"
  sha256 "1652be3c44e86b1ca4868bb32c5fd1fbf02e7cd780ad62441e53803a9d5d31cf"

  url "https://github.com/bradhannah/cut-the-bs/releases/download/v#{version}/cut-the-bs-macos-universal.tar.gz"
  name "Cut the BS"
  desc "Desktop app for tailoring resumes and cover letters"
  homepage "https://github.com/bradhannah/cut-the-bs"

  app "cut-the-bs.app"

  postflight do
    system_command "/usr/bin/xattr",
                   args:         ["-dr", "com.apple.quarantine", "#{appdir}/cut-the-bs.app"],
                   must_succeed: false
  end

  zap trash: [
    "~/Library/Application Support/cut-the-bs",
    "~/Library/Preferences/com.wails.cut-the-bs.plist",
    "~/Library/Saved Application State/com.wails.cut-the-bs.savedState",
  ]

  caveats do
    <<~EOS
      This app is currently unsigned. If macOS still blocks launch, run:
        xattr -dr com.apple.quarantine "/Applications/cut-the-bs.app"
    EOS
  end
end
