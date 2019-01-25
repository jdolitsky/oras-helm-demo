# ORAS Helm Demo

Example of using ORAS as a Go library to push a Helm chart to a registry.

## Setup

First, clone this repo:

```
git clone https://github.com/jdolitsky/oras-helm-demo.git
cd oras-helm-demo/
```

Next, gather required dependencies (requires Go 1.11+):

```
go mod vendor
```

Finally, start a Distribution registry server at `localhost:5000` with the following command:

```
docker run -dp 5000:5000 --restart=always --name registry registry:2
```

The will run in the background. Use `docker logs -f registry` to see the logs and `docker rm -f registry` to stop.

## Run

Run `main.go` with some arguments:

```
go run main.go mychart/ localhost:5000/mychart:latest
```

The first arg, `mychart/`, refers to a Helm chart directory path.

The second arg, `localhost:5000/mychart:latest` is a reference
to a remote registry address.

This will push the chart as 2 separate layers with the following media types:
1. `application/vnd.cncf.helm.chart.config.v1+json` (metadata)
2. `application/vnd.cncf.helm.chart.content.v1+tar` (package content)

By separating `Chart.yaml` (a.k.a the metadata) from the rest of the Helm chart, we prevent storing the same content in the registry twice for different names.

## Manifest

You can use `curl` and `jq` to inspect the manifest of a pushed Helm chart:

```
curl -s -H 'Accept: application/vnd.oci.image.manifest.v1+json' \
    http://localhost:5000/v2/mychart/manifests/latest | jq
```

Output:
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
      "mediaType": "application/vnd.cncf.helm.chart.content.v1.tar",
      "digest": "sha256:929d030362719cad6877ca4e8e6d3a60345cdfa97ace3e119ca0f674c2750df4",
      "size": 18344,
      "annotations": {
        "org.opencontainers.image.title": "929d030362719cad6877ca4e8e6d3a60345cdfa97ace3e119ca0f674c2750df4"
      }
    }
  ]
}

```
