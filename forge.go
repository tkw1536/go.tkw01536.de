package modsite

import "strings"

// GitForgeModule returns a new [Module] using templates for a common git forge.
//
// ImportPath is the import path of the module.
// RepoURL is the https clone url of the repository.
// Ref should be the git reflike pointing to HEAD. This is typically a branch name.
func GitForgeModule(ImportPath string, RepoURL string, ref string) Module {
	url := strings.TrimSuffix(RepoURL, ".git")
	url = strings.TrimSuffix(url, "/")

	return Module{
		ImportPath: ImportPath,

		VCSType: "git",
		RepoURL: url,

		HomePage: url,

		DirectoryLinkTemplate: url + "/tree/" + ref + "{/dir}",
		FileLinkTemplate:      url + "/blob/" + ref + "{/dir}/{file}#L{line}",
	}
}
