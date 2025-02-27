// Copyright (c) 2024-2025 YLD Limited
// SPDX-License-Identifier: MIT

package main

// func getOutputDirectory(flags *actoflag.ActoCli) (string, error) {
// 	var outputPath string
// 	var err error

// 	if flags.OutputDirectory != "" {
// 		outputPath = flags.OutputDirectory
// 	}

// 	if !filepath.IsAbs(outputPath) {
// 		outputPath, err = filepath.Abs(outputPath)
// 		if err != nil {
// 			return "", err
// 		}
// 	}

// 	if outputPath == "" {
// 		currentDirectory, err := os.Getwd()
// 		if err != nil {
// 			return "", err
// 		}
// 		outputPath = currentDirectory
// 	}

// 	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
// 		return "", err
// 	}

// 	return outputPath, nil
// }

// func parse(flags *actoflag.ActoCli) error {
// 	path, err := getPath(flags)
// 	if err != nil {
// 		return err
// 	}

// 	fileReader := filereader.New()

// 	hclBody, err := fileReader.DoHcl(path, flags.Recursive)
// 	if err != nil {
// 		return err
// 	}

// 	hclParser := hclparser.NewGitHub(hclBody)

// 	parsedWorkflows, err := hclParser.Do()
// 	if err != nil {
// 		return err
// 	}

// 	yamlwriter := yamlwriter.New(parsedWorkflows)

// 	listOfFiles, err := yamlwriter.Do()
// 	if err != nil {
// 		return err
// 	}

// 	outputDirectory, err := getOutputDirectory(flags)
// 	if err != nil {
// 		return err
// 	}

// 	fw := filewriter.New()

// 	for file, bytes := range listOfFiles {
// 		filePath := filepath.Join(outputDirectory, file)

// 		if flags.DryRun {
// 			fmt.Printf("# file: %s\n", filePath)
// 			fmt.Println(string(bytes))
// 			continue
// 		}

// 		if err := fw.Do(filePath, bytes); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func decode(flags *actoflag.ActoCli) error {
// 	path, err := getPath(flags)
// 	if err != nil {
// 		return err
// 	}

// 	fileReader := filereader.New()
// 	parsedWorkflows, err := fileReader.DoYaml(path, flags.Recursive)
// 	if err != nil {
// 		return err
// 	}

// 	outputDirectory, err := getOutputDirectory(flags)
// 	if err != nil {
// 		return err
// 	}

// 	fw := filewriter.New()

// 	for _, parsedWorkflow := range parsedWorkflows {
// 		bytes, err := parsedWorkflow.Decode()
// 		if err != nil {
// 			return err
// 		}

// 		file := fmt.Sprintf("%s%s", parsedWorkflow.Filename, filereader.FileHclExtensions[0])
// 		filePath := filepath.Join(outputDirectory, file)

// 		if flags.DryRun {
// 			fmt.Printf("# file: %s\n", filePath)
// 			fmt.Println(string(bytes))
// 			continue
// 		}

// 		if err := fw.Do(filePath, bytes); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
