package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/wercker/journalhook"
)

func titusHook(bundlePath string) error {
	journalhook.Enable()

	logrus.Debugf("Using bundle file: %s\n", bundlePath)
	jsonFile, err := os.OpenFile(bundlePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("Couldn't open OCI spec file: %w", err)
	}
	defer jsonFile.Close()

	jsonContent, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return fmt.Errorf("Couldn't read OCI spec file: %w", err)
	}
	//logrus.Infof("My json: %s", jsonContent)
	var spec specs.Spec
	err = json.Unmarshal(jsonContent, &spec)
	if err != nil {
		return fmt.Errorf("Couldn't unmarshal OCI spec file: %w", err)
	}

	tiniCommand := []string{"/dev/titus-init", "-v", "-v", "-v", spec.Process.Args[0], "--"}
	if len(spec.Process.Args) > 1 {
		spec.Process.Args = append(tiniCommand, spec.Process.Args[1:]...)
	}
	logrus.Infof("INJECTING TINI! New args are %+v", spec.Process.Args)
	tiniMount := specs.Mount{
		Destination: "/dev/titus-init",
		Options: []string{
			"rbind",
			"rprivate",
			"ro",
		},
		Source: "/apps/titus-executor/bin/tini-static",
		Type:   "bind",
	}
	spec.Mounts = append(spec.Mounts, tiniMount)

	jsonOutput, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("Couldn't marshal OCI spec file: %w", err)
	}
	_, err = jsonFile.WriteAt(jsonOutput, 0)
	if err != nil {
		return fmt.Errorf("Couldn't write OCI spec file: %w", err)
	}

	return nil
}
