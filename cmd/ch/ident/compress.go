package ident

import "unicode/utf8"

func SafeStringForCompressString12(id string) string {
	buf := make([]rune, 0, len(id))
	count := 0
	for _, c := range id {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '.', c == '-', c == '~', c == '!':
			buf = append(buf, c)
			count++
		case c >= 0x2000 && c <= 0x9fff:
			if count > 10 {
				return string(buf)
			}
			buf = append(buf, c)
			count += 2
		default:
			buf = append(buf, '_')
			count++
		}
		if count == 12 {
			break
		}
	}
	return string(buf)
}

func CompressString12(id string) []byte {
	buf := make([]byte, 0, len(id))
	for _, c := range id {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c == '.', c == '-', c == '~', c == '!':
			buf = append(buf, byte(c))
		case c >= 0x2000 && c <= 0x9fff:
			if len(buf) > 10 {
				return buf
			}
			c = (c - 0x2000) | 0x8000
			buf = append(buf, byte(c>>8), byte(c))
		default:
			buf = append(buf, '_')
		}
		if len(buf) == 12 {
			break
		}
	}
	return buf
}

func DecompressString12(buf []byte) string {
	p := make([]byte, 0, len(buf))
	for i := 0; i < len(buf); {
		c := buf[i]
		if c < 0x80 {
			p = append(p, c)
			i++
			continue
		}
		if i == len(buf)-1 {
			// invalid
			return ""
		}
		p = append(p, 0, 0, 0)
		utf8.EncodeRune(p[len(p)-3:], (rune(c)<<8+rune(buf[i+1]))&0x7fff+0x2000)
		i += 2
	}
	return string(p)
}
