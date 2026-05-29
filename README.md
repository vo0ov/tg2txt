<h1 align="center">✨ tg2txt</h1>

<p align="center">
  Tiny Go CLI that converts <strong>Telegram Desktop JSON exports</strong> into clean, searchable TXT transcripts
</p>

<p align="center">
  <a href="https://github.com/vo0ov/tg2txt/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/vo0ov/tg2txt/ci.yml?branch=main&label=CI&logo=githubactions&logoColor=white" alt="CI"/></a>
  <a href="https://github.com/vo0ov/tg2txt/actions/workflows/cd.yml"><img src="https://img.shields.io/github/actions/workflow/status/vo0ov/tg2txt/cd.yml?branch=main&label=CD&logo=githubactions&logoColor=white" alt="CD"/></a>
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26"/>
  <img src="https://img.shields.io/badge/Telegram-export-26A5E4?logo=telegram&logoColor=white" alt="Telegram export"/>
  <img src="https://img.shields.io/badge/Platforms-Windows%20%7C%20macOS%20%7C%20Linux-2EA44F" alt="Platforms"/>
  <img src="https://img.shields.io/badge/License-MIT-green" alt="MIT License"/>
</p>

---

## 📖 About

**tg2txt** turns Telegram Desktop's `result.json` export into a compact plain-text chat log. It keeps the transcript easy to read, grep, archive, diff, index, or pass into other tools.

The output includes messages, timestamps, sender names, replies, forwards, media markers, reactions, and Telegram service events.

---

## ✨ Features

| Area               | Description                                                                  |
| ------------------ | ---------------------------------------------------------------------------- |
| **Text Export**    | Converts Telegram JSON exports into readable `.txt` transcripts              |
| **Formatting**     | Preserves basic chunks: links, bold, italic, code, pre, and spoilers         |
| **Media Markers**  | Adds labels for photos, videos, voice messages, stickers, files, and GIFs    |
| **Context**        | Keeps replies, forwarded-from labels, contacts, polls, and locations         |
| **Reactions**      | Writes reaction summaries with emoji, names, and counts                      |
| **Service Events** | Handles calls, joins, leaves, pins, renamed chats, and group updates         |
| **CLI Flags**      | Supports custom input/output, `--no-service`, `--no-header`, and `--version` |
| **Release Builds** | Ships archives for Windows, macOS, and Linux across common architectures     |

---

## 📦 Installation

Download the archive for your OS from [GitHub Releases](https://github.com/vo0ov/tg2txt/releases).

| OS      | Architecture  | Archive                                |
| ------- | ------------- | -------------------------------------- |
| Linux   | amd64         | `tg2txt_<version>_linux_amd64.tar.gz`  |
| Linux   | arm64         | `tg2txt_<version>_linux_arm64.tar.gz`  |
| Linux   | armv7         | `tg2txt_<version>_linux_armv7.tar.gz`  |
| macOS   | Intel         | `tg2txt_<version>_darwin_amd64.tar.gz` |
| macOS   | Apple Silicon | `tg2txt_<version>_darwin_arm64.tar.gz` |
| Windows | amd64         | `tg2txt_<version>_windows_amd64.zip`   |
| Windows | arm64         | `tg2txt_<version>_windows_arm64.zip`   |
| Windows | 386           | `tg2txt_<version>_windows_386.zip`     |

---

## 🌍 Add To PATH

After these steps, `tg2txt` will work from any terminal folder.

### 🪟 Windows PowerShell

```powershell
Expand-Archive .\tg2txt_vX.Y.Z_windows_amd64.zip -DestinationPath .\tg2txt
New-Item -ItemType Directory -Force "$HOME\bin"
Move-Item .\tg2txt\tg2txt.exe "$HOME\bin\"

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$HOME\bin*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$HOME\bin", "User")
}
```

Open a new terminal:

```powershell
tg2txt --version
```

### 🍎 macOS

Apple Silicon:

```bash
tar -xzf tg2txt_vX.Y.Z_darwin_arm64.tar.gz
mkdir -p ~/.local/bin
mv tg2txt_vX.Y.Z_darwin_arm64/tg2txt ~/.local/bin/
chmod +x ~/.local/bin/tg2txt
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
tg2txt --version
```

Intel:

```bash
tar -xzf tg2txt_vX.Y.Z_darwin_amd64.tar.gz
mkdir -p ~/.local/bin
mv tg2txt_vX.Y.Z_darwin_amd64/tg2txt ~/.local/bin/
chmod +x ~/.local/bin/tg2txt
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
tg2txt --version
```

### 🐧 Linux

```bash
tar -xzf tg2txt_vX.Y.Z_linux_amd64.tar.gz
mkdir -p ~/.local/bin
mv tg2txt_vX.Y.Z_linux_amd64/tg2txt ~/.local/bin/
chmod +x ~/.local/bin/tg2txt
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
tg2txt --version
```

For ARM64, use `tg2txt_vX.Y.Z_linux_arm64.tar.gz`. For ARMv7, use `tg2txt_vX.Y.Z_linux_armv7.tar.gz`.

---

## ⚙️ Usage

```text
tg2txt [flags]

-i FILE        input Telegram JSON export (default: result.json)
-o FILE        output TXT file           (default: chat.txt)
--no-service   skip service events
--no-header    skip the "# Chat Name" header
--version      show build information
```

```bash
tg2txt
tg2txt -i result.json -o chat.txt
tg2txt -i backup.json --no-service
tg2txt --no-header
```

---

## 🧾 Output

```text
# Family Chat

[15.01.26 13:50] #421 Alice: Hey! Are we meeting at 7?
[15.01.26 13:51] #422 Bob: [🎤 voice message (0:23)]  [Reactions: 👍 by Alice]
[15.01.26 13:55] #423 :: Alice: 📌 pinned a message
```

---

## 🧱 Architecture

```text
cmd/
└── main.go                  Entry point

internal/
├── cli/                     Flag parsing, help text, stdout/stderr handling
├── converter/               JSON file loading and TXT file writing
├── formatter/               Message, media, reaction, and service formatting
├── telegram/                Telegram export JSON schema
└── version/                 Build-time version, commit, date via ldflags
```

---

## 🧪 Development

### 🛠 Commands

```bash
make run          # run from source
make test         # go test -v ./...
make race         # tests with race detector
make lint         # golangci-lint
make sec          # gosec
make compile      # compile all packages
make build        # build dist/tg2txt for this OS
make release      # build release archives
make ci           # race + lint + sec + compile
```

### 🏗 Build From Source

```bash
go build -o tg2txt ./cmd
./tg2txt --version
```

---

## 🚀 Release Matrix

| Target  | Architectures             |
| ------- | ------------------------- |
| Linux   | `amd64`, `arm64`, `armv7` |
| macOS   | `amd64`, `arm64`          |
| Windows | `amd64`, `arm64`, `386`   |

Releases are produced by [CD](https://github.com/vo0ov/tg2txt/actions/workflows/cd.yml) when a `v*` tag is pushed.

---

## 📜 License

This project is released under the **MIT License**.

See [LICENSE](./LICENSE).
