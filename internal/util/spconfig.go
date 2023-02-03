package util

import (
	"fmt"

	sailpointbetasdk "github.com/sailpoint-oss/golang-sdk/sdk-output/beta"
)

func PrintJob(job sailpointbetasdk.SpConfigJob) {
	fmt.Printf("Job Type: %s\nJob ID: %s\nStatus: %s\nExpired: %s\nCreated: %s\nModified: %s\nCompleted: %s\n", job.Type, job.JobId, job.Status, job.GetExpiration(), job.GetCreated(), job.GetModified(), job.GetCompleted())
}
