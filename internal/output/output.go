package output

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path"
	"sort"

	"github.com/charmbracelet/log"
	"github.com/mrz1836/go-sanitize"
	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/util"
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

func WriteTable(writer io.Writer, headers []string, entries [][]string, sortKey string) {
	table := tablewriter.NewWriter(writer)
	table.SetHeader(headers)

	// Find the index of the header that matches the sortKey
	sortIndex := -1
	for i, header := range headers {
		if header == sortKey {
			sortIndex = i
			break
		}
	}

	// If a valid sortKey is provided, sort the entries by that column
	if sortIndex != -1 {
		sort.Slice(entries, func(i, j int) bool {
			return entries[i][sortIndex] < entries[j][sortIndex]
		})
	}

	// Append sorted (or unsorted) entries to the table
	for _, line := range entries {
		table.Append(line)
	}

	table.Render()
}

func WriteJson(jsonString []byte) string {
	var err error
	var data map[string]interface{}

	err = json.Unmarshal(jsonString, &data)
	if err != nil {
		log.Error(err)
	}

	return util.RenderMarkdown("```json\n" + util.PrettyPrint(data) + "\n```")
}
