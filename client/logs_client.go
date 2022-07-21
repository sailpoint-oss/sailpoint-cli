package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"path"
	"time"
)

const TimeFormatLocal = `2006-01-02T15:04:05.000-07:00`
const TimeLocationLocal = "Local"

type LogsClient struct {
	client   Client
	endpoint string
}

// NewConnClient returns a client for the provided (connectorID, version, config)
func NewLogsClient(client Client, endpoint string) *LogsClient {
	return &LogsClient{
		client:   client,
		endpoint: endpoint,
	}
}

const LogsEndpoint = "/beta/platform-logs/query"
const StatsEndpoint = "/beta/platform-logs/stats"

func logsResourceUrl(endpoint string, queryParms *map[string]string, resourceParts ...string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("invalid endpoint: %s (%q)", err, endpoint)
	}
	u.Path = path.Join(append([]string{u.Path}, resourceParts...)...)
	//set query parms
	if queryParms != nil {
		q := u.Query()
		for key, value := range *queryParms {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}
	return u.String()
}

type LogMessage struct {
	TenantID   string      `json:"tenantID"`
	Timestamp  time.Time   `json:"timestamp"`
	Level      string      `json:"level"`
	Event      string      `json:"event"`
	Component  string      `json:"component"`
	TargetID   string      `json:"targetID"`
	TargetName string      `json:"targetName"`
	RequestID  string      `json:"requestID"`
	Message    interface{} `json:"message"`
}

func (l LogMessage) RawString() string {
	json, err := json.Marshal(l)
	if err != nil {
		return fmt.Sprintf("%v", l)
	}
	return string(json)
}

func (l LogMessage) MessageString() string {
	if msgJson, ok := l.Message.(map[string]interface{}); ok {
		if jsonString, err := json.Marshal(msgJson); err == nil {
			return fmt.Sprintf("%v", string(jsonString))
		}
	}
	return fmt.Sprintf("%v", l.Message)
}

func (l LogMessage) TimestampFormatted() string {
	loc, err := time.LoadLocation(TimeLocationLocal)
	if err != nil {
		return l.Timestamp.Format(time.RFC3339)
	}
	return l.Timestamp.In(loc).Format(TimeFormatLocal)
}

type LogEvents struct {
	// The token for the next set of items in the forward direction. If you have reached the
	// end of the stream, it returns the same token you passed in.
	NextToken *string `json:"nextToken,omitempty"`
	//The log messages
	Logs []LogMessage `json:"logs"`
}

type LogFilter struct {
	StartTime  *time.Time `json:"startTime,omitempty"`
	EndTime    *time.Time `json:"endTime,omitempty"`
	Component  string     `json:"component,omitempty"`
	LogLevels  []string   `json:"logLevels,omitempty"`
	TargetID   string     `json:"targetID,omitempty"`
	TargetName string     `json:"targetName,omitempty"`
	RequestID  string     `json:"requestID,omitempty"`
	Event      string     `json:"event,omitempty"`
}
type LogInput struct {
	Filter    LogFilter `json:"filter"`
	NextToken string    `json:"nextToken"`
}

func (c *LogsClient) GetLogs(ctx context.Context, logInput LogInput) (*LogEvents, error) {

	input, err := json.Marshal(logInput)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Post(ctx, logsResourceUrl(c.endpoint, nil), "application/json", bytes.NewReader(input))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error retrieving logs, non-200 response: %s body: %s", resp.Status, body)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var logEvents LogEvents
	err = json.Unmarshal(raw, &logEvents)
	if err != nil {
		return nil, err
	}
	return &logEvents, nil
}

type TenantStats struct {
	TenantID       string           `json:"tenantID"`
	ConnectorStats []ConnectorStats `json:"connectors"`
}

type ConnectorStats struct {
	ConnectorID    string         `json:"connectorID"`
	ConnectorAlias string         `json:"alias"`
	Stats          []CommandStats `json:"stats"`
}
type CommandStats struct {
	CommandType     string  `json:"commandType"`
	InvocationCount uint32  `json:"invocationCount"`
	ErrorCount      uint32  `json:"errorCount"`
	ErrorRate       float64 `json:"errorRate"`
	ElapsedAvg      float64 `json:"elapsedAvg"`
	Elapsed95th     float64 `json:"elapsed95th"`
}

func (c CommandStats) Columns() []string {
	return []string{c.CommandType,
		fmt.Sprintf("%v", c.InvocationCount),
		fmt.Sprintf("%v", c.ErrorCount),
		fmt.Sprintf("%.2f", c.ErrorRate),
		fmt.Sprintf("%v", timeDuration(c.ElapsedAvg)),
		fmt.Sprintf("%v", timeDuration(c.Elapsed95th))}
}
func timeDuration(n float64) time.Duration {
	nRounded := math.Round(n*100) / 100
	return time.Duration(nRounded * float64(time.Millisecond))
}

func (c *LogsClient) GetStats(ctx context.Context, from time.Time, connectorID string) (*TenantStats, error) {
	queryFilter := fmt.Sprintf(`from eq "%v"`, from.Format(time.RFC3339))
	if connectorID != "" {
		queryFilter = queryFilter + fmt.Sprintf(` and connectorID eq "%v"`, connectorID)
	}
	resp, err := c.client.Get(ctx, logsResourceUrl(c.endpoint, &map[string]string{"filters": queryFilter}))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error retrieving logs, non-200 response: %s. Body: %s", resp.Status, body)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tenantStats TenantStats
	err = json.Unmarshal(raw, &tenantStats)
	if err != nil {
		return nil, err
	}
	return &tenantStats, nil
}

// Define the order of time formats to attempt to use to parse our input absolute time
var absoluteTimeFormats = []string{
	time.RFC3339,
	"2006-01-02",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05.000-07:00",
	"2006-01-02T15:04:05.000Z",
}

// Parse the input string into a time.Time object.
// Provide the currentTime as a parameter to support relative time.
func ParseTime(timeStr string, currentTime time.Time) (time.Time, error) {
	relative, err := time.ParseDuration(timeStr)
	if err == nil {
		return currentTime.Add(-relative), nil
	}

	// Iterate over available absolute time formats until we find one that works
	for _, timeFormat := range absoluteTimeFormats {
		absolute, err := time.Parse(timeFormat, timeStr)

		if err == nil {
			return absolute, err
		}
	}

	return time.Time{}, fmt.Errorf("could not parse relative or absolute time")
}
