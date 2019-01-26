# oras-helm-demo

Demo of using [`oras`](https://github.com/deislabs/oras) as a Go library to push/pull a Helm chart to/from a registry.

## Setup

Start a Distribution registry server at `localhost:5000` with the following command:

```
docker run -dp 5000:5000 --restart=always --name registry registry:2
```

The will run in the background. Use `docker logs -f registry` to see the logs and `docker rm -f registry` to stop.

## Examples

To run examples below, clone this repo and gather required dependencies (requires Go 1.11+):

```
git clone https://github.com/jdolitsky/oras-helm-demo.git && cd oras-helm-demo/
GO111MODULE=on go mod vendor
```

### Push Helm chart to registry

Souce code for `push.go` can be found [here](./push.go).

Run `push.go` with 2 arguments:

```
go run push.go mychart/ localhost:5000/mychart:latest
```

The first arg, `mychart/`, refers to a Helm chart directory path.

The second arg, `localhost:5000/mychart:latest` is a reference
to a remote registry address.

This will push the chart as 2 separate layers with the following media types:
1. `application/vnd.cncf.helm.chart.config.v1+json` (metadata)
2. `application/vnd.cncf.helm.chart.content.v1+tar` (package content)

By separating `Chart.yaml` (a.k.a the metadata) from the rest of the Helm chart, we prevent storing the same content in the registry twice for different names.

### Pull Helm chart from registry

Souce code for `pull.go` can be found [here](./pull.go).

Run `pull.go` with a single argument:

```
go run pull.go localhost:5000/mychart:latest
```

This will download and convert the stored Helm chart into a usable format, saving it to `./output/<chartname>`.

## The Manifest

You can use `curl` and `jq` to inspect the manifest of a Helm chart stored in a registry:

```
curl -s -H 'Accept: application/vnd.oci.image.manifest.v1+json' \
    http://localhost:5000/v2/mychart/manifests/latest | jq
```

Example manifest:
```
{
  "schemaVersion": 2,
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "digest": "sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a",
    "size": 2
  },
  "layers": [
    {
      "mediaType": "application/vnd.cncf.helm.chart.meta.v1+json",
      "digest": "sha256:c356ec641a696eb5f3320bed9e8ceeb505fcc84b7ee072a85a8098fc362e13b7",
      "size": 210,
      "annotations": {
        "org.opencontainers.image.title": "meta.json"
      }
    },
    {
      "mediaType": "application/vnd.cncf.helm.chart.content.v1+tar",
      "digest": "sha256:e5e50410addbc4d1aa16100e42e4eb99f2bb4b04157de130a528e6e5d8c71774",
      "size": 431,
      "annotations": {
        "chart.name": "mychart",
        "chart.version": "2.7.1",
        "org.opencontainers.image.title": "content.tgz"
      }
    }
  ]
}
```
