package sdkcmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
	"github.com/sailpoint-oss/sailpoint-cli/internal/output"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	Limit   int32
	Offset  int32
	Count   bool
	Filters string
	Sorters string
}

func AddListFlags(cmd *cobra.Command, opts *ListOptions) {
	opts.Limit = 250
	cmd.Flags().Int32VarP(&opts.Limit, "limit", "l", opts.Limit, "Maximum number of results to return")
	cmd.Flags().Int32Var(&opts.Offset, "offset", 0, "Offset of the first result to return")
	cmd.Flags().BoolVar(&opts.Count, "count", false, "Request total count metadata from the API")
	cmd.Flags().StringVar(&opts.Filters, "filter", "", "API filter expression")
	cmd.Flags().StringVar(&opts.Sorters, "sort", "", "API sort expression")
}

func WriteTable(cmd *cobra.Command, headers []string, rows [][]string, sortKey string, structuredValue any) error {
	return output.WriteTableOrStructured(cmd.OutOrStdout(), headers, rows, sortKey, structuredValue)
}

func WriteStructured(cmd *cobra.Command, value any) error {
	return output.WriteStructured(cmd.OutOrStdout(), value)
}

func SDKError(resp *http.Response, err error) error {
	if err == nil {
		return nil
	}
	if resp == nil {
		return err
	}

	var body []byte
	if resp.Body != nil {
		body, _ = io.ReadAll(resp.Body)
	}
	return clierror.APIStatus(resp.StatusCode, resp.Status, body)
}

func ReadJSONFile[T any](filePath string) (T, error) {
	var value T
	if filePath == "" {
		return value, clierror.Usage("a JSON payload file is required")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return value, fmt.Errorf("failed to read JSON payload file: %w", err)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("failed to parse JSON payload file: %w", err)
	}
	return value, nil
}
