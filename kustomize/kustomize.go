package kustomize

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/ristretto/z"
	"github.com/hashicorp/go-getter"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kustomize/v4/commands/build"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"strconv"
)

type Renderer interface {
	Render() error
}

type KustomizerWrapper struct {
	FSys       filesys.FileSystem
	Kustomizer *krusty.Kustomizer
}

func New() KustomizerWrapper {
	fsys := filesys.MakeFsOnDisk()
	kustomize := krusty.MakeKustomizer(
		build.HonorKustomizeFlags(krusty.MakeDefaultOptions()),
	)
	return KustomizerWrapper{Kustomizer: kustomize, FSys: fsys}
}

func (k KustomizerWrapper) Render(source, path string) ([]unstructured.Unstructured, error) {
	var unstructuredManifests []unstructured.Unstructured

	destination, err := k.getSourceContent(source)
	if err != nil {
		return unstructuredManifests, err
	}
	resMap, err := k.Kustomizer.Run(k.FSys, filepath.Join(destination, path))
	if err != nil {
		return unstructuredManifests, err
	}
	resources, err := json.Marshal(resMap.Resources())
	if err != nil {
		return unstructuredManifests, fmt.Errorf("error marshalling kustomize resources: %w", err)
	}
	err = json.Unmarshal(resources, &unstructuredManifests)
	if err != nil {
		return unstructuredManifests, fmt.Errorf("error converting kustomize resources to unstructured manifests %w", err)
	}
	return unstructuredManifests, nil
}

func (k KustomizerWrapper) getSourceContent(source string) (string, error) {
	destination := filepath.Join(os.TempDir(), "kustomize"+strconv.Itoa(int(z.FastRand())))
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	client := getter.Client{
		Src:  source,
		Dst:  destination,
		Pwd:  pwd,
		Ctx:  context.TODO(),
		Mode: getter.ClientModeAny,
	}
	if err := client.Get(); err != nil {
		return "", err
	}
	return destination, nil
}
