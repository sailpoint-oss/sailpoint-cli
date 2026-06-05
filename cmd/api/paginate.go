// Copyright (c) 2024, SailPoint Technologies, Inc. All rights reserved.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
	"github.com/sailpoint-oss/sailpoint-cli/internal/clierror"
)

const defaultPageSize = 250

type PaginationConfig struct {
	Pages  int
	All    bool
	Limit  int
	Offset int
}

func parseQueryParams(queryParams []string) (limit, offset int, hasLimit, hasOffset bool) {
	for _, param := range queryParams {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])
		switch key {
		case "limit":
			if v, err := strconv.Atoi(val); err == nil {
				limit = v
				hasLimit = true
			}
		case "offset":
			if v, err := strconv.Atoi(val); err == nil {
				offset = v
				hasOffset = true
			}
		}
	}
	return
}

func buildPaginatedEndpoint(endpoint string, queryParams []string, offset, limit int, includeCount bool) (string, error) {
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	query := parsedURL.Query()

	for _, param := range queryParams {
		parts := strings.SplitN(param, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		keyLower := strings.ToLower(key)
		if keyLower == "limit" || keyLower == "offset" || keyLower == "count" {
			continue
		}
		query.Add(key, strings.TrimSpace(parts[1]))
	}

	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))
	if includeCount {
		query.Set("count", "true")
	}

	parsedURL.RawQuery = query.Encode()
	return parsedURL.String(), nil
}

func getWithRetry(ctx context.Context, spClient client.Client, url string, headers map[string]string, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = spClient.Get(ctx, url, headers)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		resp.Body.Close()

		if attempt == maxRetries {
			return nil, fmt.Errorf("rate limited after %d retries", maxRetries)
		}

		waitTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				waitTime = time.Duration(seconds) * time.Second
			}
		}

		log.Debug("Rate limited, retrying", "attempt", attempt+1, "wait", waitTime)
		time.Sleep(waitTime)
	}

	return resp, nil
}

func paginatedGet(ctx context.Context, spClient client.Client, endpoint string, headers map[string]string, queryParams []string, pageCfg PaginationConfig) ([]byte, string, error) {
	limit := pageCfg.Limit
	offset := pageCfg.Offset

	reqURL, err := buildPaginatedEndpoint(endpoint, queryParams, offset, limit, true)
	if err != nil {
		return nil, "", err
	}

	log.Debug("Paginated GET", "url", reqURL, "page", 1)
	resp, err := getWithRetry(ctx, spClient, reqURL, headers, 3)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp.Status, clierror.APIStatus(resp.StatusCode, resp.Status, body)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	var firstPage []json.RawMessage
	if err := json.Unmarshal(body, &firstPage); err != nil {
		return nil, "", fmt.Errorf("pagination only supported for endpoints returning JSON arrays")
	}

	totalCountStr := resp.Header.Get("X-Total-Count")
	totalCount := -1
	if totalCountStr != "" {
		if tc, err := strconv.Atoi(totalCountStr); err == nil {
			totalCount = tc
		}
	}

	if pageCfg.All && totalCount < 0 {
		return nil, "", fmt.Errorf("--all requires X-Total-Count header but it was not present in the response")
	}

	var totalPages int
	if pageCfg.All {
		totalPages = int(math.Ceil(float64(totalCount-offset) / float64(limit)))
	} else {
		totalPages = pageCfg.Pages
	}

	allItems := make([]json.RawMessage, 0, len(firstPage))
	allItems = append(allItems, firstPage...)
	lastStatus := resp.Status

	if len(firstPage) < limit {
		totalPages = 1
	}

	for page := 2; page <= totalPages; page++ {
		offset += limit

		reqURL, err = buildPaginatedEndpoint(endpoint, queryParams, offset, limit, false)
		if err != nil {
			return nil, "", err
		}

		log.Debug("Paginated GET", "url", reqURL, "page", page)
		pageResp, err := getWithRetry(ctx, spClient, reqURL, headers, 3)
		if err != nil {
			merged, _ := json.Marshal(allItems)
			return merged, lastStatus, fmt.Errorf("failed on page %d: %w (returning %d items collected so far)", page, err, len(allItems))
		}

		pageBody, readErr := io.ReadAll(pageResp.Body)
		pageResp.Body.Close()

		if pageResp.StatusCode < 200 || pageResp.StatusCode >= 300 {
			log.Warn("Non-success status on page", "page", page, "status", pageResp.Status)
			merged, _ := json.Marshal(allItems)
			return merged, lastStatus, fmt.Errorf("page %d returned status %s (returning %d items collected so far)", page, pageResp.Status, len(allItems))
		}

		if readErr != nil {
			merged, _ := json.Marshal(allItems)
			return merged, lastStatus, fmt.Errorf("failed to read page %d response: %w (returning %d items collected so far)", page, readErr, len(allItems))
		}

		var pageItems []json.RawMessage
		if err := json.Unmarshal(pageBody, &pageItems); err != nil {
			merged, _ := json.Marshal(allItems)
			return merged, lastStatus, fmt.Errorf("page %d returned non-array response (returning %d items collected so far)", page, len(allItems))
		}

		allItems = append(allItems, pageItems...)
		lastStatus = pageResp.Status

		if len(pageItems) < limit {
			break
		}
	}

	merged, err := json.Marshal(allItems)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal results: %w", err)
	}

	statusMsg := fmt.Sprintf("%s (fetched %d items across pages)", lastStatus, len(allItems))
	return merged, statusMsg, nil
}
