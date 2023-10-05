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
	saveName := GetSanitizedPath(fileName, "json")

	log.Debug("Saving JSON file", "path", folderPath, "file", saveName)

	dataToSave, err := json.MarshalIndent(formattedResponse, "", "  ")
	if err != nil {
		return err
	}

	log.Debug("Formatted Data", "data", string(dataToSave))

	saveErr := WriteFile(folderPath, saveName, dataToSave)
	if saveErr != nil {
		return saveErr
	}

	return nil
}

func WriteFile(folderPath string, filePath string, data []byte) error {

	// Create the folder if it doesn't exist
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, 0777)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path.Join(folderPath, filePath), os.O_CREATE|os.O_RDWR, 0777)
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
func GetSanitizedPath(fileName string, extension string) string {
	return sanitize.PathName(fileName) + "." + extension
}

func WriteTable(writer io.Writer, headers []string, entries [][]string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(headers)
	for _, line := range entries {
		table.Append(line)
	}
	table.Render()
}
