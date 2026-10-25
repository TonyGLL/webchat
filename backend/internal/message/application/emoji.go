package application

import "unicode/utf8"

const (
	zwj           = 0x200D
	variationSel  = 0xFE0F
	keycapMark    = 0x20E3
	tagTerminator = 0xE007F
)

// pictographicRanges approximates Unicode's Extended_Pictographic property (the basis of emoji
// grapheme clusters), so no extra dependency is needed.
var pictographicRanges = [][2]rune{
	{0x00A9, 0x00A9}, {0x00AE, 0x00AE}, {0x203C, 0x203C}, {0x2049, 0x2049}, {0x2122, 0x2122},
	{0x2139, 0x2139}, {0x2194, 0x2199}, {0x21A9, 0x21AA}, {0x231A, 0x231B}, {0x2328, 0x2328},
	{0x2388, 0x2388}, {0x23CF, 0x23CF}, {0x23E9, 0x23F3}, {0x23F8, 0x23FA}, {0x24C2, 0x24C2},
	{0x25AA, 0x25AB}, {0x25B6, 0x25B6}, {0x25C0, 0x25C0}, {0x25FB, 0x25FE}, {0x2600, 0x27BF},
	{0x2934, 0x2935}, {0x2B05, 0x2B07}, {0x2B1B, 0x2B1C}, {0x2B50, 0x2B50}, {0x2B55, 0x2B55},
	{0x3030, 0x3030}, {0x303D, 0x303D}, {0x3297, 0x3297}, {0x3299, 0x3299},
	{0x1F000, 0x1F0FF}, {0x1F10D, 0x1F10F}, {0x1F12F, 0x1F12F}, {0x1F16C, 0x1F171},
	{0x1F17E, 0x1F17F}, {0x1F18E, 0x1F18E}, {0x1F191, 0x1F19A}, {0x1F1AD, 0x1F1E5},
	{0x1F201, 0x1F20F}, {0x1F21A, 0x1F21A}, {0x1F22F, 0x1F22F}, {0x1F232, 0x1F23A},
	{0x1F23C, 0x1F23F}, {0x1F249, 0x1F3FA}, {0x1F400, 0x1F53D}, {0x1F546, 0x1F64F},
	{0x1F680, 0x1F6FF}, {0x1F774, 0x1F77F}, {0x1F7D5, 0x1F7FF}, {0x1F80C, 0x1F80F},
	{0x1F848, 0x1F84F}, {0x1F85A, 0x1F85F}, {0x1F888, 0x1F88F}, {0x1F8AE, 0x1F8FF},
	{0x1F90C, 0x1F93A}, {0x1F93C, 0x1F945}, {0x1F947, 0x1FAFF}, {0x1FC00, 0x1FFFD},
}

func isPictographic(r rune) bool {
	for _, rg := range pictographicRanges {
		if r >= rg[0] && r <= rg[1] {
			return true
		}
	}
	return false
}

func isSkinModifier(r rune) bool      { return r >= 0x1F3FB && r <= 0x1F3FF }
func isRegionalIndicator(r rune) bool { return r >= 0x1F1E6 && r <= 0x1F1FF }
func isTag(r rune) bool               { return r >= 0xE0020 && r <= 0xE007E }

// isSingleEmoji reports whether s is exactly one emoji grapheme cluster: a pictograph with
// optional variation selector and skin tone, joined by ZWJ into sequences; or a keycap, a
// flag (two regional indicators) or a tag sequence (subdivision flags). Plain text, several
// emoji, and combining-mark ("zalgo") sequences are rejected.
func isSingleEmoji(s string) bool {
	if s == "" || !utf8.ValidString(s) {
		return false
	}
	rs := []rune(s)
	n := len(rs)

	// keycap: [0-9#*] FE0F? 20E3
	if c := rs[0]; (c >= '0' && c <= '9') || c == '#' || c == '*' {
		i := 1
		if i < n && rs[i] == variationSel {
			i++
		}
		return i == n-1 && rs[i] == keycapMark
	}
	// flag: two regional indicators
	if isRegionalIndicator(rs[0]) {
		return n == 2 && isRegionalIndicator(rs[1])
	}
	// element (ZWJ element)*
	i := 0
	for {
		if i >= n || !isPictographic(rs[i]) {
			return false
		}
		base := rs[i]
		i++
		if i < n && rs[i] == variationSel {
			i++
		}
		if i < n && isSkinModifier(rs[i]) {
			i++
		} else if base == 0x1F3F4 && i < n && isTag(rs[i]) { // subdivision flag: base, tags, terminator
			for i < n && isTag(rs[i]) {
				i++
			}
			if i >= n || rs[i] != tagTerminator {
				return false
			}
			i++
		}
		if i == n {
			return true
		}
		if rs[i] != zwj {
			return false
		}
		i++
	}
}
