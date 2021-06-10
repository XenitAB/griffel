package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/grafana-tools/sdk"
)

func writeDashboard(format string, path string, board *sdk.Board) error {
	switch format {
	case "json":
		return writeJson(path, board)
	case "operator":
		return errors.New("not implemented")
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
}

func writeJson(path string, board *sdk.Board) error {
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
