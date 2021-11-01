package util

import (
	"github.com/grafana-tools/sdk"
)

func StringPointer(input string) *string {
	return &input
}

func IntPointer(input int) *int {
	return &input
}

func Int64Pointer(input int64) *int64 {
	return &input
}

func PanelPointerSlice(panels []sdk.Panel) []*sdk.Panel {
	newPanels := []*sdk.Panel{}
	for i := range panels {
		newPanels = append(newPanels, &panels[i])
	}
	return newPanels
}

func PanelSlice(panels []*sdk.Panel) []sdk.Panel {
	newPanels := []sdk.Panel{}
	for _, p := range panels {
		newPanels = append(newPanels, *p)
	}
	return newPanels
}
