package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	BASE_MODULE    = "github.com/hashgraph/hedera-protobufs-go"
	COMMON_MODULE  = BASE_MODULE + "/common"
	PROTO_DIR      = "proto"
	PROTO_REPO     = "https://github.com/hashgraph/hedera-protobufs"
	PROTO_REVISION = "8c27786cec93abab974309074feaef9b48a695b7"
)

// Common module contains proto files that are commonly referenced and could cause cyclic references
var common []string = []string{
	"timestamp.proto",
	"basic_types.proto",
}

// Proto modules that we can skip
var modulesToSkip []string = []string{
	"mirror", // we are not going to call any mirror methods
}

func main() {
	// get project directory
	// nolint:dogsled
	_, buildFilename, _, _ := runtime.Caller(0)
	projectDir := path.Join(buildFilename, "../../../..")

	// remove all existing files
	removeAllWithExt(projectDir, ".pb.go")

	// clone repository
	cloneRepo(projectDir)
	defer func() {
		repoPath := projectDir + "/" + PROTO_DIR
		if _, err := os.Stat(repoPath); err == nil {
			err := os.RemoveAll(repoPath)
			if err != nil {
				panic(err)
			}
		}
	}()

	// invoke the build command for this module
	buildProto(projectDir)
}

func cloneRepo(dir string) {
	cmdArguments := []string{
		"clone",
		PROTO_REPO,
		"--depth=1",
		"--revision",
		PROTO_REVISION,
		PROTO_DIR,
	}

	cmd := exec.Command("git", cmdArguments...)
	cmd.Dir = dir

	mustRunCommand(cmd)
}

func skipFile(pathFromRoot string) bool {
	if !strings.HasSuffix(pathFromRoot, ".proto") {
		return true
	}
	for _, m := range modulesToSkip {
		protoModule := PROTO_DIR + "/" + m
		if strings.HasPrefix(pathFromRoot, protoModule) {
			return true
		}
	}
	return false
}

// Adjust proto files and create protoc command
func buildProto(dir string) {
	// collect all proto files
	var servicesProtoFiles []string
	err := filepath.Walk(path.Join(dir, "proto"), func(filename string, info fs.FileInfo, err error) error {
		pathFromRoot := strings.TrimPrefix(filename, dir+"/")
		if skipFile(pathFromRoot) {
			return nil
		}

		// make sure that each proto file contains "option go_package"
		err = appendGoModule(pathFromRoot, filename)
		if err != nil {
			panic(err)
		}

		servicesProtoFiles = append(servicesProtoFiles, pathFromRoot)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// skip base module directiores
	goOptDir := fmt.Sprintf("--go_opt=module=%s", BASE_MODULE)
	goGrpcOptDir := fmt.Sprintf("--go-grpc_opt=module=%s", BASE_MODULE)
	cmdArguments := []string{
		"--go_out=.",
		goOptDir,
		"--go-grpc_out=.",
		goGrpcOptDir,
		"-Iproto/services",
		"-Iproto/block",
		"-Iproto/sdk",
		"-Iproto/streams",
		"-Iproto/platform",
		"-Iproto",
	}
	cmdArguments = append(cmdArguments, servicesProtoFiles...)
	cmd := exec.Command("protoc", cmdArguments...)
	cmd.Dir = dir

	// generate proto files
	mustRunCommand(cmd)
}

func appendGoModule(pathFromRoot, fullPath string) error {
	isCommon := false
	for _, c := range common {
		isCommon = isCommon || strings.HasSuffix(pathFromRoot, c)
	}
	var targetModule string
	if isCommon {
		targetModule = COMMON_MODULE
	} else {
		targetModule = getRelativeModule(pathFromRoot)
	}

	return appendModule(fullPath, targetModule)
}

func getRelativeModule(pathFromRoot string) string {
	pathParts := strings.Split(pathFromRoot, "/")
	finalMod := BASE_MODULE
	for _, submod := range pathParts {
		if submod == PROTO_DIR || strings.Contains(submod, ".proto") {
			continue
		}
		finalMod = finalMod + "/" + submod
	}

	return finalMod
}

func appendModule(fullPath, mod string) error {
	goOptionPackage := fmt.Sprintf(`option go_package = "%s";`, mod)
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for write: %w", err)
	}
	defer f.Close()

	if _, err = f.WriteString(goOptionPackage); err != nil {
		return fmt.Errorf("failed to append goOptionPackage: %w", err)
	}

	return nil
}

func mustRunCommand(cmd *exec.Cmd) {
	_, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			fmt.Print(string(exitErr.Stderr))
			os.Exit(exitErr.ExitCode())
		}

		panic(err)
	}
}

func removeAllWithExt(dir string, ext string) {
	err := filepath.Walk(dir, func(filename string, info fs.FileInfo, err error) error {
		if strings.HasSuffix(filename, ext) {
			err := os.Remove(filename)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
}
