package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsRemote(t *testing.T) {
	remote := []string{
		"user@host:/srv/app",
		"host:/srv/app",
		"host:~/app",
		"user@host:~/app",
	}
	local := []string{
		"/srv/app",
		"/home/user/app",
		".",
		"C:\\Users\\x\\app",
		"relative/path",
		"",
		"host",
	}
	for _, p := range remote {
		if !isRemote(p) {
			t.Errorf("expected %q to be remote", p)
		}
	}
	for _, p := range local {
		if isRemote(p) {
			t.Errorf("expected %q to be local", p)
		}
	}
}

func TestProjectTarget(t *testing.T) {
	if got := newProject("/srv/app").target("icons/sprite.svg"); got != "/srv/app/icons/sprite.svg" {
		t.Errorf("unexpected target: %s", got)
	}
	p := newProject("user@host:/srv/app")
	if !p.remote() {
		t.Fatal("expected remote project")
	}
	if got := p.target("icons/sprite.svg"); got != "/srv/app/icons/sprite.svg" {
		t.Errorf("unexpected remote target: %s", got)
	}
}

func TestProjectLocalReadWriteBackup(t *testing.T) {
	base := t.TempDir()
	p := newProject(base)

	if err := p.writeFile("sprite.svg", []byte("<svg></svg>")); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	data, err := p.readFile("sprite.svg")
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(data) != "<svg></svg>" {
		t.Errorf("unexpected content: %s", data)
	}

	if err := p.backup("sprite.svg"); err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(base, "sprite.svg.orig")); err != nil {
		t.Errorf("backup file not created: %v", err)
	}
}

func TestProjectLocalReadMissing(t *testing.T) {
	p := newProject(t.TempDir())
	if _, err := p.readFile("missing.svg"); err == nil {
		t.Fatal("expected error reading missing file, got nil")
	} else if !strings.Contains(err.Error(), "missing.svg") {
		t.Fatalf("error should mention the file: %v", err)
	}
}
