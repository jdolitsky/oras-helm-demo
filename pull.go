package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containerd/containerd/remotes/docker"
	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/oras"
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

	for _, layer := range layers {
		//digest := layer.Digest.Hex()
		fmt.Println(layer)
	}

	fmt.Println("Success!")
}

// TODO: remove once WARN lines removed from oras/containerd
func init() {
	logrus.SetLevel(logrus.ErrorLevel)
}
