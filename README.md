# ByteVeil

<p align="center">
  <strong>Hide files, folders, and text inside ordinary files.</strong>
</p>

<p align="center">
  A lightweight Go CLI for file-based steganography with optional compression and AES-256-GCM encryption.
</p>

<p align="center">
  <a href="https://github.com/batmanpriv/ByteVeil/releases/tag/v1.0.0">
    <img src="https://img.shields.io/badge/version-1.0.0-blue.svg" alt="Version">
  </a>
  <a href="https://github.com/batmanpriv/ByteVeil">
    <img src="https://img.shields.io/badge/built%20with-Go-00ADD8.svg" alt="Built with Go">
  </a>
  <a href="https://github.com/batmanpriv/ByteVeil">
    <img src="https://img.shields.io/github/stars/batmanpriv/ByteVeil?style=flat" alt="GitHub Stars">
  </a>
</p>

---

## Overview

**ByteVeil** is a command-line steganography tool written in Go that lets you hide a file, an entire folder, or a text message inside a cover file.

Unlike tools that focus exclusively on hiding data inside image pixels, ByteVeil works at the **file/container level**. It detects the cover format and uses a format-specific embedding strategy when available, with a generic fallback for unsupported formats.

Payloads are compressed with **gzip** before storage. When a password is supplied, the compressed payload is encrypted using **AES-256-GCM**, with a random salt and nonce derived/generated for each encoded payload.

ByteVeil is designed for simple, portable, file-based data hiding:

```text
Input payload
     │
     ▼
┌─────────────┐
│   gzip      │
│ compression │
└──────┬──────┘
       │
       ▼
┌─────────────────────┐
│ Optional encryption │
│    AES-256-GCM      │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│ ByteVeil container  │
│ metadata + payload  │
│ + integrity check   │
└──────────┬──────────┘
           │
           ▼
     Cover file
           │
           ▼
      Stego file
```

## Features

* Hide **files**
* Hide **folders/directories**
* Hide **text messages**
* Optional **AES-256-GCM encryption**
* **PBKDF2-HMAC-SHA256** password-based key derivation
* **Random cryptographic salt and nonce**
* **gzip compression** before encryption/storage
* CRC32 integrity checking for the ByteVeil payload container
* Automatic detection of the embedded payload
* Original filename preservation
* Automatic directory detection and extraction
* Format-aware embedding for several common file formats
* Generic fallback for unsupported formats
* Single static Go binary
* No external runtime dependencies

---

## Supported Cover Formats

ByteVeil currently recognizes the following formats:

| Format                  |        Detection | Embedding strategy                       |
| ----------------------- | ---------------: | ---------------------------------------- |
| JPEG                    |              Yes | Appended after the JPEG data             |
| PNG                     |              Yes | Custom ancillary PNG chunk before `IEND` |
| MP4 / ISO-BMFF          |              Yes | Appended `free` box                      |
| MP3                     |              Yes | `PRIV` frame inside an ID3v2 tag         |
| GIF                     |              Yes | Appended after the file                  |
| BMP                     |              Yes | Appended after the bitmap data           |
| RIFF / WAV / AVI / WEBP |              Yes | Appended payload                         |
| OGG                     |              Yes | Appended payload                         |
| PDF                     |              Yes | Appended after `%%EOF`                   |
| ZIP                     |              Yes | ZIP EOCD comment field                   |
| DOCX                    | Yes, through ZIP | ZIP comment field                        |
| XLSX                    | Yes, through ZIP | ZIP comment field                        |
| PPTX                    | Yes, through ZIP | ZIP comment field                        |
| JAR                     | Yes, through ZIP | ZIP comment field                        |
| APK                     | Yes, through ZIP | ZIP comment field                        |
| Other files             |         Fallback | Generic append                           |

The format detection is based on file signatures rather than file extensions.

For example, a file named `image.bin` containing a valid PNG signature can still be detected as PNG.

