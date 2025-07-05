package modsite

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	_ "embed"
)

//go:embed "site_module.html"
var siteModuleHTML string

var siteModuleTemplate = template.Must(template.New("site_module.html").Parse(siteModuleHTML))

type siteTemplateContext struct {
	Module
	Footer template.HTML
}

// RenderSite renders the site template for the given writer and footer
func (mod Module) RenderSite(footer template.HTML, w io.Writer) error {
	if err := mod.ValidateGoImport(); err != nil {
		return fmt.Errorf("site does not have a valid go import: %w", err)
	}
	if err := siteModuleTemplate.Execute(w, siteTemplateContext{Module: mod, Footer: footer}); err != nil {
		return fmt.Errorf("error importing template: %w", err)
	}
	return nil
}

func (mod Module) HomePageURL() string {
	if mod.HomePage != "" && mod.HomePage != "_" {
		return mod.HomePage
	}
	return "https://pkg.go.dev/" + mod.ImportPath
}

func (mod Module) SourceCodeURL() string {
	if mod.DirectoryLinkTemplate == "" {
		return ""
	}
	if mod.HomePage != "" && mod.HomePage != "_" {
		return mod.HomePage
	}
	return "https://pkg.go.dev/" + mod.ImportPath
}

// BuildSite builds a map of filepath to html content for an html site containing the given modules.
func BuildSite(url string, indexContent template.HTML, footerContent template.HTML, modules []Module) (files map[string]string, err error) {
	files = make(map[string]string, len(files)+1)

	urls := make(map[string]string, len(files))

	var builder strings.Builder
	for _, mod := range modules {
		dest, ok := getModuleHTMLPath(mod, url)
		if !ok {
			return nil, fmt.Errorf("failed to get module path for %q", mod.ImportPath)
		}
		urls[mod.ImportPath] = strings.TrimSuffix(dest, "index.html")

		if _, ok := files[dest]; ok {
			return nil, fmt.Errorf("duplicate destination path %q", dest)
		}

		builder.Reset()
		if err := mod.RenderSite(footerContent, &builder); err != nil {
			return nil, fmt.Errorf("failed to render html for %q: %w", mod.ImportPath, err)
		}

		files[dest] = builder.String()
	}

	// build an index page if it doesn't exist
	if _, ok := files["index.html"]; !ok {
		builder.Reset()

		if err := buildIndex(url, indexContent, footerContent, urls, &builder); err != nil {
			return nil, fmt.Errorf("failed to render index content: %w", err)
		}
		files["index.html"] = builder.String()
	}
	return files, nil
}

func getModuleHTMLPath(mod Module, base string) (path string, ok bool) {
	realBase := strings.TrimSuffix(base, "/")
	realBase = strings.TrimPrefix(realBase, "/")
	if !strings.HasPrefix(mod.ImportPath, realBase) {
		return "", false
	}

	path = strings.TrimSuffix(mod.ImportPath[len(realBase):], "/")
	path = filepath.Join(path, "index.html")
	return path, true
}

//go:embed "site_index.html"
var siteIndexHTML string

var siteIndexTemplate = template.Must(template.New("site_index.html").Parse(siteIndexHTML))

type siteIndexContext struct {
	URL     string
	Modules map[string]string // map from module name to local URL
	Content template.HTML
	Footer  template.HTML
}

func buildIndex(url string, content template.HTML, footer template.HTML, modules map[string]string, w io.Writer) error {
	if err := siteIndexTemplate.Execute(w, siteIndexContext{
		URL:     url,
		Modules: modules,
		Content: content,
		Footer:  footer,
	}); err != nil {
		return fmt.Errorf("failed to build index: %w", err)
	}
	return nil
}
