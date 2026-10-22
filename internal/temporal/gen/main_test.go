package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const tzdata = `# version 2026d
# redo posix_only
R d 1916 o - Jun 14 23s 1 S
Z Europe/Berlin 0:53:28 - LMT 1893 Ap
1 c CE%sT
Z Asia/Kolkata 5:53:28 - LMT 1854 Jun 28
L Europe/Berlin Europe/Oslo
L Europe/Oslo Arctic/Longyearbyen
L Asia/Kolkata Asia/Calcutta
`

const windows = `<?xml version="1.0" encoding="UTF-8" ?>
<supplementalData>
	<windowsZones>
		<mapTimezones>
			<mapZone other="W. Europe Standard Time" territory="001" type="Europe/Berlin"/>
			<mapZone other="W. Europe Standard Time" territory="AD" type="Europe/Andorra"/>
			<mapZone other="India Standard Time" territory="001" type="Asia/Calcutta"/>
		</mapTimezones>
	</windowsZones>
</supplementalData>
`

func TestGenerate(t *testing.T) {
	dir := t.TempDir()
	tz := filepath.Join(dir, "tzdata.zi")
	win := filepath.Join(dir, "windowsZones.xml")
	out := filepath.Join(dir, "zones_gen.go")
	os.WriteFile(tz, []byte(tzdata), 0o644)
	os.WriteFile(win, []byte(windows), 0o644)
	if err := generate(tz, win, out); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(out)
	src := regexp.MustCompile(`:\s+"`).ReplaceAllString(string(raw), `: "`) // gofmt aligns the entries
	for _, want := range []string{
		`tzdataVersion = "2026d"`,
		`"europe/oslo": "Europe/Oslo"`,
		`"arctic/longyearbyen": "Arctic/Longyearbyen"`,
		`"Arctic/Longyearbyen": "Europe/Berlin"`, // a link to a link ends at the zone
		`"Asia/Calcutta": "Asia/Kolkata"`,
		`"W. Europe Standard Time": "Europe/Berlin"`,
		`"India Standard Time": "Asia/Calcutta"`,
	} {
		if !strings.Contains(src, want) {
			t.Errorf("generated table lacks %s", want)
		}
	}
	if strings.Contains(src, "Andorra") {
		t.Error("a territory other than 001 was taken")
	}
	if _, _, _, err := readTzdata(filepath.Join(dir, "none")); err == nil {
		t.Error("a missing tzdata file")
	}
	os.WriteFile(tz, []byte("R d 1916 o - Jun 14 23s 1 S\n"), 0o644)
	if _, _, _, err := readTzdata(tz); err == nil {
		t.Error("a file with no zones")
	}
	if _, err := readWindows(filepath.Join(dir, "none")); err == nil {
		t.Error("a missing Windows file")
	}
	os.WriteFile(win, []byte("<notxml"), 0o644)
	if _, err := readWindows(win); err == nil {
		t.Error("malformed xml")
	}
	os.WriteFile(win, []byte(`<supplementalData><windowsZones><mapTimezones></mapTimezones></windowsZones></supplementalData>`), 0o644)
	if _, err := readWindows(win); err == nil {
		t.Error("no entries")
	}
	if err := generate(tz, win, out); err == nil {
		t.Error("generate with bad inputs")
	}
}
