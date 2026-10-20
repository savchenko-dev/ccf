package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var version = "dev"

const pinsFileName = ".ccf-pins"

const maxSlugLen = 200

const usage = `ccf - fork a pinned Claude Code session

Usage:
  ccf [name] [claude args...]   fork the pinned session (pins it first if none exist)
  ccf pin [name] [session-id]   pin a session (default: name "main", latest session here)
  ccf ls                        list pins for the current directory
  ccf rm <name>                 remove a pin
  ccf help | version

Pins are per directory and live next to Claude Code's own session files.
`

var reserved = []string{"pin", "ls", "rm", "help", "version"}

type pin struct {
	name string
	id   string
}

type store struct {
	cwd        string
	projectDir string
	pins       []pin
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ccf:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	s, err := loadStore(cwd)
	if err != nil {
		return err
	}

	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "pin":
		return s.pin(args[1:])
	case "ls":
		return s.list()
	case "rm":
		if len(args) < 2 {
			return errors.New("usage: ccf rm <name>")
		}
		return s.remove(args[1])
	case "help", "-h", "--help":
		fmt.Print(usage)
		return nil
	case "version", "--version":
		fmt.Println("ccf", version)
		return nil
	}
	return s.fork(args)
}

func claudeConfigDir() (string, error) {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

var nonAlnum = regexp.MustCompile(`[^A-Za-z0-9]`)

func projectDirFor(cwd string) (string, error) {
	config, err := claudeConfigDir()
	if err != nil {
		return "", err
	}
	projects := filepath.Join(config, "projects")
	dir := projectDirIn(projects, cwd)
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}
	if resolved, err := filepath.EvalSymlinks(cwd); err == nil && resolved != cwd {
		if alt := projectDirIn(projects, resolved); alt != dir {
			if _, err := os.Stat(alt); err == nil {
				return alt, nil
			}
		}
	}
	return dir, nil
}

func projectDirIn(projects, cwd string) string {
	slug := nonAlnum.ReplaceAllString(cwd, "-")
	dir := filepath.Join(projects, slug)
	if len(slug) <= maxSlugLen {
		return dir
	}
	if _, err := os.Stat(dir); err == nil {
		return dir
	}
	matches, _ := filepath.Glob(filepath.Join(projects, slug[:maxSlugLen]+"-*"))
	if len(matches) == 1 {
		return matches[0]
	}
	return dir
}

func loadStore(cwd string) (*store, error) {
	dir, err := projectDirFor(cwd)
	if err != nil {
		return nil, err
	}
	s := &store{cwd: cwd, projectDir: dir}
	data, err := os.ReadFile(s.pinsFile())
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			s.pins = append(s.pins, pin{name: fields[0], id: fields[1]})
		}
	}
	return s, nil
}

func (s *store) pinsFile() string {
	return filepath.Join(s.projectDir, pinsFileName)
}

func (s *store) indexOf(name string) int {
	return slices.IndexFunc(s.pins, func(p pin) bool { return p.name == name })
}

func (s *store) save() error {
	if len(s.pins) == 0 {
		err := os.Remove(s.pinsFile())
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.MkdirAll(s.projectDir, 0o755); err != nil {
		return err
	}
	var b strings.Builder
	for _, p := range s.pins {
		fmt.Fprintf(&b, "%s %s\n", p.name, p.id)
	}
	return os.WriteFile(s.pinsFile(), []byte(b.String()), 0o644)
}

func (s *store) latestSession() (string, error) {
	files, _ := filepath.Glob(filepath.Join(s.projectDir, "*.jsonl"))
	var latest string
	var latestMod int64
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		if mod := info.ModTime().UnixNano(); latest == "" || mod > latestMod {
			latest, latestMod = f, mod
		}
	}
	if latest == "" {
		return "", fmt.Errorf("no sessions in %s", s.cwd)
	}
	return strings.TrimSuffix(filepath.Base(latest), ".jsonl"), nil
}

func (s *store) pin(args []string) error {
	name := "main"
	if len(args) > 0 {
		name = args[0]
	}
	if slices.Contains(reserved, name) {
		return fmt.Errorf("'%s' is a reserved word", name)
	}
	var id string
	if len(args) > 1 {
		id = args[1]
	} else {
		var err error
		if id, err = s.latestSession(); err != nil {
			return err
		}
	}
	if i := s.indexOf(name); i >= 0 {
		s.pins[i].id = id
	} else {
		s.pins = append(s.pins, pin{name: name, id: id})
	}
	if err := s.save(); err != nil {
		return err
	}
	fmt.Printf("ccf: %s = %s\n", name, id)
	return nil
}

func (s *store) list() error {
	if len(s.pins) == 0 {
		return fmt.Errorf("no pins in %s", s.cwd)
	}
	for _, p := range s.pins {
		fmt.Printf("%-16s %s\n", p.name, p.id)
	}
	return nil
}

func (s *store) remove(name string) error {
	i := s.indexOf(name)
	if i < 0 {
		return fmt.Errorf("no pin '%s'", name)
	}
	s.pins = slices.Delete(s.pins, i, i+1)
	return s.save()
}

func (s *store) fork(args []string) error {
	var chosen pin
	switch {
	case len(args) > 0 && s.indexOf(args[0]) >= 0:
		chosen = s.pins[s.indexOf(args[0])]
		args = args[1:]
	case len(s.pins) == 0:
		if err := s.pin(nil); err != nil {
			return err
		}
		chosen = s.pins[0]
	case len(s.pins) == 1:
		chosen = s.pins[0]
	default:
		var err error
		if chosen, err = s.choose(); err != nil {
			return err
		}
	}
	return execClaude(append([]string{"-r", chosen.id, "--fork-session"}, args...))
}

func (s *store) choose() (pin, error) {
	if _, err := exec.LookPath("fzf"); err == nil {
		return s.chooseWithFzf()
	}
	return s.chooseWithPrompt()
}

func (s *store) chooseWithFzf() (pin, error) {
	var lines strings.Builder
	for _, p := range s.pins {
		fmt.Fprintf(&lines, "%-16s %s\n", p.name, p.id)
	}
	cmd := exec.Command("fzf", "--height=~10", "--prompt=fork from> ")
	cmd.Stdin = strings.NewReader(lines.String())
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	fields := strings.Fields(string(out))
	if err != nil || len(fields) == 0 {
		return pin{}, errors.New("nothing selected")
	}
	return s.pins[s.indexOf(fields[0])], nil
}

func (s *store) chooseWithPrompt() (pin, error) {
	for i, p := range s.pins {
		fmt.Fprintf(os.Stderr, "%2d) %-16s %s\n", i+1, p.name, p.id)
	}
	fmt.Fprint(os.Stderr, "fork from> ")
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		return pin{}, errors.New("nothing selected")
	}
	answer := strings.TrimSpace(line)
	if n, err := strconv.Atoi(answer); err == nil && n >= 1 && n <= len(s.pins) {
		return s.pins[n-1], nil
	}
	if i := s.indexOf(answer); i >= 0 {
		return s.pins[i], nil
	}
	return pin{}, fmt.Errorf("no pin '%s'", answer)
}
