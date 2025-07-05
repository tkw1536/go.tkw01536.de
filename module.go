package modsite

import (
	"errors"
	"fmt"
	"html/template"
	"strings"
	"unicode"
)

// Module holds information for a go module with a custom VCSPath.
// See [Module.GoImport] and [Module.GoSource].
type Module struct {
	ImportPath string // ImportPath of the module.

	VCSType string
	RepoURL string

	HomePage string // URL to the homepage

	DirectoryLinkTemplate string
	FileLinkTemplate      string
}

// GoImport returns the value of the "go-import" tag
// in the [format supported by the go command].
//
// If the struct contains invalid data, returns the empty string instead.
//
// [supported by the go command]: https://pkg.go.dev/cmd/go#hdr-Remote_import_paths
func (mod Module) GoImport() string {
	if err := mod.ValidateGoImport(); err != nil {
		return ""
	}

	elems := make([]string, 3)
	elems[0] = mod.ImportPath
	elems[1] = mod.VCSType
	elems[2] = mod.RepoURL

	return strings.Join(elems, " ")
}

var (
	errInvalidModuleName            = errors.New("invalid module name")
	errUnsupportedVCS               = errors.New("unsupported VCS")
	errInvalidRepoURL               = errors.New("invalid VCS URL")
	errInvalidRepoSubdirectory      = errors.New("invalid repository directory")
	errInvalidHomePage              = errors.New("invalid home page")
	errInvalidDirectoryLinkTemplate = errors.New("invalid directory link template")
	errInvalidFileLinkTemplate      = errors.New("invalid file link template")
)

var supportedVCS = map[string]struct{}{
	"bzr": {}, "fossil": {}, "git": {}, "hg": {}, "svn": {},
}

// ValidateGoImport checks if this module can generate a "go-module" meta tag
// and returns a non-nil error if not.
func (mod Module) ValidateGoImport() error {
	if mod.ImportPath == "" || strings.ContainsFunc(mod.ImportPath, unicode.IsSpace) {
		return errInvalidModuleName
	}

	// must have a supported VCS
	_, ok := supportedVCS[mod.VCSType]
	if !ok {
		return errUnsupportedVCS
	}

	// must have a valid VCS URL
	if mod.RepoURL == "" || strings.ContainsFunc(mod.RepoURL, unicode.IsSpace) {
		return errInvalidRepoURL
	}

	return nil
}

// GoImportTag renders this module information into a "go-import" meta tag.
// If this module contains invalid information, returns the empty string instead.
//
// See [Module.GoImport] for details.
func (mod Module) GoImportTag() template.HTML {
	content := mod.GoImport()
	if content == "" {
		return ""
	}
	return renderMeta("go-import", content)
}

// GoSource returns the value of the "go-source" meta tag in the [specified format] supported by godoc.
//
// Templates for building a link to a source directory and file are derived from the [Module.DirectoryLinkTemplate] and [Module.FileLinkTemplate] fields.
//
// Both templates may contain the following placeholders:
//
// - "{dir}": The import path with prefix and leading "/" trimmed.
// - "{/dir}": If {dir} is not the empty string, then {/dir} is replaced by "/" + {dir}. Otherwise, {/dir} is replaced with the empty string.
//
// The SourceFileTemplate may additionally contain the following placeholders:
//
// - "{file}": The name of the file.
// - "{line}": The decimal line number.
//
// As a special case, if [Module.HomePage] is the empty string, falls back to the
// defaults of godoc.
//
// If the data is invalid, returns the empty string.
//
// [specified format]: https://github.com/golang/gddo/wiki/Source-Code-Links
func (mod Module) GoSource() string {
	// import path may not have spaces
	if mod.ImportPath == "" || strings.ContainsFunc(mod.ImportPath, unicode.IsSpace) {
		return ""
	}

	// home page may be empty, but must not contain any spaces.
	if mod.HomePage == "" {
		mod.HomePage = "_"
	}
	if strings.ContainsFunc(mod.HomePage, unicode.IsSpace) {
		return ""
	}

	// directory link template must exist and not contain any spaces.
	if mod.DirectoryLinkTemplate == "" || strings.ContainsFunc(mod.DirectoryLinkTemplate, unicode.IsSpace) {
		return ""
	}

	// file link template must exist and not contain any spaces.
	if mod.FileLinkTemplate == "" || strings.ContainsFunc(mod.FileLinkTemplate, unicode.IsSpace) {
		return ""
	}

	return fmt.Sprintf("%s %s %s %s", mod.ImportPath, mod.HomePage, mod.DirectoryLinkTemplate, mod.FileLinkTemplate)
}

// ValidateGoSource checks if this module can generate a "go-source" meta tag
// and returns a non-nil error if not.
func (mod Module) ValidateGoSource() error {
	if mod.ImportPath == "" || strings.ContainsFunc(mod.ImportPath, unicode.IsSpace) {
		return errInvalidModuleName
	}

	if mod.HomePage != "" && strings.ContainsFunc(mod.HomePage, unicode.IsSpace) {
		return errInvalidHomePage
	}

	// directory link template must exist and not contain any spaces.
	if mod.DirectoryLinkTemplate == "" || strings.ContainsFunc(mod.DirectoryLinkTemplate, unicode.IsSpace) {
		return errInvalidDirectoryLinkTemplate
	}

	// file link template must exist and not contain any spaces.
	if mod.FileLinkTemplate == "" || strings.ContainsFunc(mod.FileLinkTemplate, unicode.IsSpace) {
		return errInvalidFileLinkTemplate
	}

	return nil
}

// GoSourceTag renders this module information into a "go-source" meta tag.
// If this module contains invalid information, returns the empty string instead.
//
// See [Module.GoSource] for details.
func (mod Module) GoSourceTag() template.HTML {
	content := mod.GoSource()
	if content == "" {
		return ""
	}
	return renderMeta("go-source", content)
}

var metaTemplate = template.Must(template.New("").Parse(`<meta name="{{.Name}}" value="{{.Value}}">`))

type metaTemplateContext struct {
	Name  string
	Value string
}

// renderMeta renders a meta tag with the given name and value.
func renderMeta(name string, value string) template.HTML {
	var builder strings.Builder

	if err := metaTemplate.Execute(&builder, metaTemplateContext{Name: name, Value: value}); err != nil {
		panic(fmt.Sprintf("metaTemplate failed: %v", err))
	}

	return template.HTML(builder.String())
}
