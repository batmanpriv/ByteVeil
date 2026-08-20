# ByteVeil

<p align="center">
  <strong>Hide data in plain sight.</strong>
  <img src="https://github.com/user-attachments/assets/626f89a5-d5f6-49dd-8633-8b719a55ba88">
</p>

<p align="center">
  A lightweight, file-based steganography CLI written in Go.
  Hide files, folders, or text inside ordinary files with optional AES-256-GCM encryption.
</p>

<p align="center">
  <a href="https://github.com/batmanpriv/ByteVeil/releases/tag/v1.0.0">
    <img src="https://img.shields.io/badge/version-1.0.0-blue.svg" alt="Version">
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/Go-1.20%2B-00ADD8.svg" alt="Go">
  </a>
  <a href="https://github.com/batmanpriv/ByteVeil">
    <img src="https://img.shields.io/github/stars/batmanpriv/ByteVeil?style=flat" alt="GitHub Stars">
  </a>
  <a href="https://github.com/topics/steganography">
    <img src="https://img.shields.io/badge/Topic-Steganography-purple.svg" alt="Steganography">
  </a>
  <a href="https://github.com/topics/golang">
    <img src="https://img.shields.io/badge/Topic-Go-00ADD8.svg" alt="Go Topic">
  </a>
  <a href="https://github.com/topics/encryption">
    <img src="https://img.shields.io/badge/Topic-Encryption-red.svg" alt="Encryption">
  </a>
  <a href="https://github.com/batmanpriv/ByteVeil/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/License-MIT-green.svg" alt="License">
  </a>
</p>

---

## Overview

**ByteVeil** is a command-line tool for file-based steganography.

It allows you to hide:

* Individual files
* Entire directories
* Text messages

inside a **cover file**.

Payloads are compressed with gzip before being stored. If a password is provided, ByteVeil additionally encrypts the compressed payload using **AES-256-GCM** with a key derived through **PBKDF2-HMAC-SHA256**.

ByteVeil detects the cover file format and uses a format-specific embedding strategy when supported. For unrecognized formats, it falls back to a generic append strategy.

The result is a single file that can later be decoded by ByteVeil to recover the original payload.

```text
                  ENCODE

 File / Folder / Text
          │
          ▼
   ┌──────────────┐
   │    gzip      │
   │ compression  │
   └──────┬───────┘
          │
          ▼
   ┌──────────────┐
   │ Optional     │
   │ AES-256-GCM  │
   │ encryption   │
   └──────┬───────┘
          │
          ▼
   ┌──────────────┐
   │  ByteVeil    │
   │   payload    │
   └──────┬───────┘
          │
          ▼
     Cover File
          │
          ▼
      Stego File
```

---

## Features

* Hide files inside other files
* Hide entire directories
* Hide plain-text messages
* Optional AES-256-GCM encryption
* PBKDF2-HMAC-SHA256 password-based key derivation
* 200,000 PBKDF2 iterations
* Random cryptographic salt
* Random AES-GCM nonce
* gzip compression before encryption
* CRC32 payload integrity checking
* Original filename preservation
* Automatic file/directory detection during extraction
* Format-aware embedding for multiple common formats
* Generic fallback for unsupported formats
* Single Go binary
* Uses only the Go standard library
* Works on Windows, Linux, and macOS when built for the target platform

---

# Supported Formats

ByteVeil recognizes the following file formats:

| Format         | Detection | Embedding Method                  |
| -------------- | --------: | --------------------------------- |
| JPEG           |       Yes | Appended after JPEG data          |
| PNG            |       Yes | Custom `stEg` chunk before `IEND` |
| MP4 / ISO-BMFF |       Yes | Appended `free` box               |
| MP3            |       Yes | `PRIV` frame in ID3v2             |
| GIF            |       Yes | Appended after file data          |
| BMP            |       Yes | Appended after bitmap data        |
| RIFF           |       Yes | Generic append                    |
| WAV            |       Yes | Generic append                    |
| AVI            |       Yes | Generic append                    |
| WEBP           |       Yes | Generic append                    |
| OGG            |       Yes | Generic append                    |
| PDF            |       Yes | Appended after `%%EOF`            |
| ZIP            |       Yes | ZIP EOCD comment                  |
| DOCX           |       Yes | ZIP EOCD comment                  |
| XLSX           |       Yes | ZIP EOCD comment                  |
| PPTX           |       Yes | ZIP EOCD comment                  |
| JAR            |       Yes | ZIP EOCD comment                  |
| APK            |       Yes | ZIP EOCD comment                  |
| Other files    |  Fallback | Generic append                    |