---

## Important: What ByteVeil Is and Is Not

ByteVeil performs **file-level steganography**.

It does **not** modify image pixels using techniques such as LSB steganography.

For supported formats, ByteVeil places its payload in an appropriate part of the file structure or appends it in a way that allows ByteVeil to recover it later.

This has an important consequence:

> ByteVeil is designed to survive ordinary file transfer, but it is not designed to survive media re-encoding or recompression.

For example:

* Sending a PNG as a file generally preserves the embedded payload.
* Copying a file from one disk to another preserves the payload.
* Uploading a file without modification generally preserves the payload.
* A platform that decodes and re-encodes a JPEG/PNG/video can destroy the payload.

Services that automatically compress or transcode media may therefore break the hidden data.

---

# Installation

## Option 1 — Build from source

ByteVeil is written in Go.

Clone the repository:

```bash
git clone https://github.com/batmanpriv/ByteVeil.git
cd ByteVeil
```

Build the binary:

```bash
go build -o byteveil .
```

Run it:

```bash
./byteveil
```

On Windows:

```powershell
go build -o byteveil.exe .
```

---

## Option 2 — Install with Go

If your Go environment is configured:

```bash
go install github.com/batmanpriv/ByteVeil@v1.0.0
```

Then run:

```bash
ByteVeil
```

Depending on your Go environment and platform, the installed binary directory may need to be added to your `PATH`.

---

# Quick Start

## Hide a file inside an image

```bash
byteveil encode \
  -in cover.png \
  -out secret.png \
  -data secret.pdf
```

This creates `secret.png` containing the hidden `secret.pdf`.

Extract it with:

```bash
byteveil decode \
  -in secret.png \
  -out extracted.pdf
```

---

## Hide a text message

```bash
byteveil encode \
  -in cover.jpg \
  -out message.jpg \
  -text "Meet me at 9pm."
```

Extract the message:

```bash
byteveil decode \
  -in message.jpg
```

If no output path is provided, ByteVeil uses the filename stored in the payload.

For text payloads, the stored filename is:

```text
message.txt
```

---

## Hide a file with encryption

```bash
byteveil encode \
  -in cover.png \
  -out secret.png \
  -data secret.pdf \
  -password "correct horse battery staple"
```

Extract it using the same password:

```bash
byteveil decode \
  -in secret.png \
  -out secret.pdf \
  -password "correct horse battery staple"
```

Without the correct password, the encrypted payload cannot be decrypted.

---

## Hide an entire folder

ByteVeil can accept a directory as the payload.

```bash
byteveil encode \
  -in cover.png \
  -out backup.png \
  -data ./my-folder \
  -password "my-password"
```

The directory is first packed into a TAR stream.

ByteVeil stores the directory name as metadata and marks the payload as a directory.

Extract it:

```bash
byteveil decode \
  -in backup.png \
  -out ./restored-folder \
  -password "my-password"
```

The contents are automatically extracted into the output directory.

---

# Command Reference

ByteVeil has two main commands:

```text
encode
decode
```

---

## `encode`

The `encode` command creates a stego file containing a hidden payload.

### Syntax

```bash
byteveil encode \
  -in <cover-file> \
  -out <output-file> \
  (-data <file-or-folder> | -text <message>) \
  [-password <password>]
```

### Arguments

| Argument    |    Required | Description                                    |
| ----------- | ----------: | ---------------------------------------------- |
| `-in`       |         Yes | Path to the cover file                         |
| `-out`      |         Yes | Path of the resulting stego file               |
| `-data`     | Conditional | File or directory to hide                      |
| `-text`     | Conditional | Text message to hide                           |
| `-password` |          No | Password used to enable AES-256-GCM encryption |

Exactly one of `-data` or `-text` must be supplied.

### Examples

File:

```bash
byteveil encode \
  -in cover.jpg \
  -out output.jpg \
  -data document.pdf
```

