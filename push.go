package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
	chartmeta "k8s.io/helm/pkg/proto/hapi/chart"
)

const (
	helmChartMetaMediaType    = "application/vnd.cncf.helm.chart.meta.v1+json"
	helmChartContentMediaType = "application/vnd.cncf.helm.chart.content.v1+tar"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	ctx := context.Background()
	memoryStore := content.NewMemoryStore()
	resolver := docker.NewResolver(docker.ResolverOptions{})
	cwd, err := os.Getwd()
	check(err)

	// Read command line args
	chartDir := os.Args[1]
	remoteRef := os.Args[2]
	fmt.Printf("Attempting to push %s to %s...\n", chartDir, remoteRef)

	// Load chart directory, extract/separate the name and version from other metadata
	chart, err := chartutil.LoadDir(chartDir)
	check(err)
	name := chart.Metadata.Name
	version := chart.Metadata.Version
	chart.Metadata.Name = ""
	chart.Metadata.Version = ""
	metaJsonRaw, err := json.Marshal(chart.Metadata)
	check(err)
	fmt.Printf("name: %s\nversion: %s\nmetadata: %s\n", name, version, metaJsonRaw)

	// Create meta layer
	metaLayer := memoryStore.Add("", helmChartMetaMediaType, metaJsonRaw)

	// Create content layer, which is the chart minus the metadata (Chart.yaml)
	chart.Metadata = &chartmeta.Metadata{Name: "-", Version: "-"}
	tempDirPrefix := filepath.Join(cwd, ".push")
	os.MkdirAll(tempDirPrefix, 0755)
	tempDir, err := ioutil.TempDir(tempDirPrefix, "oras-helm-demo")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	tmpFile, err := chartutil.Save(chart, tempDir)
	check(err)
	contentRaw, err := ioutil.ReadFile(tmpFile)
	check(err)
	contentLayer := memoryStore.Add("", helmChartContentMediaType, contentRaw)

	// Add the chart name and version as annotations on the content layer
	contentLayer.Annotations = map[string]string{
		"name":    name,
		"version": version,
	}

	// Push to remote
	layers := []ocispec.Descriptor{metaLayer, contentLayer}
	err = oras.Push(ctx, resolver, remoteRef, memoryStore, layers)
	check(err)

	fmt.Println("Success!")
}

// TODO: remove once WARN lines removed from oras/containerd
func init() {
	logrus.SetLevel(logrus.ErrorLevel)
}
