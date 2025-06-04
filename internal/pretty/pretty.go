package pretty

import (
	"encoding/json"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/log"
)

var renderer *glamour.TermRenderer

func init() {
	var err error
	renderer, err = glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
	)
	if err != nil {
		panic(err)
	}

}

func Print(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Error("Error marshalling interface", "error", err)
	}
	return (string(b))
}
