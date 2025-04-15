// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/jsonpath"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var headerFlags []string
	var prettyPrint bool
	var queryParams map[string]string
	var jsonPath string

	cmd := &cobra.Command{
		Use:     "delete [endpoint]",
		Short:   "Make a DELETE request to a SailPoint API endpoint",
		Long:    "\nMake a DELETE request to a SailPoint API endpoint\n\n",
		Example: "sail api delete /beta/accounts/123",
		Aliases: []string{"d"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.InitConfig()
			if err != nil {
				return err
			}

			// Get the SailPoint client configuration
			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			// Create a client
			spClient := client.NewSpClient(cfg)

			endpoint := args[0]
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}

			// Prepare headers
			headers := make(map[string]string)
			// Always add Accept header for JSON
			headers["Accept"] = "application/json"
			// Add any additional headers
			for _, header := range headerFlags {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid header format (use Key: Value): %s", header)
				}
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}

			ctx := context.Background()
			log.Info("Making DELETE request", "endpoint", endpoint)

			// Make the request
			resp, err := spClient.Delete(ctx, endpoint, queryParams, headers)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			// Read response body
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// If JSONPath is specified, evaluate it
			if jsonPath != "" {
				result, err := jsonpath.EvaluateJSONPathToString(responseBody, jsonPath)
				if err != nil {
					return fmt.Errorf("failed to evaluate JSONPath: %w", err)
				}
				responseBody = []byte(result)
			}

			// Check if response is JSON and pretty print if requested
			if prettyPrint {
				var jsonData interface{}
				if err := json.Unmarshal(responseBody, &jsonData); err == nil {
					prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
					if err == nil {
						responseBody = prettyJSON
					}
				}
			}

			if jsonPath != "" {
				cmd.Print(string(responseBody))
			} else {
				cmd.Println(string(responseBody))
				fmt.Printf("Status: %s\n", resp.Status)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&headerFlags, "header", "H", []string{}, "HTTP headers (can be used multiple times, format: 'Key: Value')")
	cmd.Flags().BoolVarP(&prettyPrint, "pretty", "p", false, "Pretty print JSON response")
	cmd.Flags().StringToStringVarP(&queryParams, "query", "q", nil, "Query parameters (can be used multiple times, format: 'key=value')")
	cmd.Flags().StringVarP(&jsonPath, "jsonpath", "j", "", "JSONPath expression to evaluate on the response")

	return cmd
}
