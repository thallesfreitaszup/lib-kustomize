package cache

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"strings"
)

type Cache interface {
	Get(key interface{}) (interface{}, bool)
	Set(key, value interface{}, cost int64) bool
}
type Wrapper struct {
	client Cache
}

func (w Wrapper) GetManifests(source string) ([]unstructured.Unstructured, error) {
	var unstructuredManifests []unstructured.Unstructured
	var etag string
	repo, owner := w.GetRepoOwner(source)
	apiUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	etagItem, got := w.client.Get(source)
	if !got {
		response, err := doRequest(apiUrl, map[string]string{})
		if err != nil {
			return nil, err
		}
		fmt.Println(response.Header)
		etag = response.Header.Get("ETag")
		println(etag)
		set := w.client.Set(source, etag, 1)
		if !set {
			return nil, err
		}
		return nil, errors.New("first request, not cached yet")
	} else {
		etag = etagItem.(string)
	}
	headers := map[string]string{
		"If-None-Match": etag,
	}
	response, err := doRequest(apiUrl, headers)
	if err != nil {
		return unstructuredManifests, err
	}
	if response.StatusCode == http.StatusNotModified {
		item, got := w.client.Get(etag)
		if !got {
			return nil, fmt.Errorf("failed to get value from key %s", item)
		}
		unstructuredManifests = item.([]unstructured.Unstructured)
		if err != nil {
			return nil, err
		}
		return unstructuredManifests, nil
	}
	return nil, errors.New("resource modified, should download it again")
}

func doRequest(url string, headers map[string]string) (*http.Response, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", url, nil)

	for key, value := range headers {
		request.Header.Add(key, value)
	}
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (w Wrapper) GetRepoOwner(source string) (string, string) {
	arrSource := strings.Split(source, "/")
	return arrSource[len(arrSource)-1], arrSource[len(arrSource)-2]
}

func (w Wrapper) Add(source string, manifests []unstructured.Unstructured) error {
	itemETag, got := w.client.Get(source)
	if !got {
		return errors.New("error getting etag on cache")
	}
	set := w.client.Set(itemETag, manifests, 1)
	if !set {
		return errors.New("failed to set manifests to cache")
	}
	return nil
}

func New(client Cache) Wrapper {
	return Wrapper{
		client: client,
	}
}
