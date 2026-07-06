package spconfig

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
	sailpoint "github.com/sailpoint-oss/golang-sdk/v3"
	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/v3/sp_config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
)

func PrintJob(x interface{}) {
	switch job := x.(type) {
	case sailpointbetasdk.Spconfigjob:
		fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified())
	case sailpointbetasdk.Spconfigexportjob:
		fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified())
	case sailpointbetasdk.Spconfigexportjobstatus:
		fmt.Printf("Job Type: %s\nJob ID: %s\nDescription: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\nCompleted: %s\n", job.Type, job.JobId, job.GetDescription(), job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified(), job.GetCompleted())
	case sailpointbetasdk.Spconfigimportjobstatus:
		fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\nCompleted: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified(), job.GetCompleted())
	}
}

func DownloadExport(apiClient sailpoint.APIClient, jobId string, fileName string, folderPath string) error {

	for {
		response, _, err := apiClient.SPConfigAPI.GetSpConfigExportStatusV1(context.TODO(), jobId).Execute()
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
				exportData, _, err := apiClient.SPConfigAPI.GetSpConfigExportV1(context.TODO(), jobId).Execute()
				if err != nil {
					return err
				}
				log.Info("Saving export data", "filePath", path.Join(folderPath, fileName+".json"))
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
		response, _, err := apiClient.SPConfigAPI.GetSpConfigImportStatusV1(context.TODO(), jobId).Execute()
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
				importData, _, err := apiClient.SPConfigAPI.GetSpConfigImportV1(context.TODO(), jobId).Execute()
				if err != nil {
					return err
				}
				log.Info("Saving import data", "filePath", path.Join(folderPath, fileName+".json"))
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
