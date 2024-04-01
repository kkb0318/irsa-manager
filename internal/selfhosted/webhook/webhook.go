package webhook

import (
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

type WebhookManifests struct {
	Manifests []*unstructured.Unstructured
}

func Read() ([]*unstructured.Unstructured, error) {
	files := []string{
		"auth.yaml",
		"deployment.yaml",
		"mutatingwebhook-ca-bundle.yaml",
		"mutatingwebhook.yaml",
		"service.yaml",
	}
	objs := []*unstructured.Unstructured{}
	for _, f := range files {
		tmp, err := ReadFile(f)
		if err != nil {
			return nil, err
		}
		objs = append(objs, tmp...)
	}
	return objs, nil

}

func ReadFile(filePath string) ([]*unstructured.Unstructured, error) {
	yamlContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// YAMLデコーダの準備
	decUnstructured := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	// マニフェストを分割
	documents := strings.Split(string(yamlContent), "---")
	objs := make([]*unstructured.Unstructured, len(documents))
	for i, doc := range documents {
		if strings.TrimSpace(doc) == "" {
			continue
		}
		// 各マニフェストを解析
		obj := &unstructured.Unstructured{}
		_, _, err := decUnstructured.Decode([]byte(doc), nil, obj)
		if err != nil {
			return nil, err
		}
		objs[i] = obj
	}
	return objs, nil
}