Format detection is performed using file signatures rather than filename extensions.

For example, a file named:

```text
photo.dat
```

can still be recognized as PNG if its binary signature is a valid PNG signature.

---

# Installation

## Build from Source

Clone the repository:

```bash
git clone https://github.com/batmanpriv/ByteVeil.git
```

Enter the project directory:

```bash
cd ByteVeil
```

Build ByteVeil:

```bash
go build -o ByteVeil .
```

On Windows:

```powershell
go build -o ByteVeil.exe .
```

You can then run the binary directly.

Linux/macOS:

```bash
./ByteVeil
```

Windows:

```powershell
.\ByteVeil.exe
```

---

## Install with Go

You can also install ByteVeil directly using `go install`:

```bash
go install github.com/batmanpriv/ByteVeil@v1.0.0
```

Make sure your Go binary directory is included in your system `PATH`.

Then run:

```bash
ByteVeil
```

---

# Quick Start

## Hide a File

Hide `secret.pdf` inside `cover.png`:

```bash
ByteVeil encode -in cover.png -out secret.png -data secret.pdf
```

Extract it:

```bash
ByteVeil decode -in secret.png -out extracted.pdf
```

---

## Hide a Text Message

Hide a text message inside a JPEG:

```bash
ByteVeil encode -in cover.jpg -out message.jpg -text "Meet me at 9pm."
```

Extract it:

```bash
ByteVeil decode -in message.jpg
```

If no output path is specified, ByteVeil uses the filename stored inside the payload.

Text messages are stored with:

```text
message.txt
```

as their default filename.

---

## Hide an Encrypted File

Provide a password to enable encryption:

```bash
ByteVeil encode -in cover.png -out secret.png -data secret.pdf -password "your-strong-password"
```

Extract it using the same password:

```bash
ByteVeil decode -in secret.png -out secret.pdf -password "your-strong-password"
```

Without the correct password, the encrypted payload cannot be successfully decrypted.

---

## Hide an Entire Directory

ByteVeil can hide an entire directory:

```bash
ByteVeil encode -in cover.png -out backup.png -data ./my-folder -password "your-strong-password"
```

The directory is packed into a TAR stream before compression and optional encryption.

Extract it:

```bash
ByteVeil decode -in backup.png -out ./restored-folder -password "your-strong-password"
```

ByteVeil automatically detects that the payload represents a directory and extracts it into the specified output directory.

---

# Command Reference

ByteVeil provides two commands:

```text
encode
decode
```

---

## `encode`

Creates a new file containing a hidden ByteVeil payload.

### Syntax

```text
ByteVeil encode -in <cover> -out <output> (-data <file-or-folder> | -text <message>) [-password <password>]
```

### Options

| Flag        |    Required | Description                                    |
| ----------- | ----------: | ---------------------------------------------- |
| `-in`       |         Yes | Path to the cover file                         |
| `-out`      |         Yes | Path to the resulting stego file               |
| `-data`     | Conditional | File or directory to hide                      |
| `-text`     | Conditional | Text message to hide                           |
| `-password` |          No | Password used to enable AES-256-GCM encryption |

Exactly one of `-data` or `-text` must be provided.

### File Example

```bash
ByteVeil encode -in cover.png -out output.png -data secret.pdf
```

### Text Example

```bash
ByteVeil encode -in cover.jpg -out output.jpg -text "This is a hidden message."
```

### Encrypted Example

```bash
ByteVeil encode -in cover.png -out output.png -data secret.pdf -password "strong-password"
```

### Directory Example

```bash
ByteVeil encode -in cover.png -out output.png -data ./my-folder -password "strong-password"
```

---

## `decode`

Extracts a hidden ByteVeil payload.

### Syntax

```text
ByteVeil decode -in <stego-file> [-out <output>] [-password <password>]
```

### Options

| Flag        | Required | Description                              |
| ----------- | -------: | ---------------------------------------- |
| `-in`       |      Yes | Stego file containing the hidden payload |
| `-out`      |       No | Output file or directory                 |
| `-password` |       No | Password for encrypted payloads          |

### Basic Example

```bash
ByteVeil decode -in output.png
```

### Specify Output

```bash
ByteVeil decode -in output.png -out restored.pdf
```

### Encrypted Payload

```bash
ByteVeil decode -in output.png -out restored.pdf -password "strong-password"
```

### Directory Payload

```bash
ByteVeil decode -in output.png -out ./restored -password "strong-password"
```

---

# How It Works

ByteVeil uses a custom binary payload container.

The encoding pipeline is:

