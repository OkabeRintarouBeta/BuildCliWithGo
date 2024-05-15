package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type config struct {
	// extension to filter out
	ext string
	// min file size
	size int64
	// list files
	list bool
	// delete files
	del bool
	// log destination writer
	wLog io.Writer
	// archive diretory
	archive string
}

func main() {
	// Parsing command line flags
	root := flag.String("root", ".", "Root directory to start")
	// Action options
	list := flag.Bool("list", false, "List files only")
	del := flag.Bool("del", false, "Delete files")
	// Filter options
	ext := flag.String("ext", "", "File extension to filter out")
	size := flag.Int64("size", 0, "Minimum file size")
	// Log option
	logfile := flag.String("log", "", "Log deletes to this file")
	archive := flag.String("archive", "", "Archive directory")

	flag.Parse()

	var (
		f   = os.Stdout
		err error
	)

	if *logfile != "" {
		f, err = os.OpenFile(*logfile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	cfg := config{
		ext:     *ext,
		size:    *size,
		list:    *list,
		del:     *del,
		wLog:    f,
		archive: *archive,
	}
	if err := run(*root, os.Stdout, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(root string, outWriter io.Writer, cfg config) error {
	delLogger := log.New(cfg.wLog, "DELETED FILE:", log.LstdFlags)
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filterOut(path, cfg, info) {
			return nil
		}
		if cfg.list {
			return listFiles(path, outWriter)
		}
		if cfg.archive != "" {
			if err := archiveFile(cfg.archive, root, path); err != nil {
				return err
			}
		}
		if cfg.del {
			return delFile(path, delLogger)
		}
		return listFiles(path, outWriter)
	})
}
