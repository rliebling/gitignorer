package gitignorer

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rliebling/env"
)

type StackedFilter struct {
	dir string
	PathFilter
}

type ComposedFilter struct {
	filters []StackedFilter
}

func NewComposedFilter() (*ComposedFilter, error) {
	global, _ := openGlobalGitignore()
	project, _ := openParentGitignore()

	globalFilter := &GitFilter{}
	projectFilter := &GitFilter{}
	if global != nil {
		globalFilter, _ = NewFilterFromReader(global)
	}
	if project != nil {
		projectFilter, _ = NewFilterFromReader(project)
	}
	return &ComposedFilter{filters: []StackedFilter{StackedFilter{"", globalFilter}, StackedFilter{"", projectFilter}}}, nil
}

func (f *ComposedFilter) ComposeLocalGitignore(path string) {
	f.popOutOfScope(path)
	f.pushLocalFilter(path)
}

func (f *ComposedFilter) popOutOfScope(path string) {
	lastToKeep := len(f.filters) - 1
	for i := lastToKeep; i >= 0; i-- {
		filter := f.filters[i]
		if strings.HasPrefix(path, filter.dir) {
			break
		}
		lastToKeep--
	}
	if 0 <= lastToKeep && lastToKeep < len(f.filters)-1 {
		f.filters = f.filters[0 : lastToKeep+1]
	}
}
func (f *ComposedFilter) pushLocalFilter(path string) {
	gitignore, err := os.Open(filepath.Join(path, ".gitignore"))
	if err != nil {
		return
	}

	filter, _ := NewFilterFromReader(gitignore)
	f.filters = append(f.filters, StackedFilter{path, filter})

}

func (f *ComposedFilter) Match(path string) bool {
	for _, stackedFilter := range f.filters {
		if stackedFilter.Match(path) {
			return true
		}
	}
	return false
}

func openGlobalGitignore() (io.Reader, error) {
	homeDir, _ := env.GetHomedir()
	return os.Open(filepath.Join(homeDir, ".gitignore"))
}

func openParentGitignore() (io.Reader, error) {
	// should scan up the dir tree to root for a .gitignore
	wd, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(wd, ".gitignore")); err == nil {
		return nil, nil // nothing to do if there's one here.  will be picked up in walk
	}
	parent, _ := filepath.Split(wd)
	for parent != "" {
		f, err := os.Open(filepath.Join(parent, ".gitignore"))
		if err == nil {
			return f, nil
		}
		parent, _ = filepath.Split(parent[0 : len(parent)-1])
	}
	return nil, nil
}