```text
Original Payload
      │
      ├── File
      ├── Folder
      └── Text
      │
      ▼
 Directory → TAR
      │
      ▼
    gzip
      │
      ▼
 Optional AES-256-GCM
      │
      ▼
 ByteVeil Inner Payload
      │
      ▼
 ByteVeil Outer Blob
      │
      ▼
 Format-Specific Embedding
      │
      ▼
 Stego File
```

The decoding pipeline reverses these operations:

```text
Stego File
    │
    ▼
Locate ByteVeil Blob
    │
    ▼
Validate Payload
    │
    ▼
Decrypt if Required
    │
    ▼
gzip Decompression
    │
    ▼
TAR Extraction if Directory
    │
    ▼
Original Payload
```

---

# ByteVeil Payload Format

The outer ByteVeil blob begins with the following magic marker:

```text
STEGOv1\x00
```

The outer structure is:

```text
+-------------------+
| STEGOv1\x00       |
+-------------------+
| Payload Length    |
| 8 bytes, BE       |
+-------------------+
| Inner Payload     |
+-------------------+
```

The inner payload begins with:

```text
STG1
```

The current payload version is:

```text
1
```

The inner payload contains metadata describing the stored data.

Conceptually:

```text
+-----------------------+
| STG1                  |
+-----------------------+
| Version               |
+-----------------------+
| Flags                 |
+-----------------------+
| Filename (optional)   |
+-----------------------+
| Salt (optional)       |
+-----------------------+
| Nonce (optional)      |
+-----------------------+
| Data Length           |
+-----------------------+
| Compressed/Encrypted |
| Payload               |
+-----------------------+
| CRC32                 |
+-----------------------+
```

The payload format is versioned so that future versions can introduce compatible format changes.

---

# Compression

ByteVeil compresses payload data using **gzip** before storing it.

For normal files and text:

```text
Original Data
     ↓
   gzip
     ↓
Compressed Data
```

For directories:

```text
Directory
    ↓
   TAR
    ↓
TAR Stream
    ↓
  gzip
    ↓
Compressed TAR
```

Compression happens before encryption.

This is important because encrypted data generally does not compress effectively.

---

# Encryption

Encryption is optional.

When `-password` is provided, ByteVeil derives a 256-bit encryption key using:

```text
PBKDF2-HMAC-SHA256
```

with:

```text
Iterations: 200,000
Key length: 32 bytes
Salt length: 16 bytes
```

The resulting key is used with:

```text
AES-256-GCM
```

A fresh random salt and nonce are generated for every encoded payload.

The encryption flow is:

```text
Password
   │
   ▼
PBKDF2-HMAC-SHA256
   │
   ├── 200,000 iterations
   ├── 16-byte random salt
   └── 32-byte derived key
   │
   ▼
AES-256-GCM
   │
   └── 12-byte random nonce
```

The salt and nonce are stored alongside the encrypted payload because they are not secrets.

The password itself is never stored in the ByteVeil payload.

---

# Integrity Protection

ByteVeil stores a CRC32 checksum for the inner payload.

During decoding, the checksum is verified before the payload is processed.

This allows ByteVeil to detect accidental corruption or truncation.

For encrypted payloads, AES-GCM additionally provides authenticated encryption.

Therefore:

```text
CRC32
  → detects payload corruption/truncation

AES-GCM authentication
  → detects invalid/tampered encrypted data
```

CRC32 is **not** a cryptographic security mechanism.

---

# Directory Handling

When `-data` points to a directory, ByteVeil creates a TAR archive in memory.

The TAR archive can contain:

* Regular files
* Directories
* Symbolic links

The original directory name is stored in the payload metadata.

For example:

```text
my-project/
├── main.go
├── README.md
├── src/
│   └── app.go
└── assets/
    └── config.json
```

is packed into a TAR stream before compression.

When decoded, ByteVeil recognizes the directory flag and extracts the archive.

---

# Format-Specific Embedding

## JPEG

JPEG payloads are appended after the original JPEG data.

Conceptually:

```text
[JPEG DATA][ByteVeil BLOB]
```

The JPEG content itself is not modified by the embedding operation.

---

## PNG

PNG payloads are stored in a custom ancillary chunk:

```text
stEg
```

The chunk is inserted before the PNG `IEND` chunk.

Conceptually:

```text
PNG Signature
     │
     ├── IHDR
     ├── ...
     ├── stEg
     │     └── ByteVeil Blob
     └── IEND
```

The custom chunk also has its own CRC.

---

## MP4 / ISO-BMFF

For MP4-family files, ByteVeil adds a `free` box containing the ByteVeil payload.