Text:

```bash
byteveil encode \
  -in cover.jpg \
  -out output.jpg \
  -text "This is a hidden message."
```

Encrypted file:

```bash
byteveil encode \
  -in cover.png \
  -out output.png \
  -data secret.zip \
  -password "strong-password"
```

Encrypted directory:

```bash
byteveil encode \
  -in cover.png \
  -out output.png \
  -data ./project \
  -password "strong-password"
```

---

## `decode`

The `decode` command extracts the hidden payload.

### Syntax

```bash
byteveil decode \
  -in <stego-file> \
  [-out <output>] \
  [-password <password>]
```

### Arguments

| Argument    | Required | Description                              |
| ----------- | -------: | ---------------------------------------- |
| `-in`       |      Yes | Stego file containing the hidden payload |
| `-out`      |       No | Output file or directory                 |
| `-password` |       No | Password for encrypted payloads          |

If `-out` is omitted, ByteVeil attempts to restore the original filename stored in the payload.

For directory payloads, the extracted data is restored as a directory.

### Examples

```bash
byteveil decode \
  -in output.png
```

Specify an output file:

```bash
byteveil decode \
  -in output.png \
  -out restored.pdf
```

Encrypted payload:

```bash
byteveil decode \
  -in output.png \
  -out restored.pdf \
  -password "strong-password"
```

Directory:

```bash
byteveil decode \
  -in backup.png \
  -out ./restored \
  -password "strong-password"
```

---

# How It Works

ByteVeil uses a small custom binary container to wrap the hidden payload.

The high-level process is:

```text
                   ENCODE

   file / folder / text
            │
            ▼
      gzip compression
            │
            ▼
   ┌────────────────────┐
   │ Password supplied? │
   └─────────┬──────────┘
        yes  │  no
             │
       ┌─────▼─────┐
       │ AES-256-  │
       │    GCM    │
       └─────┬─────┘
             │
             ▼
       ByteVeil payload
             │
             ▼
      format-specific
         embedding
             │
             ▼
         stego file
```

Decoding reverses the process:

```text
                 DECODE

              stego file
                  │
                  ▼
          locate STEGOv1 blob
                  │
                  ▼
          validate payload
                  │
                  ▼
          decrypt if required
                  │
                  ▼
           gzip decompression
                  │
                  ▼
        file / folder / text
```

---

# Payload Format

The outer ByteVeil blob begins with the following magic marker:

```text
STEGOv1\x00
```

It is followed by an 8-byte big-endian payload length and the inner payload.

The inner payload contains:

```text
STG1
version
flags
optional filename
optional salt
optional nonce
payload length
payload data
CRC32
```

The inner payload begins with:

```text
STG1
```

The current payload version is:

```text
1
```

This gives ByteVeil a versioned container format that can be extended in future releases.

---

# Compression

Payload data is compressed using gzip before being stored.

This applies to:

* files
* directories after TAR packaging
* text messages

Conceptually:

```text
original data
     │
     ▼
   gzip
     │
     ▼
compressed payload
```

Compression happens before encryption.

This is intentional because encrypted data generally does not compress effectively.

---

# Encryption

Encryption is optional.

If `-password` is provided, ByteVeil uses:

```text
PBKDF2-HMAC-SHA256
        │
        ▼
   200,000 iterations
        │
        ▼
    32-byte key
        │
        ▼
   AES-256-GCM
```

The implementation uses:

* PBKDF2-HMAC-SHA256
* 200,000 PBKDF2 iterations
* 32-byte derived key
* 16-byte random salt
* 12-byte random GCM nonce
* AES-GCM authenticated encryption

The salt and nonce are generated randomly for each encoded payload.

The encrypted data is authenticated by AES-GCM, so decryption fails when the password is incorrect or the ciphertext has been modified.

---

# Integrity Checking

ByteVeil also stores a CRC32 checksum for the inner payload.

