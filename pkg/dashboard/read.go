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
)

func readDashboard(kind string, value string) (*sdk.Board, error) {
	var b []byte
	var err error
	switch kind {
	case "GrafanaLabs":
		b, err = getFromGrafanaLabs(value)
	case "URL":
		b, err = getFromUrl(value)
	case "Path":
		b, err = getFromPath(value)
	default:
		return nil, errors.New("unknown dashboard source kind")
	}
	if err != nil {
		return nil, err
	}

	board := &sdk.Board{}
	err = json.Unmarshal(b, board)
	if err != nil {
		return nil, err
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
