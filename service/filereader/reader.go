// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package filereader

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/yldio/acto/internal/actoerrors"
)

const (
	FileHCLExt  = ".hcl"
	FileYAMLExt = ".yaml"
)

type Reader struct {
	parser    *hclparse.Parser
	path      string
	files     []string
	recursive bool
}

func (read *Reader) GetFiles() []string {
	return read.files
}

func New(path string, recursive bool) *Reader {
	return &Reader{
		parser:    hclparse.NewParser(),
		path:      path,
		recursive: recursive,
	}
}

func (read *Reader) ReadHclSrc(src []byte, filename string) (hcl.Body, error) {
	hclFile, diags := read.parser.ParseHCL(src, filename)
	if diags.HasErrors() {
		var body hcl.Body
		return body, actoerrors.ProcessHCLDiags(diags)
	}

	return hclFile.Body, nil
}

func (read *Reader) ReadHclFile(filename string) (hcl.Body, error) {
	file, diags := read.parser.ParseHCLFile(filename)
	if diags.HasErrors() {
		var body hcl.Body
		return body, actoerrors.ProcessHCLDiags(diags)
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

func (read *Reader) readHclFile() ([]hcl.Body, error) {
	var emptyBody []hcl.Body

	bodyFile, err := read.ReadHclFile(read.path)
	if err != nil {
		return emptyBody, err
	}

	return []hcl.Body{bodyFile}, nil
}

func (read *Reader) readFile() ([]hcl.Body, error) {
	var emptyBody []hcl.Body

	_, err := os.Stat(read.path)
	if err != nil {
		return emptyBody, err
	}

	fileExt := filepath.Ext(read.path)

	switch fileExt {
	case FileHCLExt:
		return read.readHclFile()
	default:
		return emptyBody, fmt.Errorf("unrecognized file extention '%s'", fileExt)
	}
}

func (read *Reader) readDirectory() ([]hcl.Body, error) {
	var emptyBody []hcl.Body

	files, err := os.ReadDir(read.path)
	if err != nil {
		return emptyBody, err
	}

	list, err := readDir(read.path, files, read.recursive, FileHCLExt)
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
}

func (read *Reader) Do(fileExt string) ([]hcl.Body, error) {
	var emptyBody []hcl.Body

	info, err := os.Stat(read.path)
	if err != nil {
		return emptyBody, err
	}

	if info.IsDir() {
		return read.readDirectory()
	} else {
		return read.readFile()
	}
}

func readDir(parentDirectory string, files []fs.DirEntry, recursive bool, allowedExt string) ([]string, error) {
	var listOfFiles []string
	for _, file := range files {
		fullpath := filepath.Join(parentDirectory, file.Name())

		if !file.IsDir() {
			if filepath.Ext(file.Name()) != allowedExt {
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

		listOfSubFiles, err := readDir(fullpath, subFiles, recursive, FileHCLExt)
		if err != nil {
			return []string{}, err
		}

		listOfFiles = append(listOfFiles, listOfSubFiles...)
	}
	return listOfFiles, nil
}

func (read *Reader) ReadDir(allowedExt string, path ...string) error {
	curPath := read.path

	if len(path) > 0 {
		curPath = path[0]
	}

	info, err := os.Stat(curPath)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		if filepath.Ext(curPath) != allowedExt {
			return nil
		}

		read.files = append(read.files, curPath)

		return nil
	}

	files, err := os.ReadDir(curPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		fullpath := filepath.Join(curPath, file.Name())

		if err := read.ReadDir(allowedExt, fullpath); err != nil {
			return err
		}
	}

	return nil
}
