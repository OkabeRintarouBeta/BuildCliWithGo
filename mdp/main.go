package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

const (
	defaultTemplate = `<!DOCTYPE html><html><head><meta http-equiv="content-type" content="text/html; charset=utf-8"><title>{{ .Title }}</title> </head><body>{{ .Body }}</body></html>`
)

type content struct {
	Title string
	Body  template.HTML
}

func main() {
	filename := flag.String("file", "", "Markdown file to preview")
	skipPreview := flag.Bool("s", false, "Skip auto-preview")
	templateFile := flag.String("t", "", "Alternate template name")
	flag.Parse()

	if *filename == "" {
		flag.Usage()
		os.Exit(1)
	}

	if err := run(*filename, *templateFile, os.Stdout, *skipPreview); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(filename, templateFile string, writer io.Writer, skipPreview bool) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	htmlData, err := parseContent(input, templateFile)
	if err != nil {
		return err
	}

	// Create temp files and check for error
	temp, err := os.CreateTemp("", "mdp-*.html")
	if err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	outName := temp.Name()
	fmt.Fprintln(writer, outName)

	if err := saveHTML(outName, htmlData); err != nil {
		return err
	}
	if skipPreview {
		return nil
	}
	defer os.Remove(outName)
	return preview(outName)
}

func parseContent(input []byte, templateFile string) ([]byte, error) {

	output := blackfriday.MarkdownCommon(input)
	body := bluemonday.UGCPolicy().SanitizeBytes(output)
	t, err := template.New("mdp").Parse(defaultTemplate)
	if err != nil {
		return nil, err
	}
	if templateFile != "" {
		t, err = template.ParseFiles(templateFile)
		if err != nil {
			return nil, err
		}
	}
	templateContent := content{
		Title: "Markdown Preview Tool",
		Body:  template.HTML(body),
	}
	var buffer bytes.Buffer
	if err := t.Execute(&buffer, templateContent); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func saveHTML(filename string, input []byte) error {
	return os.WriteFile(filename, input, 0644)
}

func preview(fname string) error {
	cName := ""
	cParams := []string{}

	switch runtime.GOOS {
	case "linux":
		cName = "xdg-open"
	case "windows":
		cName = "cmd.exe"
		cParams = []string{"/C", "start"}
	case "darwin":
		cName = "open"
	default:
		return fmt.Errorf("OS not supported")
	}
	cParams = append(cParams, fname)
	cPath, err := exec.LookPath(cName)
	if err != nil {
		return err
	}

	err = exec.Command(cPath, cParams...).Run()
	// Give the browser some time to open the file before deleting it
	time.Sleep(2 * time.Second)
	return err
}
