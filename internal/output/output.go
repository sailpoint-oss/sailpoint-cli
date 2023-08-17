package output

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path"

	"github.com/charmbracelet/log"
	"github.com/mrz1836/go-sanitize"
	"github.com/olekukonko/tablewriter"
)

func SaveJSONFile[T any](formattedResponse T, fileName string, folderPath string) error {
	savePath := GetSanitizedPath(folderPath, fileName, "json")

	log.Debug("Saving JSON file", "file path", savePath)

	dataToSave, err := json.MarshalIndent(formattedResponse, "", "  ")
	if err != nil {
		return err
	}

	log.Debug("Formatted Data", "data", string(dataToSave))

	return WriteFile(savePath, dataToSave)
}

func WriteFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}

	fileWriter := bufio.NewWriter(file)

	_, err = fileWriter.Write(data)
	if err != nil {
		return err
	}

	err = fileWriter.Flush()
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

// GetSanitizedPath makes sure the provided path works on all OS
func GetSanitizedPath(filePath string, fileName string, extension string) string {
	return path.Join(filePath, sanitize.PathName(fileName)+"."+extension)
}

type TableEntry struct {
	Name string
	ID   string
}

func WriteTable(writer io.Writer, entries []TableEntry) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader([]string{"Name", "ID"})
	for _, line := range entries {
		table.Append([]string{line.Name, line.ID})
	}
	table.Render()
}
