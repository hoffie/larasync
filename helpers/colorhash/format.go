package colorhash

import (
	"encoding/hex"
	"fmt"
)

const (
	charsPerLine  = 16
	charsPerBlock = 4
)

// clearColor returns control codes for resetting all formatting
func clearColor() string {
	return "\x1b[0m"
}

// colorBlock chooses a foreground and background color for the given 4-byte hex
// block. It does this by using the byte value as a color index.
func colorBlock(b string) string {
	asHex, err := hex.DecodeString(b)
	if err != nil || len(asHex) != 2 {
		// should not happen; if it does, leave the string uncolored
		return ""
	}
	return fmt.Sprintf("\x1b[48;5;%d;38;5;%dm", int(asHex[0]), int(asHex[1]))
}

// Format returns a colored representation of the given hash.
// Only 128-byte hex hashes are supported right now.
func Format(s string) string {
	res := ""
	if len(s) != 128 {
		return s
	}
	for x := 0; x < len(s); x += charsPerLine {
		line := s[x : x+charsPerLine]
		res += "  "
		for y := 0; y < len(line); y += charsPerBlock {
			block := line[y : y+charsPerBlock]
			res += colorBlock(block)
			res += block
		}
		res += clearColor()
		res += "\n"
	}
	return res
}
