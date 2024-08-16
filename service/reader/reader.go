package reader

import (
	"errors"
	"flag"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
)

const (
	allowedExtension = ".hcl"
)

type Reader struct {
	parser    *hclparse.Parser
	directory string
	file      string
	recursive bool
}

func New(directory string, file string, recursive bool) *Reader {
	return &Reader{
		parser:    hclparse.NewParser(),
		directory: directory,
		file:      file,
		recursive: recursive,
	}
}

func (read *Reader) ReadHclSrc(src []byte, filename string) (hcl.Body, error) {
	hclFile, diags := read.parser.ParseHCL(src, filename)
	if diags.HasErrors() {
		var body hcl.Body
		return body, errors.New(diags.Error())
	}

	return hclFile.Body, nil
}

func (read *Reader) ReadHclFile(filename string) (hcl.Body, error) {
	file, diags := read.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		var body hcl.Body
		return body, errors.New(diags.Error())
	}

	return file.Body, nil
}

func (read *Reader) ReturnHclBodies() []hcl.Body {
	files := read.parser.Files()

	var bodies []hcl.Body
	for _, file := range files {
		bodies = append(bodies, file.Body)
	}

	return bodies
}

func (read *Reader) Do() ([]hcl.Body, error) {
	var emptyBody []hcl.Body

	if read.file != "" {
		_, err := os.Stat(read.file)
		if err != nil {
			return emptyBody, err
		}

		if filepath.Ext(read.file) != allowedExtension {
			return emptyBody, errors.New("only allowed .hcl files")
		}

		bodyFile, err := read.ReadHclFile(read.file)
		if err != nil {
			return emptyBody, err
		}

		return []hcl.Body{bodyFile}, nil
	} else if read.directory != "" {
		files, err := os.ReadDir(read.directory)
		if err != nil {
			return emptyBody, err
		}

		list, err := recurDir(read.directory, files, read.recursive)
		if err != nil {
			return emptyBody, err
		}

		var bodies []hcl.Body
		for _, file := range list {
			bodyFile, err := read.ReadHclFile(file)
			if err != nil {
				return emptyBody, err
			}
			bodies = append(bodies, bodyFile)
		}

		return bodies, nil
	} else {
		flag.Usage()
		os.Exit(0)
	}

	return emptyBody, nil
}

func recurDir(parentDirectory string, files []fs.DirEntry, recursive bool) ([]string, error) {
	var listOfFiles []string
	for _, file := range files {
		fullpath := filepath.Join(parentDirectory, file.Name())

		if !file.IsDir() {
			if filepath.Ext(file.Name()) != allowedExtension {
				continue
			}

			listOfFiles = append(listOfFiles, fullpath)
			continue
		}

		if !recursive {
			continue
		}

		subFiles, err := os.ReadDir(fullpath)
		if err != nil {
			return []string{}, err
		}

		listOfSubFiles, err := recurDir(fullpath, subFiles, recursive)
		if err != nil {
			return []string{}, err
		}

		listOfFiles = append(listOfFiles, listOfSubFiles...)
	}
	return listOfFiles, nil
}
