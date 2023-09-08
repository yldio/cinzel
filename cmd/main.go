package main

import (
	"fmt"
	"os"

	"github.com/yldio/atos/internal/flags"
	"github.com/yldio/atos/internal/parsers"
	"github.com/yldio/atos/internal/reader"
)

func main() {
	flag := flags.NewParseFlags()

	atosReader := reader.NewReader(flag.Directory, flag.File, flag.Recursive)

	if err := atosReader.Do(); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	atosHclParser := parsers.NewHclParser()

	hclBodies := atosReader.ReturnHclBodies()

	if err := atosHclParser.ParseFiles(hclBodies); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	if err := atosHclParser.Do(); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	atosYamlParser := parsers.NewYamlParser()

	content := atosHclParser.GetContent()

	if err := atosYamlParser.ParseToYaml(content); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	if err := atosYamlParser.Do(); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	for _, yaml := range atosYamlParser.GetContent() {
		fmt.Println(string(yaml))
	}

	fmt.Println("<3 atos finished converting!")
}
