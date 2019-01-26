package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/helm/pkg/chartutil"
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

	// Read command line args
	chartDir := os.Args[1]
	remoteRef := os.Args[2]
	fmt.Printf("Attempting to push %s to %s...\n", chartDir, remoteRef)

	// Load chart directory
	chart, err := chartutil.LoadDir(chartDir)
	check(err)
	fmt.Printf("name: %s\nversion: %s\n", chart.Metadata.Name, chart.Metadata.Version)

	// Create meta layer
	metaMediaType := "application/vnd.cncf.helm.chart.meta.v1+json"
	metaRaw := []byte("hello")
	metaLayer := memoryStore.Add("", metaMediaType, metaRaw)

	// Create content layer
	contentMediaType := "application/vnd.cncf.helm.chart.content.v1+tar"
	contentRaw := []byte("hello")
	contentLayer := memoryStore.Add("", contentMediaType, contentRaw)
	contentLayer.Annotations = map[string]string{
		"name":    chart.Metadata.Name,
		"version": chart.Metadata.Version,
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
