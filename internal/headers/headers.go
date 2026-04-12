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

	if strings.Contains(parts[0], " ") {
		return 0, false, fmt.Errorf("invalid header: field-name %q contains invalid characters", parts[0])
	}

	h[parts[0]] = strings.TrimSpace(parts[1])

	return numBytesParsed, false, nil
}
