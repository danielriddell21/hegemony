package cli

import (
	"bytes"
	"strings"
	"testing"
)

func runRoot(t *testing.T, version string, args ...string) (string, error) {
	t.Helper()
	root := newRootCmd(version)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestHeadlessCommandPrintsTable(t *testing.T) {
	out, err := runRoot(t, "test", "headless", "--seeds", "3", "--width", "12", "--height", "12", "--ticks", "150")
	if err != nil {
		t.Fatalf("headless: %v", err)
	}
	if !strings.Contains(out, "WinRate") {
		t.Errorf("output missing table header:\n%s", out)
	}
}

func TestVersionFlag(t *testing.T) {
	out, err := runRoot(t, "1.2.3", "--version")
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.Contains(out, "1.2.3") {
		t.Errorf("version output = %q, want it to contain 1.2.3", out)
	}
}

func TestRunWithoutGUIReports(t *testing.T) {
	_, err := runRoot(t, "test", "run", "--width", "8", "--height", "8")
	if err == nil {
		t.Error("expected an error from the non-GUI build stub")
	}
}
