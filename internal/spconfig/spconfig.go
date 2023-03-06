package spconfig

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/fatih/color"
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/beta"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
)

func PrintJob(job sailpointbetasdk.SpConfigJob) {
	fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\nCompleted: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified(), job.GetCompleted())
}

func DownloadExport(jobId string, fileName string, folderPath string) error {

	apiClient, err := config.InitAPIClient()
	if err != nil {
		return err
	}

	for {
		response, _, err := apiClient.Beta.SPConfigApi.ExportSpConfigJobStatus(context.TODO(), jobId).Execute()
		if err != nil {
			return err
		}
		if response.Status == "NOT_STARTED" || response.Status == "IN_PROGRESS" {
			color.Yellow("Status: %s. checking again in 5 seconds", response.Status)
			time.Sleep(5 * time.Second)
		} else {
			switch response.Status {
			case "COMPLETE":
				log.Log.Info("Job Complete")
				exportData, _, err := apiClient.Beta.SPConfigApi.ExportSpConfigDownload(context.TODO(), jobId).Execute()
				if err != nil {
					return err
				}
				log.Log.Info("Saving export data", "filePath", path.Join(folderPath, fileName))
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

func DownloadImport(jobId string, fileName string, folderPath string) error {

	apiClient, err := config.InitAPIClient()
	if err != nil {
		return err
	}

	for {
		response, _, err := apiClient.Beta.SPConfigApi.ImportSpConfigJobStatus(context.TODO(), jobId).Execute()
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
				importData, _, err := apiClient.Beta.SPConfigApi.ImportSpConfigDownload(context.TODO(), jobId).Execute()
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
