package output

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mrz1836/go-sanitize"
	"github.com/olekukonko/tablewriter"
	"github.com/sailpoint-oss/sailpoint-cli/internal/redact"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func SaveJSONFile[T any](formattedResponse T, fileName string, folderPath string) error {
	saveName := GetSanitizedPath(fileName, "json")

	log.Debug("Saving JSON file", "path", folderPath, "file", saveName)

	dataToSave, err := json.MarshalIndent(formattedResponse, "", "  ")
	if err != nil {
		return err
	}

	log.Debug("Formatted data", "bytes", len(dataToSave), "redactedPreview", string(redact.JSONBytes(dataToSave)))

	saveErr := WriteFile(folderPath, saveName, dataToSave)
	if saveErr != nil {
		return saveErr
	}

	return nil
}

func WriteFile(folderPath string, filePath string, data []byte) error {

	// Create the folder if it doesn't exist
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, 0755)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(path.Join(folderPath, filePath), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
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
	if IsMachineReadable() {
		_ = WriteStructured(writer, tableRecords(headers, entries))
		return
	}

	table := tablewriter.NewWriter(writer)
	// Convert []string to []any for the Header method
	headerAny := make([]any, len(headers))
	for i, h := range headers {
		headerAny[i] = h
	}
	table.Header(headerAny...)

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

const (
	FormatTable = "table"
	FormatJSON  = "json"
	FormatYAML  = "yaml"
)

func CurrentFormat() string {
	if viper.GetBool("json") {
		return FormatJSON
	}

	format := strings.ToLower(strings.TrimSpace(viper.GetString("output")))
	switch format {
	case "", FormatTable:
		return FormatTable
	case FormatJSON, FormatYAML:
		return format
	default:
		return FormatTable
	}
}

func IsMachineReadable() bool {
	format := CurrentFormat()
	return format == FormatJSON || format == FormatYAML
}

func WriteStructured(writer io.Writer, value any) error {
	switch CurrentFormat() {
	case FormatJSON:
		encoder := json.NewEncoder(writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(value)
	case FormatYAML:
		data, err := yaml.Marshal(value)
		if err != nil {
			return err
		}
		_, err = writer.Write(data)
		return err
	default:
		_, err := fmt.Fprintln(writer, Pretty(value))
		return err
	}
}

func WriteTableOrStructured(writer io.Writer, headers []string, entries [][]string, sortKey string, structuredValue any) error {
	if IsMachineReadable() {
		return WriteStructured(writer, structuredValue)
	}

	WriteTable(writer, headers, entries, sortKey)
	return nil
}

func Pretty(value any) string {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(data)
}

func tableRecords(headers []string, entries [][]string) []map[string]string {
	records := make([]map[string]string, 0, len(entries))
	for _, entry := range entries {
		record := make(map[string]string, len(headers))
		for i, header := range headers {
			if i >= len(entry) {
				continue
			}
			record[normalizeHeader(header)] = entry[i]
		}
		records = append(records, record)
	}
	return records
}

func normalizeHeader(header string) string {
	normalized := strings.TrimSpace(strings.ToLower(header))
	if normalized == "" {
		return "marker"
	}
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")
	return normalized
}
