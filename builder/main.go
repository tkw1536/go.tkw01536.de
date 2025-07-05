package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.tkw01536.de/modsite"
)

var modules = []modsite.Module{
	modsite.GitForgeModule("go.tkw01536.de/modsite", "https://github.com/tkw1536/go.tkw01536.de", "main"),
	modsite.GitForgeModule("go.tkw01536.de/pkglib", "https://github.com/tkw1536/pkglib", "main"),
}

const (
	outDir = "public"
	domain = "go.tkw01536.de"
	index  = `This domain is used to serve <a href="https://tkw01536.de" rel="me" target="_blank">my personal</a> go modules.`
	footer = `For legal reasons I must link <a rel="privacy-policy" href="https://inform.everyone.wtf" target="_blank">my Privacy Policy and Imprint</a>.`
)

func main() {
	content, err := modsite.BuildSite(
		domain,
		index,
		footer,
		modules,
	)
	if err != nil {
		log.Fatal("failed to build site: %w", err)
	}

	if err := writeFiles(outDir, content); err != nil {
		log.Fatal("failed to write files: %w", err)
	}
}

func writeFiles(dest string, files map[string]string) error {
	if err := os.RemoveAll(dest); err != nil {
		return fmt.Errorf("failed to remove destination directory: %w", err)
	}
	for path, data := range files {
		file := filepath.Join(dest, path)

		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %q: %w", dir, err)
		}

		log.Printf("writing %q", file)
		if err := os.WriteFile(file, []byte(data), os.ModePerm); err != nil {
			return fmt.Errorf("failed to create file %q: %w", file, err)
		}
	}
	return nil
}
