package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/grafana-tools/sdk"
	"github.com/xenitab/griffel/pkg/config"
)

func readDashboard(kind config.SourceKind, value string) (*sdk.Board, error) {
	var b []byte
	var err error
	switch kind {
	case config.SourceKindGrafanaLabs:
		b, err = getFromGrafanaLabs(value)
	case config.SourceKindURL:
		b, err = getFromUrl(value)
	case config.SourceKindPath:
		b, err = getFromPath(value)
	default:
		return nil, errors.New("unknown dashboard source kind")
	}
	if err != nil {
		return nil, fmt.Errorf("could not get source: %w", err)
	}

	board := &sdk.Board{}
	err = json.Unmarshal(b, board)
	if err != nil {
		return nil, fmt.Errorf("could not parse dashboard: %w", err)
	}
	return board, nil
}

func getFromGrafanaLabs(id string) ([]byte, error) {
	url := fmt.Sprintf("https://grafana.com/api/dashboards/%s/revisions/1/download", id)
	return getFromUrl(url)
}

func getFromUrl(url string) ([]byte, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}

func getFromPath(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return []byte{}, err
	}
	return b, nil
}
