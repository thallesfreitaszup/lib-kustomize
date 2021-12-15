package kustomize

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type Renderer interface {
	Run(fSys filesys.FileSystem, path string) (resmap.ResMap, error)
}

type Getter interface {
	Get() error
}

type KustomizerWrapper struct {
	FSys        filesys.FileSystem
	Renderer    Renderer
	Client      Getter
	Destination string
	Source      string
	Path        string
}

// New Instantiate a new Wrapper of Kustomize that will do the `kustomize build` of the source
func New(kustomizer Renderer, client Getter, destination, source, path string) KustomizerWrapper {
	fsys := filesys.MakeFsOnDisk()

	return KustomizerWrapper{Renderer: kustomizer, FSys: fsys, Client: client, Destination: destination, Source: source, Path: path}
}

// Render downloads the content of the source url and calls the kustomizer run to do the build of
// manifests stored on source
func (k KustomizerWrapper) Render() ([]unstructured.Unstructured, error) {
	var unstructuredManifests []unstructured.Unstructured

	err := k.getSourceContent()
	if err != nil {
		return unstructuredManifests, err
	}

	resMap, err := k.Renderer.Run(k.FSys, filepath.Join(k.Destination, k.Path))
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

func (k KustomizerWrapper) getSourceContent() error {

	if err := k.Client.Get(); err != nil {
		return err
	}
	return nil
}
