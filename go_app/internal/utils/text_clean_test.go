package utils

import (
	"strings"
	"testing"
)

func TestCleanEscapedText_RepairsBlogApostrophesAndPreservesXSSSafety(t *testing.T) {
	corruptedSingle := `Where It Went From There
I&#39;m doing DevOps, development, third-party integrations, bug fixing, and building an entirely new system from scratch in Go Fiber. Average API response time on the new system: 50ms. The old one was not hitting those numbers.

I&#39;m not going to pretend the salary is great right now. It isn&#39;t. But the experience I&#39;m getting is the kind of thing you can&#39;t buy. I&#39;m learning every single day and I genuinely love what I&#39;m doing.

There&#39;s a verse that&#39;s been sitting close to me through all of this. Matthew 6:25-26. If you know it, you know why. If you don&#39;t, go read it. It&#39;ll make more sense than anything else I&#39;ve written here.

I&#39;m exactly where I&#39;m supposed to be.`

	corruptedDouble := strings.ReplaceAll(corruptedSingle, "&#39;", "&amp;#39;")

	for _, sample := range []string{corruptedSingle, corruptedDouble} {
		cleaned := CleanEscapedText(sample)
		if strings.Contains(cleaned, "&#39;") || strings.Contains(cleaned, "&amp;") {
			t.Fatalf("CleanEscapedText failed to remove escaped apostrophes:\n%s", cleaned)
		}
		for _, expected := range []string{"I'm doing DevOps", "It isn't.", "you can't buy", "There's a verse that's", "If you don't", "It'll make more sense", "I've written"} {
			if !strings.Contains(cleaned, expected) {
				t.Errorf("expected %q in cleaned text, got:\n%s", expected, cleaned)
			}
		}
	}

	// Ensure &lt;script&gt; in prose is NOT turned into executable <script>
	xssSample := "Example payload: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
	cleanedXSS := CleanEscapedText(xssSample)
	if strings.Contains(cleanedXSS, "<script>") {
		t.Errorf("CleanEscapedText must not unescape angle brackets: %s", cleanedXSS)
	}
	if cleanedXSS != "Example payload: &lt;script&gt;alert('xss')&lt;/script&gt;" {
		t.Errorf("unexpected output: %s", cleanedXSS)
	}
}
