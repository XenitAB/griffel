package dashboard

import (
	"encoding/json"
	"os"

	"github.com/grafana-tools/sdk"
)

func writeDashboard(path string, format string, board *sdk.Board) error {
	b, err := json.MarshalIndent(board, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, b, 0644)
	if err != nil {
		return err
	}
	return nil
}
