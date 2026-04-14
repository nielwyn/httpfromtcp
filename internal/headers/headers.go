package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	str := string(data)

	crlfIndex := strings.Index(str, crlf)
	if crlfIndex == -1 {
		return 0, false, nil
	}

	// The end of the headers
	if strings.HasPrefix(str, crlf) {
		return len(crlf), true, nil
	}

	numBytesParsed := crlfIndex + len(crlf)

	parts := strings.SplitN(str[:crlfIndex], ":", 2)
	if len(parts) < 2 {
		return 0, false, fmt.Errorf("invalid header: missing colon")
	}

	key, value := parts[0], strings.TrimSpace(parts[1])
	if !isValidFieldName(key) {
		return 0, false, fmt.Errorf("invalid header: field-name %q contains invalid characters", key)
	}
	key = strings.ToLower(key)

	if existing, ok := h[key]; ok {
		h[key] = existing + ", " + value
	} else {
		h[key] = value
	}

	return numBytesParsed, false, nil
}

func isValidFieldName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", c)) {
			return false
		}
	}
	return true
}
