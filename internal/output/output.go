package output

import (
	"bufio"
	"encoding/json"
	"os"
	"path"
	"path/filepath"

	"github.com/gocarina/gocsv"
)

func SaveJSONFile[T any](formattedResponse T, fileName string, folderPath string) error {

	savePath := path.Join(folderPath, fileName)

	dataToSave, err := json.MarshalIndent(formattedResponse, "", "  ")
	if err != nil {
		return err
	}

	// Make sure the output dir exists first
	err = os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(savePath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}

	fileWriter := bufio.NewWriter(file)

	_, err = fileWriter.Write(dataToSave)
	if err != nil {
		return err
	}

	return nil
}

func SaveCSVFile[T any](formattedResponse T, fileName string, folderPath string) error {
	savePath := filepath.Join(folderPath, fileName)

	// Make sure the output dir exists first
	err := os.MkdirAll(folderPath, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(savePath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	err = gocsv.MarshalFile(formattedResponse, file)
	if err != nil {
		return err
	}

	return nil
}
