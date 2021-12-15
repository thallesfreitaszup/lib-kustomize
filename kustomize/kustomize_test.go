package kustomize_test

import (
	"errors"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	"lib-kustomize/kustomize"
	"lib-kustomize/kustomize/mocks"
	"path/filepath"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var _ = Describe("Kustomize", func() {
	var source string
	var destination string
	var path string
	var getter *mocks.Getter
	var renderer *mocks.Renderer
	BeforeEach(func() {

		source = "example.com/test"
		destination = "/destination"
		path = "path"
		getter = new(mocks.Getter)
		renderer = new(mocks.Renderer)
	})
	Context("when fails to download repository content", func() {
		It("should return error", func() {

			error := errors.New("failed to download resource")

			getter.On("Get").Return(error)

			k := kustomize.New(renderer, getter, destination, source, path)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, error)
			assert.Equal(GinkgoT(), len(manifests), 0)
		})
	})

	Context("when fails to render manifests", func() {
		It("should return error", func() {

			error := errors.New("failed to render resource")

			getter.On("Get").Return(nil)
			renderer.On("Run", filesys.MakeFsOnDisk(), filepath.Join(destination, path)).Return(resmap.New(), error)

			k := kustomize.New(renderer, getter, destination, source, path)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, error)
			assert.Equal(GinkgoT(), len(manifests), 0)
		})
	})

	Context("when successfully render manifests", func() {
		It("should return the correct unstructured manifests", func() {

			getter.On("Get").Return(nil)
			renderer.On("Run", filesys.MakeFsOnDisk(), filepath.Join(destination, path)).Return(getManifestsResponseMap(), nil)

			k := kustomize.New(renderer, getter, destination, source, path)
			manifests, renderError := k.Render()
			assert.Equal(GinkgoT(), renderError, nil)
			assert.Equal(GinkgoT(), len(manifests), 2)
			assert.Equal(GinkgoT(), manifests[0].GetName(), "deploy1")
			assert.Equal(GinkgoT(), manifests[0].GetKind(), "Deployment")
			assert.Equal(GinkgoT(), manifests[0].GetAPIVersion(), "apps/v1")
			assert.Equal(GinkgoT(), manifests[1].GetAPIVersion(), "apps/v1")
			assert.Equal(GinkgoT(), manifests[1].GetKind(), "Deployment")
			assert.Equal(GinkgoT(), manifests[1].GetName(), "deploy2")
		})
	})

})

func getManifestsResponseMap() resmap.ResMap {
	var depProvider = provider.NewDefaultDepProvider()
	var rf = depProvider.GetResourceFactory()
	resourceStr := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy1
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
---
`
	fSys := filesys.MakeFsInMemory()
	err := fSys.WriteFile("deployment.yaml", []byte(resourceStr))
	assert.NoError(GinkgoT(), err)
	ldr, err := loader.NewLoader(
		loader.RestrictionRootOnly, filesys.Separator, fSys)
	assert.NoError(GinkgoT(), err)
	var rmF = resmap.NewFactory(rf)
	resmap, err := rmF.FromFile(ldr, "deployment.yaml")
	assert.NoError(GinkgoT(), err)
	return resmap
}
