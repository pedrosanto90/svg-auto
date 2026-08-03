package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type project struct {
	host     string
	basePath string
}

func newProject(path string) *project {
	if isRemote(path) {
		idx := strings.Index(path, ":")
		return &project{host: path[:idx], basePath: path[idx+1:]}
	}
	return &project{basePath: path}
}

func isRemote(path string) bool {
	idx := strings.Index(path, ":")
	if idx <= 0 {
		return false
	}
	host := path[:idx]
	rest := path[idx+1:]
	if strings.ContainsAny(host, `/\`) || len(host) == 1 {
		return false
	}
	return strings.HasPrefix(rest, "/") || strings.HasPrefix(rest, "~/")
}

func (p *project) remote() bool {
	return p.host != ""
}

func (p *project) target(rel string) string {
	return filepath.Join(p.basePath, rel)
}

func (p *project) readFile(rel string) ([]byte, error) {
	target := p.target(rel)
	if p.remote() {
		out, err := runSSH(p.host, "cat "+shq(target), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to read remote file %s: %w", target, err)
		}
		return out, nil
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", target, err)
	}
	return data, nil
}

func (p *project) writeFile(rel string, data []byte) error {
	target := p.target(rel)
	if p.remote() {
		if _, err := runSSH(p.host, "cat > "+shq(target), data); err != nil {
			return fmt.Errorf("failed to write remote file %s: %w", target, err)
		}
		return nil
	}
	if err := os.WriteFile(target, data, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", target, err)
	}
	return nil
}

func (p *project) backup(rel string) error {
	target := p.target(rel)
	backup := target + ".orig"
	if p.remote() {
		if _, err := runSSH(p.host, "cp "+shq(target)+" "+shq(backup), nil); err != nil {
			return fmt.Errorf("failed to back up remote file %s: %w", target, err)
		}
		return nil
	}
	data, err := os.ReadFile(target)
	if err != nil {
		return fmt.Errorf("failed to read %s for backup: %w", target, err)
	}
	if err := os.WriteFile(backup, data, 0o644); err != nil {
		return fmt.Errorf("failed to write backup %s: %w", backup, err)
	}
	return nil
}

func shq(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func runSSH(host, cmd string, stdin []byte) ([]byte, error) {
	c := exec.Command("ssh", host, cmd)
	if stdin != nil {
		c.Stdin = bytes.NewReader(stdin)
	}
	var out, errb bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errb
	if err := c.Run(); err != nil {
		return nil, fmt.Errorf("ssh %s failed: %v: %s", host, err, strings.TrimSpace(errb.String()))
	}
	return out.Bytes(), nil
}
