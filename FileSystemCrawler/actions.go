package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

// Checks if the given path needs to be filtered out from the results based on:
//  1. whether it points to a directory
//  2. File size less than minimum
//  3. Extension mismatch
func filterOut(path string, cfg config, info fs.FileInfo) bool {

	if info.IsDir() || info.Size() < cfg.size {
		return true
	}
	if cfg.ext != "" && filepath.Ext(path) != cfg.ext {
		return true
	}
	return false
}

func listFiles(path string, outWriter io.Writer) error {
	_, err := fmt.Fprintln(outWriter, path)
	return err
}

func delFile(path string, delLogger *log.Logger) error {
	if err := os.Remove(path); err != nil {
		return err
	}
	delLogger.Println(path)
	return nil
}

func archiveFile(destDir, root, path string) error {
	info, err := os.Stat(destDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", destDir)
	}
	relDir, err := filepath.Rel(root, filepath.Dir(path))
	if err != nil {
		return err
	}
	dest := fmt.Sprintf("%s.gz", filepath.Base(path))
	targetPath := filepath.Join(destDir, relDir, dest)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer out.Close()

	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()

	zw := gzip.NewWriter(out)
	zw.Name = filepath.Base(path)
	if _, err := io.Copy(zw, in); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	return out.Close()
}
