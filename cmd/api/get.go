// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/config"
	"github.com/sailpoint-oss/sailpoint-cli/internal/jsonpath"
	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	var headerFlags []string
	var queryParams []string
	var outputFile string
	var prettyPrint bool
	var jsonPath string

	cmd := &cobra.Command{
		Use:     "get [endpoint]",
		Short:   "Make a GET request to a SailPoint API endpoint",
		Long:    "\nMake a GET request to a SailPoint API endpoint\n\n",
		Example: "sail api get /beta/accounts",
		Aliases: []string{"g"},
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

			// Add query parameters if any
			if len(queryParams) > 0 {
				parsedURL, err := url.Parse(endpoint)
				if err != nil {
					return fmt.Errorf("invalid endpoint URL: %w", err)
				}

				query := parsedURL.Query()
				for _, param := range queryParams {
					parts := strings.SplitN(param, "=", 2)
					if len(parts) != 2 {
						return fmt.Errorf("invalid query parameter format (use key=value): %s", param)
					}
					query.Add(parts[0], parts[1])
				}

				parsedURL.RawQuery = query.Encode()
				endpoint = parsedURL.String()
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
			log.Info("Making GET request", "endpoint", endpoint)

			// Make the request using the SailPoint client
			resp, err := spClient.Get(ctx, endpoint)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			// If JSONPath is specified, evaluate it
			if jsonPath != "" {
				result, err := jsonpath.EvaluateJSONPath(body, jsonPath)
				if err != nil {
					return fmt.Errorf("failed to evaluate JSONPath: %w", err)
				}
				body = result
			}

			// Check if response is JSON and pretty print if requested
			if prettyPrint {
				var jsonData interface{}
				if err := json.Unmarshal(body, &jsonData); err == nil {
					prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
					if err == nil {
						body = prettyJSON
					}
				}
			}

			// Output to file or stdout
			if outputFile != "" {
				if err := writeToFile(outputFile, body); err != nil {
					return fmt.Errorf("failed to write to file: %w", err)
				}
				fmt.Printf("Response saved to %s\n", outputFile)
			} else {
				cmd.Println(string(body))
			}

			fmt.Printf("Status: %s\n", resp.Status)
			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&headerFlags, "header", "H", []string{}, "HTTP headers (can be used multiple times, format: 'Key: Value')")
	cmd.Flags().StringArrayVarP(&queryParams, "query", "q", []string{}, "Query parameters (can be used multiple times, format: 'key=value')")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file to save the response (if not specified, prints to stdout)")
	cmd.Flags().BoolVarP(&prettyPrint, "pretty", "p", false, "Pretty print JSON response")
	cmd.Flags().StringVarP(&jsonPath, "jsonpath", "j", "", "JSONPath expression to evaluate on the response")

	return cmd
}

// writeToFile writes data to a file
func writeToFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}
