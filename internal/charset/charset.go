package charset

import (
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// getEncoding returns the encoding for a charset name
func getEncoding(charset string) encoding.Encoding {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	// ASCII (no transformation needed, but we can use ISO-8859-1 which is ASCII-compatible)
	case "ascii", "us-ascii":
		return charmap.ISO8859_1

	// UTF-8 (Go's native encoding, no transformation needed)
	case "utf-8", "utf8":
		return nil // nil means UTF-8 (native)

	// UTF-16
	case "utf-16", "utf16":
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	case "utf-16le", "utf16le":
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case "utf-16be", "utf16be":
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)

	// Korean
	case "euc-kr", "euckr", "cp949", "uhc", "ms949", "windows-949":
		return korean.EUCKR

	// Japanese
	case "shift-jis", "shiftjis", "shift_jis", "sjis", "cp932":
		return japanese.ShiftJIS
	case "euc-jp", "eucjp":
		return japanese.EUCJP
	case "iso-2022-jp":
		return japanese.ISO2022JP

	// Chinese (Simplified)
	case "gb2312", "gbk", "gb18030", "cp936":
		return simplifiedchinese.GBK
	case "hz-gb-2312", "hz":
		return simplifiedchinese.HZGB2312

	// Chinese (Traditional)
	case "big5", "big-5", "cp950":
		return traditionalchinese.Big5

	// ISO-8859 series (Latin)
	case "iso-8859-1", "iso88591", "latin1", "cp1252", "windows-1252":
		return charmap.ISO8859_1
	case "iso-8859-2", "iso88592", "latin2":
		return charmap.ISO8859_2
	case "iso-8859-3", "iso88593", "latin3":
		return charmap.ISO8859_3
	case "iso-8859-4", "iso88594", "latin4":
		return charmap.ISO8859_4
	case "iso-8859-5", "iso88595":
		return charmap.ISO8859_5
	case "iso-8859-6", "iso88596":
		return charmap.ISO8859_6
	case "iso-8859-7", "iso88597":
		return charmap.ISO8859_7
	case "iso-8859-8", "iso88598":
		return charmap.ISO8859_8
	case "iso-8859-9", "iso88599", "latin5":
		return charmap.ISO8859_9
	case "iso-8859-10", "iso885910", "latin6":
		return charmap.ISO8859_10
	case "iso-8859-13", "iso885913", "latin7":
		return charmap.ISO8859_13
	case "iso-8859-14", "iso885914", "latin8":
		return charmap.ISO8859_14
	case "iso-8859-15", "iso885915", "latin9":
		return charmap.ISO8859_15
	case "iso-8859-16", "iso885916", "latin10":
		return charmap.ISO8859_16

	// Cyrillic
	case "koi8-r", "koi8r":
		return charmap.KOI8R
	case "koi8-u", "koi8u":
		return charmap.KOI8U

	// Windows code pages
	case "windows-1250", "cp1250":
		return charmap.Windows1250
	case "windows-1251", "cp1251":
		return charmap.Windows1251
	case "windows-1253", "cp1253":
		return charmap.Windows1253
	case "windows-1254", "cp1254":
		return charmap.Windows1254
	case "windows-1255", "cp1255":
		return charmap.Windows1255
	case "windows-1256", "cp1256":
		return charmap.Windows1256
	case "windows-1257", "cp1257":
		return charmap.Windows1257
	case "windows-1258", "cp1258":
		return charmap.Windows1258

	default:
		// Unknown charset, return nil (treat as UTF-8)
		return nil
	}
}

// Encode converts a UTF-8 string to the specified charset bytes
func Encode(text string, charset string) ([]byte, error) {
	enc := getEncoding(charset)
	if enc == nil {
		// UTF-8 or unknown, return as-is
		return []byte(text), nil
	}

	encoder := enc.NewEncoder()
	encoded, err := encoder.Bytes([]byte(text))
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

// Decode converts bytes in the specified charset to a UTF-8 string
func Decode(data []byte, charset string) (string, error) {
	enc := getEncoding(charset)
	if enc == nil {
		// UTF-8 or unknown, return as-is
		return string(data), nil
	}

	decoder := enc.NewDecoder()
	decoded, err := decoder.Bytes(data)
	if err != nil {
		// If decoding fails, return the raw bytes as string (best effort)
		return string(data), nil
	}
	return string(decoded), nil
}

// IsSupported checks if a charset is supported
func IsSupported(charset string) bool {
	// UTF-8 is always supported
	lower := strings.ToLower(strings.TrimSpace(charset))
	if lower == "utf-8" || lower == "utf8" {
		return true
	}
	return getEncoding(charset) != nil
}
