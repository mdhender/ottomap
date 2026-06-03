package wog

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"regexp"
	"unicode/utf16"
	"unicode/utf8"
)

// readPipeline turns a .wxx byte stream into UTF-8 XML bytes whose XML
// declaration is rewritten to be acceptable to encoding/xml (version="1.0",
// encoding="utf-8").
func readPipeline(r io.Reader) ([]byte, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	// Gunzip if compressed.
	if isGzip(raw) {
		gz, err := gzip.NewReader(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("gunzip: %w", err)
		}
		raw, err = io.ReadAll(gz)
		_ = gz.Close()
		if err != nil {
			return nil, fmt.Errorf("gunzip read: %w", err)
		}
	}
	// Decode UTF-16 if a BOM is present. Worldographer files are UTF-16
	// (usually BE) with a BOM; the BOM is the authoritative byte-order signal.
	utf8Bytes, err := toUTF8(raw)
	if err != nil {
		return nil, fmt.Errorf("utf-16 decode: %w", err)
	}
	return rewriteXMLHeaderForUTF8(utf8Bytes), nil
}

// writePipeline takes UTF-8 XML bytes (with whatever XML declaration the
// schema emitted), rewrites the declaration to advertise utf-16, encodes to
// UTF-16 (with BOM), and optionally gzip-compresses.
func writePipeline(xml []byte, opts WriteOptions, w io.Writer) error {
	xml = rewriteXMLHeaderForUTF16(xml)
	u16 := fromUTF8(xml, opts.utf16BE())
	if opts.compress() {
		gz := gzip.NewWriter(w)
		if _, err := gz.Write(u16); err != nil {
			_ = gz.Close()
			return fmt.Errorf("gzip write: %w", err)
		}
		return gz.Close()
	}
	_, err := w.Write(u16)
	return err
}

func isGzip(b []byte) bool {
	return len(b) >= 2 && b[0] == 0x1f && b[1] == 0x8b
}

// toUTF8 detects an optional UTF-16 BOM and returns UTF-8 bytes. If no BOM
// is present and the data already looks like UTF-8, it's returned as-is.
func toUTF8(b []byte) ([]byte, error) {
	switch {
	case len(b) >= 2 && b[0] == 0xFE && b[1] == 0xFF: // UTF-16 BE BOM
		return decodeUTF16(b[2:], binary.BigEndian)
	case len(b) >= 2 && b[0] == 0xFF && b[1] == 0xFE: // UTF-16 LE BOM
		return decodeUTF16(b[2:], binary.LittleEndian)
	case len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF: // UTF-8 BOM
		return b[3:], nil
	}
	// No BOM. If valid UTF-8, assume UTF-8.
	if utf8.Valid(b) {
		return b, nil
	}
	// Best-effort assumption: Worldographer historically writes BE.
	return decodeUTF16(b, binary.BigEndian)
}

func decodeUTF16(b []byte, order binary.ByteOrder) ([]byte, error) {
	if len(b)%2 != 0 {
		return nil, fmt.Errorf("utf-16 data has odd byte length")
	}
	u16 := make([]uint16, len(b)/2)
	for i := 0; i < len(u16); i++ {
		u16[i] = order.Uint16(b[i*2 : i*2+2])
	}
	runes := utf16.Decode(u16)
	buf := make([]byte, 0, len(runes)*utf8.UTFMax)
	tmp := make([]byte, utf8.UTFMax)
	for _, r := range runes {
		n := utf8.EncodeRune(tmp, r)
		buf = append(buf, tmp[:n]...)
	}
	return buf, nil
}

// fromUTF8 produces UTF-16 bytes with a leading BOM.
func fromUTF8(b []byte, bigEndian bool) []byte {
	runes := bytes.Runes(b)
	u16 := utf16.Encode(runes)
	out := make([]byte, 2+len(u16)*2)
	if bigEndian {
		out[0], out[1] = 0xFE, 0xFF
		for i, u := range u16 {
			binary.BigEndian.PutUint16(out[2+i*2:], u)
		}
	} else {
		out[0], out[1] = 0xFF, 0xFE
		for i, u := range u16 {
			binary.LittleEndian.PutUint16(out[2+i*2:], u)
		}
	}
	return out
}

// xmlDeclRE matches the optional XML declaration at the very start of the
// document (after any BOM, which we've already stripped). Worldographer
// emits version="1.1" encoding="utf-16" with single quotes; Go's encoding/xml
// only accepts version="1.0".
var xmlDeclRE = regexp.MustCompile(`^<\?xml[^?]*\?>\s*`)

func rewriteXMLHeaderForUTF8(b []byte) []byte {
	const want = `<?xml version="1.0" encoding="utf-8"?>` + "\n"
	if loc := xmlDeclRE.FindIndex(b); loc != nil {
		return append([]byte(want), b[loc[1]:]...)
	}
	return append([]byte(want), b...)
}

func rewriteXMLHeaderForUTF16(b []byte) []byte {
	const want = `<?xml version='1.1' encoding='utf-16'?>` + "\n"
	if loc := xmlDeclRE.FindIndex(b); loc != nil {
		return append([]byte(want), b[loc[1]:]...)
	}
	return append([]byte(want), b...)
}
