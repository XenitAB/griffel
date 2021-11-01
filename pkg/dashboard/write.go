package dashboard

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	grafanav1alpha1 "github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-tools/sdk"
	"github.com/spf13/afero"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/xenitab/griffel/pkg/config"
)

func writeDashboard(fs afero.Fs, directory string, format config.OutputFormat, name string, board *sdk.Board) error {
	path := name
	if directory != "" {
		path = filepath.Join(directory, name)
	}
	switch format {
	case config.OutputFormatJson:
		return writeJson(fs, fmt.Sprintf("%s.json", path), board)
	case config.OutputFormatKubernetes:
		return writeKubernetes(fs, fmt.Sprintf("%s.yaml", path), name, board)
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func writeJson(fs afero.Fs, path string, board *sdk.Board) error {
	b, err := marshalBoard(board)
	if err != nil {
		return err
	}
	err = afero.WriteFile(fs, path, b, 0644)
	if err != nil {
		return err
	}
	return nil
}

func writeKubernetes(fs afero.Fs, path string, name string, board *sdk.Board) error {
	b, err := marshalBoard(board)
	if err != nil {
		return err
	}
	dashboard := grafanav1alpha1.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{
			APIVersion: grafanav1alpha1.GroupVersion.String(),
			Kind:       "GrafanaDashboard",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: string(b),
		},
	}
	err = grafanav1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return fmt.Errorf("could not add grafana operator to scheme: %w", err)
	}
	opt := k8sjson.SerializerOptions{
		Yaml: true,
	}
	s := k8sjson.NewSerializerWithOptions(k8sjson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, opt)
	file, err := fs.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = s.Encode(&dashboard, file)
	if err != nil {
		return err
	}
	return nil
}

func marshalBoard(board *sdk.Board) ([]byte, error) {
	prefix := ""
	indent := "  "
	b, err := json.MarshalIndent(board, prefix, indent)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}
