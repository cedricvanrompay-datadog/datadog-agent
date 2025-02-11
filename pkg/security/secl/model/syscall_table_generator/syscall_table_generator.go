// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"
)

func main() {
	var (
		inputTableURL      string
		outputEnumPath     string
		outputStringerPath string
		abis               string
	)

	flag.StringVar(&inputTableURL, "table-url", "", "URL of the table to use for the generation")
	flag.StringVar(&outputEnumPath, "output", "", "Output path of the generated file with the constant declarations")
	flag.StringVar(&outputStringerPath, "output-string", "", "Output path of the generated file with the stringer code")
	flag.StringVar(&abis, "abis", "", "Comma separated list of ABIs to keep")
	flag.Parse()

	if inputTableURL == "" || outputEnumPath == "" || outputStringerPath == "" {
		fmt.Fprintf(os.Stderr, "Please provide required flags\n")
		flag.Usage()
		return
	}

	abiList := strings.Split(abis, ",")

	syscalls, err := getSyscallTable(inputTableURL, abiList)
	if err != nil {
		panic(err)
	}

	outputContent, err := generateEnumCode(syscalls)
	if err != nil {
		panic(err)
	}

	if err := writeFileAndFormat(outputEnumPath, outputContent); err != nil {
		panic(err)
	}

	if err := generateStringer(outputEnumPath, outputStringerPath); err != nil {
		panic(err)
	}
}

type syscallDefinition struct {
	Number        int
	Abi           string
	Name          string
	CamelCaseName string
}

func getSyscallTable(url string, abis []string) ([]syscallDefinition, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	syscalls := make([]syscallDefinition, 0)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.Fields(trimmed)
		if len(parts) < 3 {
			return nil, errors.New("found syscall with missing fields")
		}

		number, err := strconv.ParseInt(parts[0], 10, 0)
		if err != nil {
			return nil, err
		}
		abi := parts[1]
		name := parts[2]
		camelCaseName := snakeToCamelCase(name)

		if containsStringSlice(abis, abi) {
			syscalls = append(syscalls, syscallDefinition{
				Number:        int(number),
				Abi:           abi,
				Name:          name,
				CamelCaseName: camelCaseName,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return syscalls, nil
}

const outputTemplateContent = `
// Code generated - DO NOT EDIT.
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// +build linux

package model

// Syscall represents a syscall identifier
type Syscall int

// Linux syscall identifiers
const (
	{{- range .}}
	Sys{{.CamelCaseName}} Syscall = {{.Number}}
	{{- end}}
)
`

func generateEnumCode(syscalls []syscallDefinition) (string, error) {
	tmpl, err := template.New("enum-code").Parse(outputTemplateContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, syscalls); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func snakeToCamelCase(snake string) string {
	parts := strings.Split(snake, "_")

	var b strings.Builder
	for _, part := range parts {
		b.WriteString(strings.Title(part))
	}

	return b.String()
}

func writeFileAndFormat(outputPath string, content string) error {
	tmpfile, err := os.CreateTemp(path.Dir(outputPath), "syscalls-enum")
	if err != nil {
		return err
	}

	if _, err := tmpfile.WriteString(content); err != nil {
		return err
	}

	if err := tmpfile.Close(); err != nil {
		return err
	}

	cmd := exec.Command("gofmt", "-s", "-w", tmpfile.Name())
	if err := cmd.Run(); err != nil {
		return err
	}

	return os.Rename(tmpfile.Name(), outputPath)
}

func generateStringer(inputPath, outputPath string) error {
	return exec.Command("go", "run", "golang.org/x/tools/cmd/stringer", "-type", "Syscall", "-output", outputPath, inputPath).Run()
}

func containsStringSlice(slice []string, value string) bool {
	for _, current := range slice {
		if current == value {
			return true
		}
	}
	return false
}
