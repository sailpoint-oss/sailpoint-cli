// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
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
	var prettyPrint bool
	var jsonPath string
	var pages int
	var fetchAll bool

	cmd := &cobra.Command{
		Use:     "get [endpoint]",
		Short:   "Make a GET request to a SailPoint API endpoint",
		Long:    "\nMake a GET request to a SailPoint API endpoint\n\n",
		Example: "sail api get /beta/accounts",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.GetConfig()
			if err != nil {
				return err
			}

			spClient := client.NewSpClient(cfg)

			endpoint := args[0]
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}

			// Prepare headers
			headers := make(map[string]string)
			headers["Accept"] = "application/json"
			for _, header := range headerFlags {
				parts := strings.SplitN(header, ":", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid header format (use Key: Value): %s", header)
				}
				headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}

			ctx := context.Background()

			var body []byte
			var status string
			var paginationErr error

			if pages > 0 || fetchAll {
				userLimit, userOffset, hasLimit, hasOffset := parseQueryParams(queryParams)

				pageCfg := PaginationConfig{
					Pages:  pages,
					All:    fetchAll,
					Limit:  defaultPageSize,
					Offset: 0,
				}
				if hasLimit {
					pageCfg.Limit = userLimit
				}
				if hasOffset {
					pageCfg.Offset = userOffset
				}

				body, status, err = paginatedGet(ctx, spClient, endpoint, headers, queryParams, pageCfg)
				if err != nil {
					if len(body) > 0 {
						log.Warn("Pagination incomplete", "error", err)
						paginationErr = err
					} else {
						return err
					}
				}
			} else {
				// Single request (existing behavior)
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

				log.Debug("Making GET request", "endpoint", endpoint)

				resp, err := spClient.Get(ctx, endpoint, headers)
				if err != nil {
					return fmt.Errorf("request failed: %w", err)
				}
				defer resp.Body.Close()

				body, err = io.ReadAll(resp.Body)
				if err != nil {
					return fmt.Errorf("failed to read response: %w", err)
				}
				if err := ensureSuccess(resp, body); err != nil {
					return err
				}
				status = resp.Status
			}

			// If JSONPath is specified, evaluate it
			if jsonPath != "" {
				result, err := jsonpath.EvaluateJSONPathToString(body, jsonPath)
				if err != nil {
					return fmt.Errorf("failed to evaluate JSONPath: %w", err)
				}
				body = []byte(result)
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

			if err := writeResponse(cmd, body, status, jsonPath); err != nil {
				return err
			}

			return paginationErr
		},
	}

	cmd.Flags().StringArrayVarP(&headerFlags, "header", "H", []string{}, "HTTP headers (can be used multiple times, format: 'Key: Value')")
	cmd.Flags().StringArrayVarP(&queryParams, "query", "q", []string{}, "Query parameters (can be used multiple times, format: 'key=value')")
	cmd.Flags().BoolVarP(&prettyPrint, "pretty", "p", false, "Pretty print JSON response")
	cmd.Flags().StringVarP(&jsonPath, "jsonpath", "j", "", "JSONPath expression to evaluate on the response")
	cmd.Flags().IntVarP(&pages, "pages", "n", 0, "Number of pages to fetch (250 items per page by default)")
	cmd.Flags().BoolVarP(&fetchAll, "all", "a", false, "Fetch all results by paginating automatically")
	cmd.MarkFlagsMutuallyExclusive("pages", "all")

	return cmd
}
