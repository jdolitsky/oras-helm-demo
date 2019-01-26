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

	// Read command line args
	remoteRef := os.Args[1]
	fmt.Printf("Attempting to pull %s into ./output...\n", remoteRef)

	// Pull layers from remote, filtering on only media types we care about
	allowedMediaTypes := []string{helmChartMetaMediaType, helmChartContentMediaType}
	layers, err := oras.Pull(ctx, resolver, remoteRef, memoryStore, allowedMediaTypes...)
	check(err)

	// Make sure we have layers we need to construct a Helm Chart
	var metaLayer, contentLayer ocispec.Descriptor
	var metaLayerFound, contentLayerFound bool
	for _, layer := range layers {
		if layer.MediaType == helmChartMetaMediaType {
			metaLayer = layer
			metaLayerFound = true
		} else if layer.MediaType == helmChartContentMediaType {
			contentLayer = layer
			contentLayerFound = true
		}
	}
	if !metaLayerFound || !contentLayerFound {
		panic(fmt.Sprintf("%s does not have the necessary layers", remoteRef))
	}

	// Extract chart name and version from annotations
	name, hasName := contentLayer.Annotations["chart.name"]
	version, hasVersion := contentLayer.Annotations["chart.version"]
	if !hasName || !hasVersion {
		panic(fmt.Sprintf("%s does not chart name and version saved on annoations", remoteRef))
	}

	fmt.Println(metaLayer, name, version)

	fmt.Println("Success!")
}

// TODO: remove once WARN lines removed from oras/containerd
func init() {
	logrus.SetLevel(logrus.ErrorLevel)
}
