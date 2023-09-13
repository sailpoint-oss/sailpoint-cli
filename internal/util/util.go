package util

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/log"
	"github.com/mrz1836/go-sanitize"
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

func PrettyPrint(v interface{}) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Error("Error marshalling interface", "error", err)
	}
	return (string(b))
}

func SanitizeFileName(fileName string) string {
	return sanitize.PathName(fileName)
}

func RenderMarkdown(markdown string) string {
	out, err := renderer.Render(markdown)
	if err != nil {
		panic(err)
	}

	return out
}

type Help struct {
	Long    string
	Example string
}

func ParseHelp(help string) Help {
	helpParser, err := regexp.Compile(`==([A-Za-z]+)==([\s\S]*?)====`)
	if err != nil {
		panic(err)
	}

	matches := helpParser.FindAllStringSubmatch(help, -1)

	var helpObj Help
	for _, set := range matches {
		switch strings.ToLower(set[1]) {
		case "long":
			helpObj.Long = RenderMarkdown(set[2])
		case "example":
			helpObj.Example = RenderMarkdown(set[2])
		}
	}

	return helpObj
}
