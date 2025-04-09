package headers

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	crlf = "\r\n"
)

// headers / field line format
// field-line = field-name: OWS field-value OWS \r\n
// ows = optional white space

type Headers map[string]string

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 { // not enough header info
		return 0, false, nil
	} else if idx == 0 { // crlf is at start, means we finished reading headers
		// 2 is to consume the crlf
		return 2, true, nil
	}
	header_string := string(data[:idx]) // bytes.SplitN is used instead
	key, val, found := strings.Cut(header_string, ":")
	if !found {
		return 0, false, errors.New("Could not find : in header.")
	}
	// can't be any space between key and :
	// if there isn't any space between key and char, the length should be the same
	if len(key) != len(strings.TrimRight(key, " ")) {
		return 0, false, errors.New("White space found in key.")
	}

	key = cleanKey(key)
	val = strings.TrimSpace(val)

	if !validateKey(key) {
		return 0, false, errors.New("Invalid key.")
	}

	h.Set(key, val)

	return (idx + 2), false, nil
}

func (h Headers) Set(key, value string) {
	if val, ok := h[key]; ok {
		if strings.Contains(val, value) { // if value is already in my map
			return
		}
		h[key] = fmt.Sprintf("%s, %s", val, value)
	} else {
		h[key] = value
	}
}

func validateKey(key string) bool {
	if len(key) < 1 {
		return false
	}
	validChars := "[\\w!#$%&'*+-.^`|~]"
	pattern := regexp.MustCompile(validChars)
	valid := true

	for _, r := range key {
		valid = valid && pattern.MatchString(string(r))
	}
	return valid
}

func cleanKey(key string) string {
	return strings.TrimSpace(strings.ToLower(key))
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

// validTokens checks if the data contains only valid tokens
// or characters that are allowed in a token
func validTokens(data []byte) bool {
	for _, c := range data {
		if !(c >= 'A' && c <= 'Z' ||
			c >= 'a' && c <= 'z' ||
			c >= '0' && c <= '9' ||
			c == '-') {
			return false
		}
	}
	return true
}
