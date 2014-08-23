package gitignorer

import (
	"bufio"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rliebling/env"
)

type PathFilter interface {
	Match(path string) bool
}

type FilterPattern struct {
	Pattern regexp.Regexp
	Include bool
}

type GitFilter struct {
	patterns []FilterPattern
}

func NewFilter() (*GitFilter, error) {
	r := findGlobalGitignore()
	return NewFilterFromReader(r)
}

func NewFilterFromReader(content io.Reader) (*GitFilter, error) {
	scanner := bufio.NewScanner(content)
	gf := GitFilter{}

	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if len(pattern) == 0 || pattern[0] == '#' {
			continue
		}
		p := parsePattern(pattern)
		re, err := regexp.Compile(p.Regex)
		if err != nil {
			return nil, err
		}

		gf.patterns = append(gf.patterns, FilterPattern{Pattern: *re, Include: p.Include})
	}
	log.Printf("GitIgnoreFilter: %+v\n", gf)
	return &gf, nil
}

func (gf GitFilter) Match(path string) bool {
	matches := false
	for _, p := range gf.patterns {
		isRegexMatch := p.Pattern.MatchString(path)
		if p.Include && isRegexMatch {
			matches = false
		} else if isRegexMatch {
			matches = true
		}
	}
	if matches {
		log.Printf("filtering out: %v\n", path)
	}
	return matches
}

func findGlobalGitignore() io.Reader {
	homeDir, _ := env.GetHomedir()
	f, _ := os.Open(filepath.Join(homeDir, ".gitignore"))
	return f
}
