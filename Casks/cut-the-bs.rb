cask "cut-the-bs" do
  version :latest
  sha256 :no_check

  url "https://github.com/bradhannah/cut-the-bs/releases/latest/download/cut-the-bs-macos-universal.tar.gz"
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
