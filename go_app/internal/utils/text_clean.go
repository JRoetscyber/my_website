package utils

import "strings"

var safeQuoteEntityReplacer = strings.NewReplacer(
	"&#39;", "'",
	"&#039;", "'",
	"&#x27;", "'",
	"&#X27;", "'",
	"&apos;", "'",
	"&#34;", `"`,
	"&#034;", `"`,
	"&#x22;", `"`,
	"&#X22;", `"`,
	"&quot;", `"`,
	"&#8217;", "’",
	"&#8216;", "‘",
	"&#8220;", "“",
	"&#8221;", "”",
	"&#8212;", "—",
	"&#8211;", "–",
	"&rsquo;", "’",
	"&lsquo;", "‘",
	"&rdquo;", "”",
	"&ldquo;", "“",
	"&mdash;", "—",
	"&ndash;", "–",
)

// CleanEscapedText repairs single- and multi-escaped apostrophes, quotes, and typographic
// entities (such as "I&#39;m" or "I&amp;#39;m" introduced by double template escaping)
// while leaving angle-bracket entities (&lt; / &gt;) untouched for XSS safety.
func CleanEscapedText(s string) string {
	if s == "" || !strings.ContainsRune(s, '&') {
		return s
	}

	out := s
	// Unwrap up to 4 layers of repeated &amp; double-escaping (e.g. &amp;amp;#39; -> &#39;)
	for i := 0; i < 4; i++ {
		next := strings.ReplaceAll(out, "&amp;#", "&#")
		next = strings.ReplaceAll(next, "&amp;quot;", "&quot;")
		next = strings.ReplaceAll(next, "&amp;apos;", "&apos;")
		next = strings.ReplaceAll(next, "&amp;rsquo;", "&rsquo;")
		next = strings.ReplaceAll(next, "&amp;lsquo;", "&lsquo;")
		next = strings.ReplaceAll(next, "&amp;rdquo;", "&rdquo;")
		next = strings.ReplaceAll(next, "&amp;ldquo;", "&ldquo;")
		next = strings.ReplaceAll(next, "&amp;mdash;", "&mdash;")
		next = strings.ReplaceAll(next, "&amp;ndash;", "&ndash;")
		next = strings.ReplaceAll(next, "&amp;amp;", "&amp;")
		next = safeQuoteEntityReplacer.Replace(next)
		if next == out {
			break
		}
		out = next
	}
	return out
}