Conceptually:

```text
[Existing MP4 Boxes][free][ByteVeil Blob]
```

---

## MP3

For MP3 files with a compatible ID3v2 tag, ByteVeil stores the payload inside a `PRIV` frame.

The private owner identifier used by the implementation is:

```text
stego-tool
```

If a suitable ID3v2 tag is not present, ByteVeil creates an ID3v2.3 tag containing the private frame.

---

## GIF

GIF payloads are appended after the existing file data.

---

## BMP

BMP payloads are appended after the existing bitmap data.

---

## RIFF / WAV / AVI / WEBP

These formats are recognized as RIFF-family files.

The current implementation uses a generic append strategy for these formats.

---

## OGG

OGG files are recognized and use the generic append strategy.

---

## PDF

PDF payloads are appended after the existing `%%EOF` marker.

This means strict PDF readers or processors may not always accept a modified PDF containing an appended ByteVeil payload.

---

## ZIP

ZIP-based files use the ZIP End of Central Directory comment field.

This applies to:

```text
.zip
.docx
.xlsx
.pptx
.jar
.apk
```

and other ZIP-based formats.

The ZIP comment field has a maximum size of:

```text
65,535 bytes
```

Therefore ZIP-based cover files have a hard payload-size limitation.

---

## Generic Fallback

If ByteVeil does not recognize the cover file format, it falls back to generic append embedding.

Conceptually:

```text
[Original File][ByteVeil Blob]
```

This allows ByteVeil to work with arbitrary binary files, but compatibility with third-party applications is not guaranteed.

---

# File Transfer vs. Re-Encoding

ByteVeil operates at the file-byte level.

As a result, it can survive operations that preserve the original bytes, but it generally cannot survive operations that decode and re-encode the media.

## Usually Safe

Examples of operations that normally preserve the payload:

* Copying the file
* Moving the file
* Renaming the file
* Storing it on a filesystem
* Sending it as an unmodified file
* Copying it through a normal file-transfer mechanism
* Putting it into an archive without rewriting its contents

For example:

```text
ByteVeil File
     │
     ▼
"Send as File"
     │
     ▼
Same Bytes
     │
     ▼
Payload Still Present
```

## Usually Not Safe

Media transcoding or recompression can destroy the payload.

For example:

```text
ByteVeil JPEG
     │
     ▼
Platform decodes JPEG
     │
     ▼
Platform recompresses JPEG
     │
     ▼
New JPEG
     │
     ▼
ByteVeil payload may be gone
```

The same applies to video and audio transcoding.

This is a fundamental limitation of file-level embedding.

---

# Security Considerations

## Steganography Is Not Encryption

Steganography and encryption provide different protections.

Steganography attempts to hide the presence or location of information.

Encryption protects the contents of the information.

Without a password:

```text
Hidden + Compressed
```

With a password:

```text
Hidden + Compressed + Encrypted
```

For sensitive payloads, use a strong password.

---

## Use Strong Passwords

Avoid weak passwords such as:

```text
password
123456
qwerty
hunter2
```

Prefer a long, unique passphrase.

For example:

```text
correct-horse-battery-staple-7f3a91
```

Do not commit passwords to Git repositories.

Do not place passwords in shell scripts that may be committed to source control.

---

## Password Recovery

ByteVeil does not provide password recovery.

If an encrypted payload is created with a lost password, there is no recovery mechanism built into the tool.

---

## Steganalysis

ByteVeil does not claim to provide undetectable steganography.

It uses file-level embedding rather than modifying pixel values or audio samples, but the presence of a ByteVeil payload may still be detectable by someone inspecting the file structure or searching for the ByteVeil magic marker.

The goal is practical data hiding, not provable invisibility.

---

# Limitations

## Media Re-Encoding

ByteVeil does not guarantee survival through:

* JPEG recompression
* PNG rewriting
* Video transcoding
* Audio re-encoding
* Image optimization
* File conversion
* Any operation that creates a new representation of the original media

---

## ZIP Payload Limit

ZIP-based covers are limited by the maximum ZIP comment size:

```text
65,535 bytes
```

This includes:

* ZIP
* DOCX
* XLSX
* PPTX
* JAR
* APK

For larger payloads, use another supported cover format.

---

## Memory Usage

The current implementation reads and constructs payload data in memory.

For very large files or directories, memory usage can therefore become significant.

The current implementation is not a streaming steganography pipeline.

---

## Generic Embedding

Unknown formats use generic append embedding.

Although this makes ByteVeil flexible, applications that strictly validate file length or reject trailing data may refuse to open such files.

