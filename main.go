
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var blobMagic = []byte("ByteVeilv1\x00")

const (
	innerMagic     = "STG1"
	pbkdf2Iters    = 200000
	pbkdf2KeyLen   = 32 
	saltLen        = 16
	nonceLen       = 12
	flagEncrypted  = 1 << 0
	flagHasName    = 1 << 1
	flagIsDir      = 1 << 2
	pngChunkType   = "stEg" 
	mp4BoxFree     = "free"
	id3PrivOwnerID = "ByteVeil-tool"
	zipCommentMax  = 0xFFFF 
)

func pbkdf2(password, salt []byte, iterations, keyLen int) []byte {
	hashLen := sha256.Size
	numBlocks := (keyLen + hashLen - 1) / hashLen
	var dk []byte

	for block := 1; block <= numBlocks; block++ {
		mac := hmac.New(sha256.New, password)
		mac.Write(salt)
		var be4 [4]byte
		binary.BigEndian.PutUint32(be4[:], uint32(block))
		mac.Write(be4[:])
		u := mac.Sum(nil)

		t := make([]byte, len(u))
		copy(t, u)

		prev := u
		for i := 1; i < iterations; i++ {
			mac2 := hmac.New(sha256.New, password)
			mac2.Write(prev)
			prev = mac2.Sum(nil)
			for j := range t {
				t[j] ^= prev[j]
			}
		}
		dk = append(dk, t...)
	}
	return dk[:keyLen]
}

func buildInner(payload []byte, password string, filename string, isDir bool) ([]byte, error) {
	var compressed bytes.Buffer
	gw := gzip.NewWriter(&compressed)
	if _, err := gw.Write(payload); err != nil {
		return nil, fmt.Errorf("gzip write: %w", err)
	}
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("gzip close: %w", err)
	}

	var flags byte
	encrypted := password != ""
	if encrypted {
		flags |= flagEncrypted
	}
	if filename != "" {
		flags |= flagHasName
	}
	if isDir {
		flags |= flagIsDir
	}

	var buf bytes.Buffer
	buf.WriteString(innerMagic)
	buf.WriteByte(1) 
	buf.WriteByte(flags)

	if filename != "" {
		nameBytes := []byte(filename)
		if len(nameBytes) > 0xFFFF {
			return nil, errors.New("filename too long")
		}
		var nl [2]byte
		binary.BigEndian.PutUint16(nl[:], uint16(len(nameBytes)))
		buf.Write(nl[:])
		buf.Write(nameBytes)
	}

	var dataToStore []byte
	if encrypted {
		salt := make([]byte, saltLen)
		if _, err := rand.Read(salt); err != nil {
			return nil, err
		}
		key := pbkdf2([]byte(password), salt, pbkdf2Iters, pbkdf2KeyLen)

		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, err
		}
		nonce := make([]byte, nonceLen)
		if _, err := rand.Read(nonce); err != nil {
			return nil, err
		}
		ciphertext := aesgcm.Seal(nil, nonce, compressed.Bytes(), nil)

		buf.Write(salt)
		buf.Write(nonce)
		dataToStore = ciphertext
	} else {
		dataToStore = compressed.Bytes()
	}

	var dl [8]byte
	binary.BigEndian.PutUint64(dl[:], uint64(len(dataToStore)))
	buf.Write(dl[:])
	buf.Write(dataToStore)

	sum := crc32.ChecksumIEEE(buf.Bytes())
	var cb [4]byte
	binary.BigEndian.PutUint32(cb[:], sum)
	buf.Write(cb[:])

	return buf.Bytes(), nil
}

