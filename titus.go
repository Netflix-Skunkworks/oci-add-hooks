package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/namespaces"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/wercker/journalhook"
)

func isSystemdImage(id string) bool {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = namespaces.WithNamespace(ctx, "k8s.io")

	client, err := containerd.New("/run/containerd/containerd.sock")
	if err != nil {
		logrus.Error(err)
	}
	container, err := client.LoadContainer(ctx, id)
	if err != nil {
		logrus.Error(err)
		return false
	} else {
		var (
			ociimage v1.Image
			//config   v1.ImageConfig
		)
		image, _ := container.Image(ctx)
		ic, _ := image.Config(ctx)
		p, _ := content.ReadBlob(ctx, image.ContentStore(), ic)
		_ = json.Unmarshal(p, &ociimage)
		labels := ociimage.Config.Labels
		//labels := image.Metadata().Labels
		logrus.Infof("Got conatiner id %s with image %s has labels %+v", id, image.Name(), labels)
		_, ok := labels["com.netflix.titus.systemd"]
		return ok
	}
}

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
	var spec specs.Spec
	err = json.Unmarshal(jsonContent, &spec)
	if err != nil {
		return fmt.Errorf("Couldn't unmarshal OCI spec file: %w", err)
	}

	id := path.Base(spec.Linux.CgroupsPath)
	if isSystemdImage(id) {
		logrus.Infof("%s is using a systemd-ready image. Setting TINI_HANDOFF", id)
		spec.Process.Env = append(spec.Process.Env, "TINI_HANDOFF=true")
	}

	tiniCommand := []string{"/dev/titus-init", "-v", "-v", "-v", spec.Process.Args[0], "--"}
	if len(spec.Process.Args) > 1 {
		spec.Process.Args = append(tiniCommand, spec.Process.Args[1:]...)
	} else {
		spec.Process.Args = tiniCommand
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
	logrus.Infof("My annotation are %+v", spec.Annotations)

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