---

## External Application Compatibility

ByteVeil guarantees that its own decoder can recover a valid payload from an appropriately preserved ByteVeil file.

It does **not** guarantee that every third-party application will accept every modified cover file.

---

# Error Handling

ByteVeil validates its payload before extraction.

Typical errors include:

### No payload

```text
no hidden payload marker found in this file
```

This means ByteVeil could not find the `STEGOv1` marker.

### Corrupted payload

```text
checksum mismatch: payload is corrupted or truncated
```

The CRC32 check failed.

### Missing password

```text
this payload is encrypted; a password is required
```

The payload is encrypted but no password was supplied.

### Wrong password or corrupted ciphertext

```text
decryption failed: wrong password or corrupted data
```

AES-GCM authentication failed.

### Truncated payload

```text
declared payload length exceeds file size (truncated file?)
```

The file is shorter than the length declared by the ByteVeil container.

---

# Examples

## File → PNG

```bash
ByteVeil encode -in photo.png -out hidden.png -data secret.pdf
```

Extract:

```bash
ByteVeil decode -in hidden.png -out secret.pdf
```

---

## Encrypted File → PNG

```bash
ByteVeil encode -in photo.png -out hidden.png -data secret.pdf -password "strong-password"
```

Extract:

```bash
ByteVeil decode -in hidden.png -out secret.pdf -password "strong-password"
```

---

## Text → JPEG

```bash
ByteVeil encode -in photo.jpg -out hidden.jpg -text "This message is hidden."
```

Extract:

```bash
ByteVeil decode -in hidden.jpg
```

---

## Folder → PNG

```bash
ByteVeil encode -in cover.png -out backup.png -data ./project -password "strong-password"
```

Extract:

```bash
ByteVeil decode -in backup.png -out ./restored-project -password "strong-password"
```

---

## MP3

```bash
ByteVeil encode -in music.mp3 -out hidden.mp3 -data secret.txt -password "strong-password"
```

Extract:

```bash
ByteVeil decode -in hidden.mp3 -out secret.txt -password "strong-password"
```

---

## ZIP

Because ZIP comments have a hard size limit, keep the payload small:

```bash
ByteVeil encode -in archive.zip -out hidden.zip -text "Hidden ZIP message"
```

Extract:

```bash
ByteVeil decode -in hidden.zip
```

---

# Usage Summary

```text
Encode a file:
ByteVeil encode -in <cover> -out <output> -data <file>

Encode a folder:
ByteVeil encode -in <cover> -out <output> -data <folder>

Encode text:
ByteVeil encode -in <cover> -out <output> -text "<message>"

Encrypt:
ByteVeil encode -in <cover> -out <output> -data <file> -password "<password>"

Decode:
ByteVeil decode -in <stego-file>

Decode to a specific file:
ByteVeil decode -in <stego-file> -out <output>

Decode encrypted payload:
ByteVeil decode -in <stego-file> -out <output> -password "<password>"
```

---

# Build and Development

Clone the repository:

```bash
git clone https://github.com/batmanpriv/ByteVeil.git
```

Enter the directory:

```bash
cd ByteVeil
```

Build:

```bash
go build .
```

Run:

```bash
go run . encode -in cover.png -out output.png -text "Hello from ByteVeil"
```

Build a named binary:

Linux/macOS:

```bash
go build -o ByteVeil .
```

Windows:

```powershell
go build -o ByteVeil.exe .
```

---

# Project Structure

The current implementation is intentionally compact.

```text
ByteVeil/
└── main.go
```

The main source file contains the following logical components:

```text
Payload
├── buildInner
├── parseInner
├── wrapBlob
└── findBlob

Cryptography
└── pbkdf2

Format Detection
└── sniffFormat

Embedding
├── embedJPEG
├── embedPNG
├── embedMP4
├── embedMP3
├── embedGeneric
└── embedZIP

Directory Handling
├── tarDir
└── untarToDir

CLI
├── encode
├── decode
├── usage
└── main
```

---

# Dependencies

ByteVeil currently uses the Go standard library only.

Important packages include:

```text
archive/tar
bytes
compress/gzip
crypto/aes
crypto/cipher
crypto/hmac
crypto/rand
crypto/sha256
encoding/binary
errors
flag
fmt
hash/crc32
io
os
path/filepath
strings
```

No external runtime dependency is required.

---

# Version

Current release:

```text
v1.0.0
```

Payload format:

```text
STEGOv1
STG1
version 1
```

Versioning is included in the payload format to allow future versions of ByteVeil to evolve the container format.