During decoding, the checksum is verified before the payload is processed.

This helps detect:

* truncated files
* corrupted payloads
* accidental modification
* incomplete transfers

For encrypted payloads, AES-GCM additionally provides cryptographic authentication.

The CRC32 should therefore be understood as an **integrity/corruption check**, not as a replacement for authenticated encryption.

---

# Directory Support

When `-data` points to a directory, ByteVeil does not attempt to hide the directory directly.

Instead, it:

1. Walks the directory.
2. Creates a TAR stream.
3. Preserves regular files.
4. Preserves subdirectories.
5. Preserves symbolic links.
6. Stores the directory's original base name.
7. Marks the payload as a directory.
8. Compresses the resulting TAR stream.
9. Optionally encrypts it.
10. Embeds it into the cover file.

During decoding, ByteVeil recognizes the directory flag and extracts the TAR archive into the requested output directory.

---

# Security Considerations

ByteVeil provides encryption when a password is supplied, but it should not be treated as a complete secure file-sharing system.

## Password security matters

The security of an encrypted payload depends heavily on the password.

Avoid weak passwords such as:

```text
123456
password
qwerty
hunter2
```

Prefer a long, unique passphrase or randomly generated password.

For example:

```text
correct-horse-battery-staple-7f3a91
```

Do not commit passwords to Git.

---

## Encryption is optional

Without `-password`, the payload is only compressed and hidden.

Anyone who knows how ByteVeil stores its payload can potentially extract the data.

For sensitive information, always use encryption:

```bash
byteveil encode \
  -in cover.png \
  -out secret.png \
  -data secret.pdf \
  -password "your-strong-password"
```

---

## Steganography is not encryption

Hiding data and encrypting data solve different problems.

```text
Steganography
    = hides the existence/location of data

Encryption
    = protects the contents of data
```

ByteVeil combines both when a password is supplied:

```text
hidden + encrypted
```

But the presence of a ByteVeil payload should not be considered mathematically undetectable.

---

# Format-Specific Behavior

## JPEG

JPEG payloads are appended after the existing JPEG data.

This allows the original JPEG content to remain unchanged while ByteVeil adds its payload.

```text
[JPEG data][ByteVeil blob]
```

---

## PNG

PNG files use a custom ancillary chunk:

```text
stEg
```

The ByteVeil blob is inserted before the PNG `IEND` chunk.

Conceptually:

```text
PNG signature
PNG chunks
...
stEg chunk
IEND
```

The chunk has its own CRC.

---

## MP4 / ISO-BMFF

ByteVeil adds an ISO-BMFF `free` box containing the ByteVeil blob.

```text
[existing MP4 boxes][free box containing ByteVeil payload]
```

---

## MP3

For MP3 files with a compatible ID3v2 tag, ByteVeil stores the payload inside a `PRIV` frame.

The private frame owner identifier is:

```text
stego-tool
```

If a compatible ID3v2 tag is not present, ByteVeil creates an ID3v2.3 header containing the private frame.

---

## ZIP-based formats

ZIP-family files use the ZIP End of Central Directory comment field.

This also applies to formats built on ZIP containers, such as:

* `.zip`
* `.docx`
* `.xlsx`
* `.pptx`
* `.jar`
* `.apk`

### Important limitation

The ZIP comment field is limited to:

```text
65,535 bytes
```

Therefore ZIP-based cover files have a hard payload-size ceiling.

For larger payloads, use another cover format.

---

## PDF

For PDF files, ByteVeil appends the payload after the existing `%%EOF` marker.

Some strict PDF readers or processors may reject or modify such files, especially when dealing with unusually large appended payloads.

---

## Generic fallback

For unrecognized formats, ByteVeil falls back to appending the ByteVeil blob to the original file.

This makes the tool usable with arbitrary binary files, but compatibility with external applications is not guaranteed.

---

# File Transfer and Re-Encoding