func parseInner(inner []byte, password string) (payload []byte, filename string, isDir bool, err error) {
	if len(inner) < 4+1+1+4 {
		return nil, "", false, errors.New("payload too short / corrupted")
	}
	if len(inner) < 4 {
		return nil, "", false, errors.New("payload too short")
	}
	if string(inner[0:4]) != innerMagic {
		return nil, "", false, errors.New("bad inner magic (wrong password won't cause this, but corrupted data will)")
	}

	if len(inner) < 4 {
		return nil, "", false, errors.New("truncated payload")
	}
	body := inner[:len(inner)-4]
	wantCRC := binary.BigEndian.Uint32(inner[len(inner)-4:])
	gotCRC := crc32.ChecksumIEEE(body)
	if wantCRC != gotCRC {
		return nil, "", false, errors.New("checksum mismatch: payload is corrupted or truncated")
	}

	pos := 4
	_ = inner[pos] 
	version := inner[pos]
	pos++
	if version != 1 {
		return nil, "", false, fmt.Errorf("unsupported payload version: %d", version)
	}
	flags := inner[pos]
	pos++
	isDir = flags&flagIsDir != 0

	if flags&flagHasName != 0 {
		if pos+2 > len(body) {
			return nil, "", false, errors.New("truncated filename length")
		}
		nl := int(binary.BigEndian.Uint16(inner[pos : pos+2]))
		pos += 2
		if pos+nl > len(body) {
			return nil, "", false, errors.New("truncated filename")
		}
		filename = string(inner[pos : pos+nl])
		pos += nl
	}

	encrypted := flags&flagEncrypted != 0
	var salt, nonce []byte
	if encrypted {
		if pos+saltLen+nonceLen > len(body) {
			return nil, "", false, errors.New("truncated crypto header")
		}
		salt = inner[pos : pos+saltLen]
		pos += saltLen
		nonce = inner[pos : pos+nonceLen]
		pos += nonceLen
	}

	if pos+8 > len(body) {
		return nil, "", false, errors.New("truncated data length field")
	}
	dataLen := int(binary.BigEndian.Uint64(inner[pos : pos+8]))
	pos += 8
	if pos+dataLen > len(body) {
		return nil, "", false, errors.New("truncated data")
	}
	data := inner[pos : pos+dataLen]

	var compressed []byte
	if encrypted {
		if password == "" {
			return nil, "", false, errors.New("this payload is encrypted; a password is required")
		}
		key := pbkdf2([]byte(password), salt, pbkdf2Iters, pbkdf2KeyLen)
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, "", false, err
		}
		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, "", false, err
		}
		plain, err := aesgcm.Open(nil, nonce, data, nil)
		if err != nil {
			return nil, "", false, errors.New("decryption failed: wrong password or corrupted data")
		}
		compressed = plain
	} else {
		compressed = data
	}

	gr, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, "", false, fmt.Errorf("gzip open: %w", err)
	}
	defer gr.Close()
	out, err := io.ReadAll(gr)
	if err != nil {
		return nil, "", false, fmt.Errorf("gzip read: %w", err)
	}
	return out, filename, isDir, nil
}

func wrapBlob(inner []byte) []byte {
	var buf bytes.Buffer
	buf.Write(blobMagic)
	var l [8]byte
	binary.BigEndian.PutUint64(l[:], uint64(len(inner)))
	buf.Write(l[:])
	buf.Write(inner)
	return buf.Bytes()
}

func findBlob(fileBytes []byte) ([]byte, error) {
	idx := bytes.Index(fileBytes, blobMagic)
	if idx < 0 {
		return nil, errors.New("no hidden payload marker found in this file")
	}
	lenStart := idx + len(blobMagic)
	if lenStart+8 > len(fileBytes) {
		return nil, errors.New("truncated payload length field")
	}
	innerLen := int(binary.BigEndian.Uint64(fileBytes[lenStart : lenStart+8]))
	dataStart := lenStart + 8
	if dataStart+innerLen > len(fileBytes) {
		return nil, errors.New("declared payload length exceeds file size (truncated file?)")
	}
	return fileBytes[dataStart : dataStart+innerLen], nil
}

type containerFormat int

const (
	fmtGeneric containerFormat = iota
	fmtJPEG
	fmtPNG
	fmtMP4  
	fmtMP3
	fmtGIF
	fmtBMP
	fmtRIFF 
	fmtOGG  
	fmtPDF
	fmtZIP 
)

