package grapheme

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

// breakTest is one line of GraphemeBreakTest.txt: the text and the clusters it splits into.
type breakTest struct {
	line     int
	text     string
	clusters []string
	comment  string
}

// readBreakTests parses every line of Unicode's GraphemeBreakTest.txt, which writes a case as
// code points in hexadecimal separated by ÷ (a break) or × (no break).
func readBreakTests(t testing.TB) []breakTest {
	f, err := os.Open("testdata/GraphemeBreakTest.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var tests []breakTest
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		data, comment, _ := strings.Cut(sc.Text(), "#")
		if strings.TrimSpace(data) == "" {
			continue
		}
		tc := breakTest{line: n, comment: strings.TrimSpace(comment)}
		var cluster strings.Builder
		for _, field := range strings.Fields(data) {
			switch field {
			case "÷":
				if cluster.Len() > 0 {
					tc.clusters = append(tc.clusters, cluster.String())
					cluster.Reset()
				}
			case "×":
			default:
				r, err := strconv.ParseUint(field, 16, 32)
				if err != nil {
					t.Fatalf("line %d: bad code point %q", n, field)
				}
				cluster.WriteRune(rune(r))
				tc.text += string(rune(r))
			}
		}
		if cluster.Len() > 0 {
			t.Fatalf("line %d: does not end in ÷", n)
		}
		tests = append(tests, tc)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return tests
}

// TestConformance runs every case of the official test file, and checks that the file makes
// every rule decide at least one boundary, so no rule passes for lack of a case.
func TestConformance(t *testing.T) {
	tests := readBreakTests(t)
	if len(tests) < 700 {
		t.Fatalf("read %d cases; GraphemeBreakTest.txt for %s has 766", len(tests), Version)
	}
	fired := map[string]int{}
	traceRule = func(rule string) { fired[rule]++ }
	defer func() { traceRule = nil }()
	for _, tc := range tests {
		if got := Split(tc.text); !slices.Equal(got, tc.clusters) {
			t.Errorf("line %d: Split(%+q) = %+q, want %+q\n\t%s", tc.line, tc.text, got, tc.clusters, tc.comment)
		}
	}
	for _, rule := range []string{"GB3", "GB4", "GB5", "GB6", "GB7", "GB8", "GB9", "GB9a", "GB9b", "GB9c", "GB11", "GB12", "GB999"} {
		if fired[rule] == 0 {
			t.Errorf("no case of the test file reaches rule %s", rule)
		}
	}
}

// TestVersions holds the grapheme tables, Go's unicode package, and golang.org/x/text to one
// Unicode version, as String.unicodeVersion promises (section 11, String). It fails when a Go or
// x/text upgrade moves one of them; then regenerate with go generate at the new version.
func TestVersions(t *testing.T) {
	versions := map[string]string{
		"internal/grapheme":   Version,
		"unicode":             unicode.Version,
		"x/text/unicode/norm": norm.Version,
		"x/text/cases":        cases.UnicodeVersion,
	}
	for name, v := range versions {
		if v != Version {
			t.Errorf("%s follows Unicode %s, the grapheme tables %s", name, v, Version)
		}
	}
}

// TestProperties checks one code point of each property the generator reads, so a file it
// missed, Indic_Conjunct_Break above all, shows here by name.
func TestProperties(t *testing.T) {
	for _, tc := range []struct {
		r        rune
		brk      uint8
		pict     bool
		conjunct uint8
	}{
		{'a', other, false, 0},
		{'\r', cr, false, 0},
		{'\n', lf, false, 0},
		{0x0000, control, false, 0},
		{0x0301, extend, false, incbExtend},   // COMBINING ACUTE ACCENT
		{0x200D, zwj, false, incbExtend},      // ZERO WIDTH JOINER
		{0x1F1E6, regional, false, 0},         // REGIONAL INDICATOR SYMBOL LETTER A
		{0x0600, prepend, false, 0},           // ARABIC NUMBER SIGN
		{0x0903, spacingMark, false, 0},       // DEVANAGARI SIGN VISARGA
		{0x1100, hangulL, false, 0},           // HANGUL CHOSEONG KIYEOK
		{0x1161, hangulV, false, 0},           // HANGUL JUNGSEONG A
		{0x11A8, hangulT, false, 0},           // HANGUL JONGSEONG KIYEOK
		{0xAC00, hangulLV, false, 0},          // HANGUL SYLLABLE GA
		{0xAC01, hangulLVT, false, 0},         // HANGUL SYLLABLE GAG
		{0x00A9, other, true, 0},              // COPYRIGHT SIGN
		{0x1F600, other, true, 0},             // GRINNING FACE
		{0x0915, other, false, incbConsonant}, // DEVANAGARI LETTER KA
		{0x094D, extend, false, incbLinker},   // DEVANAGARI SIGN VIRAMA
		{utf8.RuneError, other, false, 0},
	} {
		p := lookup(tc.r)
		if p.breakValue() != tc.brk || p.pictographic() != tc.pict || p.conjunct() != tc.conjunct {
			t.Errorf("U+%04X: break %d, pictographic %v, conjunct %d; want %d, %v, %d",
				tc.r, p.breakValue(), p.pictographic(), p.conjunct(), tc.brk, tc.pict, tc.conjunct)
		}
	}
}

// TestExamples names the cases the specification and its readers care about.
func TestExamples(t *testing.T) {
	for _, tc := range []struct {
		text string
		want []string
	}{
		{"", nil},
		{"ant", []string{"a", "n", "t"}},
		{"e\u0301", []string{"e\u0301"}},
		{"👨\u200D👩\u200D👧e\u0301", []string{"👨\u200D👩\u200D👧", "e\u0301"}},
		{"🇨🇭🇩🇪🇫", []string{"🇨🇭", "🇩🇪", "🇫"}},
		{"a\r\nb", []string{"a", "\r\n", "b"}},
		{"क्षि", []string{"क्षि"}},
		{"नमस्ते", []string{"न", "म", "स्ते"}},
		{"👋🏽", []string{"👋🏽"}},
		{"🏴\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F", []string{"🏴\U000E0067\U000E0062\U000E0073\U000E0063\U000E0074\U000E007F"}},
		{"각", []string{"각"}},
		{"\xff\xfe", []string{"\xff", "\xfe"}},
	} {
		if got := Split(tc.text); !slices.Equal(got, tc.want) {
			t.Errorf("Split(%+q) = %+q, want %+q", tc.text, got, tc.want)
		}
	}
}

// TestAgainstICU compares Split with Intl.Segmenter of Node, which follows ICU, over random text
// drawn from every property class. It runs when node is on the path and follows the same
// Unicode version, and skips otherwise.
func TestAgainstICU(t *testing.T) {
	if testing.Short() {
		t.Skip("short")
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not on the path")
	}
	out, err := exec.Command(node, "-p", "process.versions.unicode").Output()
	if err != nil || !strings.HasPrefix(Version, strings.TrimSpace(string(out))) {
		t.Skipf("node follows Unicode %q, the tables %s", strings.TrimSpace(string(out)), Version)
	}

	// Group code points by their properties, and draw a class first, then a member, so that
	// rare classes such as Prepend and the conjunct linkers appear as often as letters.
	classes := map[uint8][]rune{0: {'a', 'Z', '0', ' ', 'é', '中', 0xFFFD}}
	for _, s := range ranges {
		classes[s.props] = append(classes[s.props], s.lo, s.hi, s.lo+(s.hi-s.lo)/2)
	}
	keys := make([]uint8, 0, len(classes))
	for k := range classes {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	rng := rand.New(rand.NewSource(1))
	texts := make([]string, 20000)
	for i := range texts {
		var b strings.Builder
		for n := 1 + rng.Intn(10); n > 0; n-- {
			members := classes[keys[rng.Intn(len(keys))]]
			r := members[rng.Intn(len(members))]
			if r >= 0xD800 && r <= 0xDFFF {
				r = 'x'
			}
			b.WriteRune(r)
		}
		texts[i] = b.String()
	}

	input, _ := json.Marshal(texts)
	cmd := exec.Command(node, "-e", `
		const texts = JSON.parse(require("fs").readFileSync(0, "utf8"));
		const seg = new Intl.Segmenter("und", { granularity: "grapheme" });
		process.stdout.write(JSON.stringify(texts.map(t => [...seg.segment(t)].map(s => s.segment))));
	`)
	cmd.Stdin = strings.NewReader(string(input))
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("node: %v", err)
	}
	var want [][]string
	if err := json.Unmarshal(output, &want); err != nil {
		t.Fatal(err)
	}
	failures := 0
	for i, text := range texts {
		if got := Split(text); !slices.Equal(got, want[i]) && !(len(got) == 0 && len(want[i]) == 0) {
			if failures++; failures <= 20 {
				t.Errorf("Split(%+q) = %+q, ICU %+q", text, got, want[i])
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d of %d texts differ from ICU", failures, len(texts))
	}
}

// FuzzSplit checks what holds for any input, text or not: the clusters are not empty, they
// concatenate to the input, and each splits again into itself alone.
func FuzzSplit(f *testing.F) {
	for _, tc := range readBreakTests(f) {
		f.Add(tc.text)
	}
	f.Add("\xff\xfe\r\n")
	f.Fuzz(func(t *testing.T, s string) {
		clusters := Split(s)
		if strings.Join(clusters, "") != s {
			t.Fatalf("Split(%+q) = %+q, which does not concatenate to the input", s, clusters)
		}
		for _, c := range clusters {
			if c == "" {
				t.Fatalf("Split(%+q) has an empty cluster", s)
			}
			if again := Split(c); len(again) != 1 {
				t.Fatalf("cluster %+q of %+q splits again into %+q", c, s, again)
			}
		}
	})
}
