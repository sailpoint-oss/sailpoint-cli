package spconfig

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v2"
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/v2/api_beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
)

func PrintJob(x interface{}) {
	switch job := x.(type) {
	case sailpointbetasdk.SpConfigJob:
		fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified())
	case sailpointbetasdk.SpConfigExportJob:
		fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified())
	case sailpointbetasdk.SpConfigExportJobStatus:
		fmt.Printf("Job Type: %s\nJob ID: %s\nDescription: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\nCompleted: %s\n", job.Type, job.JobId, job.GetDescription(), job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified(), job.GetCompleted())
	case sailpointbetasdk.SpConfigImportJobStatus:
		fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\nCompleted: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified(), job.GetCompleted())
	}
}

func DownloadExport(apiClient sailpoint.APIClient, jobId string, fileName string, folderPath string) error {

	for {
		response, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExportStatus(context.TODO(), jobId).Execute()
		if err != nil {
			return err
		}
		if response.Status == "NOT_STARTED" || response.Status == "IN_PROGRESS" {
			color.Yellow("Status: %s. checking again in 5 seconds", response.Status)
			time.Sleep(5 * time.Second)
		} else {
			switch response.Status {
			case "COMPLETE":
				log.Info("Job Complete")
				exportData, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigExport(context.TODO(), jobId).Execute()
				if err != nil {
					return err
				}
				log.Info("Saving export data", "filePath", path.Join(folderPath, fileName))
				err = output.SaveJSONFile(exportData, fileName, folderPath)
				if err != nil {
					return err
				}
			case "CANCELLED":
				return fmt.Errorf("export task cancelled")
			case "FAILED":
				return fmt.Errorf("export task failed")
			}
			break
		}
	}

	return nil
}

func DownloadImport(apiClient sailpoint.APIClient, jobId string, fileName string, folderPath string) error {

	for {
		response, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigImportStatus(context.TODO(), jobId).Execute()
		if err != nil {
			return err
		}
		if response.Status == "NOT_STARTED" || response.Status == "IN_PROGRESS" {
			color.Yellow("Status: %s. checking again in 5 seconds", response.Status)
			time.Sleep(5 * time.Second)
		} else {
			switch response.Status {
			case "COMPLETE":
				color.Green("Downloading Import Data")
				importData, _, err := apiClient.Beta.SPConfigAPI.GetSpConfigImport(context.TODO(), jobId).Execute()
				if err != nil {
					return err
				}
				err = output.SaveJSONFile(importData, fileName, folderPath)
				if err != nil {
					return err
				}
			case "CANCELLED":
				return fmt.Errorf("import task cancelled")
			case "FAILED":
				return fmt.Errorf("import task failed")
			}
			break
		}
	}

	return nil
}
