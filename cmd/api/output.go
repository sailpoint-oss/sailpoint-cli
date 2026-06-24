package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/spf13/cobra"
)

func ensureSuccess(resp *http.Response, body []byte) error {
	if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	return clierror.APIStatus(resp.StatusCode, resp.Status, body)
}

func writeResponse(cmd *cobra.Command, body []byte, status string, jsonPath string) error {
	if jsonPath != "" {
		_, err := fmt.Fprint(cmd.OutOrStdout(), string(body))
		return err
	}

	if output.IsMachineReadable() {
		var value any
		if err := json.Unmarshal(body, &value); err == nil {
			return output.WriteStructured(cmd.OutOrStdout(), value)
		}
	}

	_, err := fmt.Fprintln(cmd.OutOrStdout(), string(body))
	if err != nil {
		return err
	}

	return writeStatus(cmd, status)
}

func writeResponseFile(cmd *cobra.Command, filename string, body []byte, status string) error {
	if err := writeToFile(filename, body); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "Response saved to %s\n", filename); err != nil {
		return err
	}
	return writeStatus(cmd, status)
}

func writeStatus(cmd *cobra.Command, status string) error {
	if output.IsMachineReadable() || status == "" {
		return nil
	}
	_, err := fmt.Fprintf(cmd.ErrOrStderr(), "Status: %s\n", status)
	return err
}

func writeToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0600)
}