One of the most important practical limitations of ByteVeil is the difference between **file transfer** and **media re-encoding**.

### Usually safe

```text
Local copy
     ↓
USB drive
     ↓
Cloud storage as a file
     ↓
Archive
     ↓
"Send as file"
```

The bytes remain unchanged, so the embedded payload remains available.

### Usually unsafe

```text
JPEG
  ↓
Social media upload
  ↓
JPEG recompression
  ↓
New JPEG
```

or:

```text
Video
  ↓
Platform transcoding
  ↓
New encoded video
```

Re-encoding creates a new file representation and may remove the ByteVeil payload.

This is a fundamental limitation of file-level embedding rather than a bug specific to ByteVeil.

---

# Examples

## Hide a PDF in a PNG

```bash
byteveil encode \
  -in photo.png \
  -out photo-secret.png \
  -data document.pdf \
  -password "my-secret"
```

Extract:

```bash
byteveil decode \
  -in photo-secret.png \
  -out document.pdf \
  -password "my-secret"
```

---

## Hide a message in a JPEG

```bash
byteveil encode \
  -in photo.jpg \
  -out photo-hidden.jpg \
  -text "Nothing to see here."
```

Extract:

```bash
byteveil decode \
  -in photo-hidden.jpg
```

---

## Hide a project directory

```bash
byteveil encode \
  -in cover.png \
  -out project.png \
  -data ./my-project \
  -password "project-password"
```

Restore:

```bash
byteveil decode \
  -in project.png \
  -out ./restored-project \
  -password "project-password"
```

---

## Hide a ZIP archive

```bash
byteveil encode \
  -in image.png \
  -out image-with-data.png \
  -data archive.zip
```

---

# Error Handling

ByteVeil validates the embedded payload before extracting it.

Examples of errors include:

```text
no hidden payload marker found in this file
```

when no ByteVeil payload is present.

```text
checksum mismatch: payload is corrupted or truncated
```

when the inner payload fails its CRC32 check.

```text
this payload is encrypted; a password is required
```

when an encrypted payload is decoded without a password.

```text
decryption failed: wrong password or corrupted data
```

when AES-GCM authentication fails.

```text
declared payload length exceeds file size
```

when the outer payload length indicates truncated or malformed data.

---

# Design Goals

ByteVeil intentionally focuses on a small set of goals:

### 1. Simplicity

The project is a single Go CLI with two primary operations:

```text
encode
decode
```

### 2. Portability

The implementation uses Go's standard library and is designed to build as a standalone binary.

### 3. Format flexibility

The payload container is independent from the cover format.

The same payload structure can be embedded into different file types using different embedding strategies.

### 4. Optional encryption

Users can choose between:

```text
compression + hiding
```

or:

```text
compression + encryption + hiding
```

### 5. Recoverability

The custom payload contains enough metadata to restore:

* original filename
* directory/file state
* payload length
* encryption parameters

---

# Limitations

ByteVeil v1.0.0 has several important limitations.

## No media re-encoding resistance

ByteVeil is not designed to survive:

* JPEG recompression
* PNG reprocessing
* video transcoding
* audio re-encoding
* image optimization that rewrites the file
* other transformations that change the underlying bytes

---

## ZIP payload size limit

ZIP-based covers are limited by the ZIP comment maximum of 65,535 bytes.

This affects:

```text
.zip
.docx
.xlsx
.pptx
.jar
.apk
```

and other ZIP-based containers.

---

## Generic embedding is not universally compatible

For unknown file types, ByteVeil simply appends its payload.

A receiving application may reject a file containing unexpected trailing bytes.

---

## No password recovery

If an encrypted payload's password is lost, ByteVeil has no password recovery mechanism.

There is no backdoor or recovery key.

---

## No steganographic statistical analysis resistance guarantee

ByteVeil is not intended to provide protection against dedicated steganalysis.

