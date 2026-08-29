package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLogOptions(t *testing.T) {
	opts, err := parseLogOptions(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.tail != 50 || opts.follow {
		t.Fatalf("unexpected defaults: %+v", opts)
	}

	opts, err = parseLogOptions([]string{"-f"})
	if err != nil || !opts.follow {
		t.Fatalf("expected follow=true, got %+v err=%v", opts, err)
	}

	opts, err = parseLogOptions([]string{"-n", "10"})
	if err != nil || opts.tail != 10 {
		t.Fatalf("expected tail=10, got %+v err=%v", opts, err)
	}

	opts, err = parseLogOptions([]string{"--tail=25"})
	if err != nil || opts.tail != 25 {
		t.Fatalf("expected tail=25, got %+v err=%v", opts, err)
	}

	if _, err := parseLogOptions([]string{"--bogus"}); err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if _, err := parseLogOptions([]string{"-n", "abc"}); err == nil {
		t.Fatal("expected error for non-numeric tail")
	}
}

func TestSetEnvKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("PORT=3300\n# comment\nADMIN_USERNAME=admin\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := setEnvKey(path, "PORT", "3301"); err != nil {
		t.Fatal(err)
	}
	if err := setEnvKey(path, "DB_PATH", "data/app.db"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, w := range []string{"PORT=3301", "# comment", "ADMIN_USERNAME=admin", "DB_PATH=data/app.db"} {
		if !strings.Contains(got, w) {
			t.Errorf("missing %q in output:\n%s", w, got)
		}
	}
}

func TestEnsureEnvFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := ensureEnvFile(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "PORT=3300") {
		t.Fatalf("unexpected default env content:\n%s", data)
	}
}

func TestTailLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	var sb strings.Builder
	for i := 1; i <= 100; i++ {
		fmt.Fprintf(&sb, "line %d\n", i)
	}
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	lines, err := tailLines(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "line 91" || lines[9] != "line 100" {
		t.Fatalf("unexpected lines: %v", lines)
	}
}

func TestParseETime(t *testing.T) {
	cases := map[string]int64{
		"42":          42,
		"05:10":       310,
		"01:02:03":    3723,
		"2-01:02:03":  176523,
		"  1:02:03  ": 3723,
	}
	for in, want := range cases {
		d, ok := parseETime(in)
		if !ok {
			t.Errorf("parseETime(%q) = not ok, want %d", in, want)
			continue
		}
		if int64(d.Seconds()) != want {
			t.Errorf("parseETime(%q) = %d, want %d", in, int64(d.Seconds()), want)
		}
	}
	if _, ok := parseETime("abc"); ok {
		t.Error("parseETime(abc) should not be ok")
	}
}
