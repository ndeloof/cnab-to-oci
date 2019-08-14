package remotes

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/deislabs/cnab-go/bundle"
	"github.com/docker/cnab-to-oci/converter"
	"github.com/docker/cnab-to-oci/tests"
	"github.com/docker/distribution/reference"
	"github.com/docker/go/canonical/json"
	ocischemav1 "github.com/opencontainers/image-spec/specs-go/v1"
	"gotest.tools/assert"
)

const (
	expectedBundleManifest = `{
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType":"application/vnd.oci.image.manifest.v1+json",
      "digest":"sha256:e3eb67b9ca812e0d594f7dbf6e9071b4dc9d06ca338a3e6f851f22f06cf16860",
      "size":189,
      "annotations":{
        "io.cnab.manifest.type":"config"
      }
    },
    {
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "digest": "sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0343",
      "size": 506,
      "annotations": {
        "io.cnab.manifest.type": "invocation"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0342",
      "size": 507,
      "annotations": {
        "io.cnab.component.name": "another-image",
        "io.cnab.manifest.type": "component"
      }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:d59a1aa7866258751a261bae525a1842c7ff0662d4f34a355d5f36826abc0341",
      "size": 507,
      "annotations": {
        "io.cnab.component.name": "image-1",
        "io.cnab.manifest.type": "component"
      }
    }
  ],
  "annotations": {
    "io.cnab.keywords": "[\"keyword1\",\"keyword2\"]",
    "io.cnab.runtime_version": "v1.0.0-WD",
    "org.opencontainers.artifactType": "application/vnd.cnab.manifest.v1",
    "org.opencontainers.image.authors": "[{\"name\":\"docker\",\"email\":\"docker@docker.com\",\"url\":\"docker.com\"}]",
    "org.opencontainers.image.description": "description",
    "org.opencontainers.image.title": "my-app",
    "org.opencontainers.image.version": "0.1.0"
  }
}`
)

func TestPush(t *testing.T) {
	pusher := &mockPusher{}
	resolver := &mockResolver{pusher: pusher}
	b := tests.MakeTestBundle()
	expectedBundleConfig, err := json.MarshalCanonical(b)
	assert.NilError(t, err)
	ref, err := reference.ParseNamed("my.registry/namespace/my-app:my-tag")
	assert.NilError(t, err)

	// push the bundle
	_, err = Push(context.Background(), b, tests.MakeRelocationMap(), ref, resolver, true)
	assert.NilError(t, err)
	assert.Equal(t, len(resolver.pushedReferences), 3)
	assert.Equal(t, len(pusher.pushedDescriptors), 3)
	assert.Equal(t, len(pusher.buffers), 3)

	// check pushed config
	assert.Equal(t, "my.registry/namespace/my-app", resolver.pushedReferences[0])
	assert.Equal(t, converter.CNABConfigMediaType, pusher.pushedDescriptors[0].MediaType)
	assert.Equal(t, oneLiner(string(expectedBundleConfig)), pusher.buffers[0].String())

	// check pushed config manifest
	assert.Equal(t, "my.registry/namespace/my-app", resolver.pushedReferences[1])
	assert.Equal(t, ocischemav1.MediaTypeImageManifest, pusher.pushedDescriptors[1].MediaType)

	// check pushed bundle manifest index
	assert.Equal(t, "my.registry/namespace/my-app:my-tag", resolver.pushedReferences[2])
	assert.Equal(t, ocischemav1.MediaTypeImageIndex, pusher.pushedDescriptors[2].MediaType)
	assert.Equal(t, oneLiner(expectedBundleManifest), pusher.buffers[2].String())
}

func oneLiner(s string) string {
	return strings.Replace(strings.Replace(s, " ", "", -1), "\n", "", -1)
}

func ExamplePush() {
	resolver := createExampleResolver()
	b := createExampleBundle()
	ref, err := reference.ParseNamed("my.registry/namespace/my-app:my-tag")
	if err != nil {
		panic(err)
	}

	// Push the bundle here
	descriptor, err := Push(context.Background(), b, tests.MakeRelocationMap(), ref, resolver, true)
	if err != nil {
		panic(err)
	}

	bytes, err := json.MarshalIndent(descriptor, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s", string(bytes))

	// Output:
	// {
	//   "mediaType": "application/vnd.oci.image.index.v1+json",
	//   "digest": "sha256:4b6f2d8802ec66b8738ba6d8a5e446433476d9f7d174b3d26dcf296ed05cb8cf",
	//   "size": 1363
	// }
}

func createExampleBundle() *bundle.Bundle {
	return tests.MakeTestBundle()
}
