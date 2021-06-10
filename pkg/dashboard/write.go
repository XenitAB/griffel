package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana-tools/sdk"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/spf13/afero"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
)

func writeDashboard(fs afero.Fs, path string, format string, name string, board *sdk.Board) error {
	if path == "" {
		return errors.New("path cannot be empty")
	}

	switch format {
	case "json":
		return writeJson(fs, path, board)
	case "operator":
		return writeOperator(fs, path, name, board)
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

func writeOperator(fs afero.Fs, path string, name string, board *sdk.Board) error {
	b, err := marshalBoard(board)
	if err != nil {
		return err
	}
	dashboard := grafanav1alpha1.GrafanaDashboard{
		TypeMeta: metav1.TypeMeta{
			APIVersion: grafanav1alpha1.SchemeGroupVersion.String(),
			Kind:       "GrafanaDashboard",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: grafanav1alpha1.GrafanaDashboardSpec{
			Json: string(b),
		},
	}
	grafanav1alpha1.AddToScheme(scheme.Scheme)
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