Its goal is practical file-level hiding, not provable undetectability.

---

# Compatibility

ByteVeil is written in Go and uses the Go standard library.

The implementation currently relies on standard packages including:

* `archive/tar`
* `bytes`
* `compress/gzip`
* `crypto/aes`
* `crypto/cipher`
* `crypto/hmac`
* `crypto/rand`
* `crypto/sha256`
* `encoding/binary`
* `hash/crc32`
* `io`
* `os`
* `path/filepath`
* `strings`

No third-party runtime dependency is required by the current implementation.

---

# Project Structure

The current project is intentionally compact.

The main implementation contains:

```text
ByteVeil
└── main.go
```

The code is organized around several responsibilities:

```text
Payload construction
├── buildInner
├── parseInner
├── wrapBlob
└── findBlob

Cryptography
└── pbkdf2

Format detection
└── sniffFormat

Embedding
├── embedJPEG
├── embedPNG
├── embedMP4
├── embedMP3
├── embedGeneric
└── embedZIP

Directory handling
├── tarDir
└── untarToDir

CLI operations
├── encode
├── decode
├── usage
└── main
```

---

# Version 1.0.0

This repository represents the **ByteVeil v1.0.0** release.

The v1 payload format uses:

```text
Outer magic:
STEGOv1\x00

Inner magic:
STG1

Payload version:
1
```

The initial release establishes the basic container format, encryption model, directory support, payload integrity checking, and multi-format embedding system.

---

# Roadmap

Potential future improvements may include:

* More format-specific embedding strategies
* Better handling of additional container formats
* Streaming large payloads instead of loading everything into memory
* More extensive automated tests
* Cross-platform release binaries
* Improved CLI ergonomics
* Payload inspection commands
* Metadata inspection
* Configurable compression
* Additional key derivation options
* Better large-file support
* Improved format validation
* Documentation and examples
* Benchmarking

The roadmap is intentionally not a commitment to a specific release schedule.

---

# Security Reporting

If you discover a security vulnerability in ByteVeil, please avoid publicly disclosing the details before the issue has been reviewed.

For security-related issues, open a private security report through GitHub if repository security reporting is enabled.

Otherwise, contact the maintainer through the GitHub profile:

**[@batmanpriv](https://github.com/batmanpriv)**

Please include:

* A clear description of the issue
* Steps to reproduce it
* Affected functionality
* Potential security impact
* A suggested mitigation, if available

Do not include real passwords, private files, credentials, tokens, or other sensitive information in an issue.

---

# Contributing

Contributions are welcome.

If you want to improve ByteVeil:

1. Fork the repository.
2. Create a feature branch.
3. Make your changes.
4. Test the changes locally.
5. Keep the implementation focused and idiomatic Go.
6. Open a pull request with a clear description.

Example:

```bash
git clone https://github.com/batmanpriv/ByteVeil.git
cd ByteVeil

git checkout -b feature/my-feature

# Make changes...

go build .
go test ./...

git add .
git commit -m "Add my feature"
git push origin feature/my-feature
```

For format-specific changes, please consider compatibility with existing files and the possibility of malformed or truncated input.

---

# Responsible Use

ByteVeil is a general-purpose data-hiding and encryption utility.

It can be useful for:

* privacy research
* security research
* file-format experimentation
* CTFs
* educational purposes
* controlled data-transfer experiments
* legitimate personal data protection
* understanding file/container structures

Use ByteVeil responsibly and only on files and systems you are authorized to access or modify.

---

# License

ByteVeil is distributed under the license included in this repository.

See [`LICENSE`](LICENSE) for the complete license text.

---

# Author

Created and maintained by **batmanpriv**.

GitHub:

**https://github.com/batmanpriv**

Project:

**https://github.com/batmanpriv/ByteVeil**

---

<p align="center">
  <strong>ByteVeil</strong>
  <br>
  Hide the data. Keep the file.
</p>