func sniffFormat(b []byte) containerFormat {
	switch {
	case len(b) >= 3 && b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF:
		return fmtJPEG
	case len(b) >= 8 && bytes.Equal(b[0:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return fmtPNG
	case len(b) >= 8 && bytes.Equal(b[4:8], []byte("ftyp")):
		return fmtMP4
	case len(b) >= 3 && bytes.Equal(b[0:3], []byte("ID3")):
		return fmtMP3
	case len(b) >= 2 && b[0] == 0xFF && (b[1]&0xE0) == 0xE0:
		return fmtMP3
	case len(b) >= 6 && (bytes.Equal(b[0:6], []byte("GIF87a")) || bytes.Equal(b[0:6], []byte("GIF89a"))):
		return fmtGIF
	case len(b) >= 2 && b[0] == 'B' && b[1] == 'M':
		return fmtBMP
	case len(b) >= 12 && bytes.Equal(b[0:4], []byte("RIFF")):
		return fmtRIFF
	case len(b) >= 4 && bytes.Equal(b[0:4], []byte("OggS")):
		return fmtOGG
	case len(b) >= 5 && bytes.Equal(b[0:5], []byte("%PDF-")):
		return fmtPDF
	case len(b) >= 4 && (bytes.Equal(b[0:4], []byte{0x50, 0x4B, 0x03, 0x04}) || bytes.Equal(b[0:4], []byte{0x50, 0x4B, 0x05, 0x06})):
		return fmtZIP
	default:
		return fmtGeneric
	}
}

func embedJPEG(cover, wrapped []byte) []byte {
	out := make([]byte, 0, len(cover)+len(wrapped))
	out = append(out, cover...)
	out = append(out, wrapped...)
	return out
}

var pngSig = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

func embedPNG(cover, wrapped []byte) ([]byte, error) {
	if len(wrapped) > 0xFFFFFFF0 {
		return nil, errors.New("payload too large for a PNG chunk")
	}
	if len(cover) < 8 || !bytes.Equal(cover[:8], pngSig) {
		return nil, errors.New("not a valid PNG file")
	}

	pos := 8
	iendStart := -1
	for pos+8 <= len(cover) {
		length := int(binary.BigEndian.Uint32(cover[pos : pos+4]))
		ctype := string(cover[pos+4 : pos+8])
		chunkTotal := 8 + length + 4
		if pos+chunkTotal > len(cover) {
			break 
		}
		if ctype == "IEND" {
			iendStart = pos
			break
		}
		pos += chunkTotal
	}
	if iendStart < 0 {
		return nil, errors.New("could not locate IEND chunk (malformed PNG?)")
	}

	var chunk bytes.Buffer
	var lenB [4]byte
	binary.BigEndian.PutUint32(lenB[:], uint32(len(wrapped)))
	chunk.Write(lenB[:])
	chunk.WriteString(pngChunkType)
	chunk.Write(wrapped)

	crcInput := append([]byte(pngChunkType), wrapped...)
	crcVal := crc32.ChecksumIEEE(crcInput)
	var crcB [4]byte
	binary.BigEndian.PutUint32(crcB[:], crcVal)
	chunk.Write(crcB[:])

	out := make([]byte, 0, len(cover)+chunk.Len())
	out = append(out, cover[:iendStart]...)
	out = append(out, chunk.Bytes()...)
	out = append(out, cover[iendStart:]...)
	return out, nil
}

func embedMP4(cover, wrapped []byte) []byte {
	var box bytes.Buffer
	total := int64(8) + int64(len(wrapped))

	if total <= 0xFFFFFFFE {
		var sz [4]byte
		binary.BigEndian.PutUint32(sz[:], uint32(total))
		box.Write(sz[:])
		box.WriteString(mp4BoxFree)
		box.Write(wrapped)
	} else {
		
		var sz [4]byte
		binary.BigEndian.PutUint32(sz[:], 1)
		box.Write(sz[:])
		box.WriteString(mp4BoxFree)
		var large [8]byte
		binary.BigEndian.PutUint64(large[:], uint64(total)+8)
		box.Write(large[:])
		box.Write(wrapped)
	}

	out := make([]byte, 0, len(cover)+box.Len())
	out = append(out, cover...)
	out = append(out, box.Bytes()...)
	return out
}

func syncsafeEncode(n uint32) [4]byte {
	var b [4]byte
	b[0] = byte((n >> 21) & 0x7F)
	b[1] = byte((n >> 14) & 0x7F)
	b[2] = byte((n >> 7) & 0x7F)
	b[3] = byte(n & 0x7F)
	return b
}

func syncsafeDecode(b []byte) uint32 {
	return uint32(b[0])<<21 | uint32(b[1])<<14 | uint32(b[2])<<7 | uint32(b[3])
}

func buildPrivFrame(wrapped []byte) []byte {
	frameData := append([]byte(id3PrivOwnerID+"\x00"), wrapped...)
	var f bytes.Buffer
	f.WriteString("PRIV")
	var sz [4]byte
	binary.BigEndian.PutUint32(sz[:], uint32(len(frameData)))
	f.Write(sz[:])
	f.Write([]byte{0x00, 0x00}) 
	f.Write(frameData)
	return f.Bytes()
}

func embedMP3(cover, wrapped []byte) ([]byte, error) {
	if len(wrapped) > (1<<28)-1024 {
		return nil, errors.New("payload too large for an ID3v2 tag")
	}
	newFrame := buildPrivFrame(wrapped)

	if len(cover) >= 10 && bytes.Equal(cover[0:3], []byte("ID3")) {
		verMajor := cover[3]
		tagSize := int(syncsafeDecode(cover[6:10]))
		framesStart := 10
		framesEnd := framesStart + tagSize
		if framesEnd > len(cover) {
			return nil, errors.New("malformed existing ID3v2 tag (declared size exceeds file)")
		}
		if verMajor == 3 || verMajor == 4 {
			
			frames := cover[framesStart:framesEnd]
			p := 0
			insertAt := len(frames)
			for p+10 <= len(frames) {
				if frames[p] == 0x00 {
					insertAt = p
					break
				}
				fsize := int(binary.BigEndian.Uint32(frames[p+4 : p+8]))
				if verMajor == 4 {
					fsize = int(syncsafeDecode(frames[p+4 : p+8]))
				}
				next := p + 10 + fsize
				if next > len(frames) || fsize < 0 {
					insertAt = p 
					break
				}
				p = next
			}
			if insertAt > len(frames) {
				insertAt = len(frames)
			}

			newFrames := make([]byte, 0, len(frames)+len(newFrame))
			newFrames = append(newFrames, frames[:insertAt]...)
			newFrames = append(newFrames, newFrame...)
			newFrames = append(newFrames, frames[insertAt:]...)

			newHeader := make([]byte, 10)
			copy(newHeader[0:3], "ID3")
			newHeader[3] = verMajor
			newHeader[4] = cover[4] 
			newHeader[5] = cover[5] 
			ss := syncsafeEncode(uint32(len(newFrames)))
			copy(newHeader[6:10], ss[:])

			out := make([]byte, 0, len(newHeader)+len(newFrames)+(len(cover)-framesEnd))
			out = append(out, newHeader...)
			out = append(out, newFrames...)
			out = append(out, cover[framesEnd:]...)
			return out, nil
		}
		
	}

	header := make([]byte, 10)
	copy(header[0:3], "ID3")
	header[3] = 3 
	header[4] = 0
	header[5] = 0
	ss := syncsafeEncode(uint32(len(newFrame)))
	copy(header[6:10], ss[:])

	out := make([]byte, 0, len(header)+len(newFrame)+len(cover))
	out = append(out, header...)
	out = append(out, newFrame...)
	out = append(out, cover...)
	return out, nil
}

func embedGeneric(cover, wrapped []byte) []byte {
	out := make([]byte, 0, len(cover)+len(wrapped))
	out = append(out, cover...)
	out = append(out, wrapped...)
	return out
}

func embedZIP(cover, wrapped []byte) ([]byte, error) {
	sig := []byte{0x50, 0x4B, 0x05, 0x06}
	windowStart := 0
	if len(cover) > 65557 { 
		windowStart = len(cover) - 65557
	}
	idx := bytes.LastIndex(cover[windowStart:], sig)
	if idx < 0 {
		return nil, errors.New("could not locate ZIP end-of-central-directory record (malformed or unusual ZIP?)")
	}
	eocdStart := windowStart + idx
	if eocdStart+22 > len(cover) {
		return nil, errors.New("malformed ZIP EOCD record")
	}

	existingCommentLen := int(binary.LittleEndian.Uint16(cover[eocdStart+20 : eocdStart+22]))
	commentStart := eocdStart + 22
	if commentStart+existingCommentLen > len(cover) {
		existingCommentLen = len(cover) - commentStart 
	}
	existingComment := cover[commentStart : commentStart+existingCommentLen]

	newComment := make([]byte, 0, len(existingComment)+len(wrapped))
	newComment = append(newComment, existingComment...)
	newComment = append(newComment, wrapped...)
	if len(newComment) > zipCommentMax {
		return nil, fmt.Errorf("payload too large for a ZIP comment (max %d bytes, need %d) - use a smaller payload or a non-ZIP-based cover file", zipCommentMax, len(newComment))
	}

	out := make([]byte, 0, commentStart+len(newComment))
	out = append(out, cover[:eocdStart+20]...)
	var cl [2]byte
	binary.LittleEndian.PutUint16(cl[:], uint16(len(newComment)))
	out = append(out, cl[:]...)
	out = append(out, newComment...)
	return out, nil
}

func tarDir(root string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		
		var link string
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}
		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		hdr.Name = filepath.ToSlash(rel)
		if info.IsDir() {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func untarToDir(data []byte, root string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return err
	}
	cleanRoot := filepath.Clean(root)

	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading archived folder: %w", err)
		}

		target := filepath.Join(cleanRoot, hdr.Name)
		
		if target != cleanRoot && !strings.HasPrefix(target, cleanRoot+string(os.PathSeparator)) {
			return fmt.Errorf("archive entry %q escapes the output folder, refusing to extract", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeSymlink:
			_ = os.Remove(target) 
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		default: 
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			mode := os.FileMode(hdr.Mode)
			if mode == 0 {
				mode = 0644
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func encode(coverPath, outPath, dataPath, text, password string) error {
	if (dataPath == "") == (text == "") {
		return errors.New("provide exactly one of -data or -text")
	}

	cover, err := os.ReadFile(coverPath)
	if err != nil {
		return fmt.Errorf("reading cover file: %w", err)
	}

	var payload []byte
	var filename string
	var isDir bool
	if dataPath != "" {
		info, statErr := os.Stat(dataPath)
		if statErr != nil {
			return fmt.Errorf("reading data path: %w", statErr)
		}
		if info.IsDir() {
			payload, err = tarDir(dataPath)
			if err != nil {
				return fmt.Errorf("archiving folder: %w", err)
			}
			filename = filepath.Base(filepath.Clean(dataPath))
			isDir = true
			fmt.Printf("packed folder %s into a %d-byte tar stream\n", dataPath, len(payload))
		} else {
			payload, err = os.ReadFile(dataPath)
			if err != nil {
				return fmt.Errorf("reading data file: %w", err)
			}
			filename = filepath.Base(dataPath)
		}
	} else {
		payload = []byte(text)
		filename = "message.txt"
	}

	inner, err := buildInner(payload, password, filename, isDir)
	if err != nil {
		return fmt.Errorf("building payload: %w", err)
	}
	wrapped := wrapBlob(inner)

	format := sniffFormat(cover)
	var out []byte
	switch format {
	case fmtJPEG:
		out = embedJPEG(cover, wrapped)
		fmt.Println("format: JPEG (appended after EOI)")
	case fmtPNG:
		out, err = embedPNG(cover, wrapped)
		if err != nil {
			return err
		}
		fmt.Println("format: PNG (custom ancillary chunk before IEND)")
	case fmtMP4:
		out = embedMP4(cover, wrapped)
		fmt.Println("format: MP4/ISO-BMFF (appended 'free' box)")
	case fmtMP3:
		out, err = embedMP3(cover, wrapped)
		if err != nil {
			return err
		}
		fmt.Println("format: MP3 (PRIV frame in ID3v2 tag)")
	case fmtGIF:
		out = embedGeneric(cover, wrapped)
		fmt.Println("format: GIF (appended after trailer)")
	case fmtBMP:
		out = embedGeneric(cover, wrapped)
		fmt.Println("format: BMP (appended after declared bitmap size)")
	case fmtRIFF:
		out = embedGeneric(cover, wrapped)
		fmt.Println("format: RIFF/WAV/AVI/WEBP (appended after declared RIFF size)")
	case fmtOGG:
		out = embedGeneric(cover, wrapped)
		fmt.Println("format: OGG (appended after last page)")
	case fmtPDF:
		out = embedGeneric(cover, wrapped)
		fmt.Println("format: PDF (appended after %%EOF; large payloads may confuse strict readers)")
	case fmtZIP:
		out, err = embedZIP(cover, wrapped)
		if err != nil {
			return err
		}
		fmt.Println("format: ZIP-based (docx/xlsx/pptx/jar/apk/...) - stored in EOCD comment field")
	default:
		out = embedGeneric(cover, wrapped)
		fmt.Println("format: unrecognized, using generic append")
	}

	if err := os.WriteFile(outPath, out, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Printf("payload: %d bytes -> stored as %d bytes (compressed%s)\n",
		len(payload), len(wrapped), map[bool]string{true: ", encrypted", false: ""}[password != ""])
	fmt.Printf("wrote %s\n", outPath)
	return nil
}

func decode(inPath, outPath, password string) error {
	fileBytes, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	inner, err := findBlob(fileBytes)
	if err != nil {
		return err
	}

	payload, filename, isDir, err := parseInner(inner, password)
	if err != nil {
		return err
	}

	target := outPath
	if target == "" {
		target = filename
		if target == "" {
			if isDir {
				target = "extracted_folder"
			} else {
				target = "extracted.bin"
			}
		}
	}

	if isDir {
		if err := untarToDir(payload, target); err != nil {
			return fmt.Errorf("extracting folder: %w", err)
		}
		fmt.Printf("extracted folder into %s/", target)
		if filename != "" {
			fmt.Printf(" (original name: %s)", filename)
		}
		fmt.Println()
		return nil
	}

	if err := os.WriteFile(target, payload, 0644); err != nil {
		return fmt.Errorf("writing extracted data: %w", err)
	}

	fmt.Printf("extracted %d bytes", len(payload))
	if filename != "" {
		fmt.Printf(" (original name: %s)", filename)
	}
	fmt.Println()
	fmt.Printf("wrote %s\n", target)
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `ByteVeil - hide a file, a folder, or a text message inside almost any cover
file (JPEG/PNG/GIF/BMP/WAV/AVI/WEBP/OGG/MP4-family/MP3/PDF/ZIP-family, or
any other file as a generic fallback), with optional AES-256-GCM encryption

Usage:
  ByteVeil encode -in <cover> -out <output> (-data <file-or-folder> | -text "message") [-password <pass>]
  ByteVeil decode -in <ByteVeil-file> -out <output> [-password <pass>]

Examples:
  ByteVeil encode -in cover.png -out out.png -data secret.pdf -password "hunter2"
  ByteVeil encode -in cover.jpg -out out.jpg -text "meet at 9pm"
  ByteVeil encode -in cover.png -out out.png -data ./my-folder -password "hunter2"
  ByteVeil decode -in out.png -out extracted.pdf -password "hunter2"
  ByteVeil decode -in out.png -out ./restored-folder -password "hunter2"

If -data points at a directory, it's packed into a tar stream (preserving
subfolders, regular files and symlinks) before hiding; decode detects this
automatically and unpacks back into a directory at -out instead of writing
a single file.

Note: this survives unmodified file transfer (e.g. Telegram "send as file"),
but NOT genuine re-encoding/recompression by platforms that transcode
media (e.g. WhatsApp/Instagram photo & video compression). No simple
embedding technique can survive real re-encoding. ZIP-based covers
(.zip/.docx/.xlsx/.pptx/.jar/.apk/...) also have a hard 65535-byte payload
ceiling, since the payload has to fit in the ZIP comment field.`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "encode":
		fs := flag.NewFlagSet("encode", flag.ExitOnError)
		in := fs.String("in", "", "cover file path")
		out := fs.String("out", "", "output ByteVeil file path")
		data := fs.String("data", "", "path to a file OR a folder to hide (folders are packed into a tar stream first)")
		text := fs.String("text", "", "a text message to hide (alternative to -data)")
		password := fs.String("password", "", "optional password (enables AES-256-GCM encryption)")
		fs.Parse(os.Args[2:])

		if *in == "" || *out == "" {
			usage()
			os.Exit(1)
		}
		if err := encode(*in, *out, *data, *text, *password); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

	case "decode":
		fs := flag.NewFlagSet("decode", flag.ExitOnError)
		in := fs.String("in", "", "ByteVeil file path")
		out := fs.String("out", "", "output path for extracted data (defaults to original filename if stored)")
		password := fs.String("password", "", "password, if the payload was encrypted")
		fs.Parse(os.Args[2:])

		if *in == "" {
			usage()
			os.Exit(1)
		}
		if err := decode(*in, *out, *password); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

	default:
		usage()
		os.Exit(1)
	}
}
