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

const cldr = `<?xml version="1.0" encoding="UTF-8" ?>
<ldmlBCP47>
	<keyword>
		<key name="tz">
			<type name="deber" alias="Europe/Berlin"/>
			<type name="sesto" alias="Europe/Stockholm"/>
			<type name="inccu" alias="Asia/Calcutta Asia/Kolkata" iana="Asia/Kolkata"/>
			<type name="utc" alias="Etc/UTC UTC Zulu"/>
		</key>
	</keyword>
</ldmlBCP47>
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
	cl := filepath.Join(dir, "timezone.xml")
	win := filepath.Join(dir, "windowsZones.xml")
	out := filepath.Join(dir, "zones_gen.go")
	os.WriteFile(tz, []byte(tzdata), 0o644)
	os.WriteFile(cl, []byte(cldr), 0o644)
	os.WriteFile(win, []byte(windows), 0o644)
	if err := generate(tz, cl, win, out); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(out)
	src := regexp.MustCompile(`:\s+"`).ReplaceAllString(string(raw), `: "`) // gofmt aligns the entries
	for _, want := range []string{
		`tzdataVersion = "2026d"`,
		`"europe/oslo": "Europe/Oslo"`,
		`"arctic/longyearbyen": "Arctic/Longyearbyen"`,
		`"Asia/Kolkata": "Asia/Calcutta"`, // CLDR's first name, whatever tzdata links to
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
	for _, absent := range []string{`"Europe/Oslo": "Europe/Berlin"`, `"Arctic/Longyearbyen":`, `"Zulu":`} {
		if strings.Contains(src, absent) {
			t.Errorf("generated table has %s: tzdata's links are not aliases, and a name outside tzdata is not listed", absent)
		}
	}
	if _, err := readCLDR(filepath.Join(dir, "none"), nil); err == nil {
		t.Error("a missing CLDR file")
	}
	os.WriteFile(cl, []byte("<ldmlBCP47/>"), 0o644)
	if _, err := readCLDR(cl, nil); err == nil {
		t.Error("a CLDR file with no groups")
	}
	os.WriteFile(cl, []byte("<notxml"), 0o644)
	if _, err := readCLDR(cl, nil); err == nil {
		t.Error("malformed CLDR xml")
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
	if err := generate(tz, cl, win, out); err == nil {
		t.Error("generate with bad inputs")
	}
}
