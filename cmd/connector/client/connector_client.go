package connclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/sailpoint-oss/sailpoint-cli/internal/client"
)

const maskedPassword = "******"

// ConnClient is an sail connect client for a specific connector
type ConnClient struct {
	client       client.Client
	version      *int
	config       json.RawMessage
	connectorRef string
	endpoint     string
}

// NewConnClient returns a client for the provided (connectorID, version, config)
func NewConnClient(client client.Client, version *int, config json.RawMessage, connectorRef string, endpoint string) *ConnClient {
	return &ConnClient{
		client:       client,
		version:      version,
		config:       config,
		connectorRef: connectorRef,
		endpoint:     endpoint,
	}
}

// TestConnectionWithConfig provides a way to run std:test-connection with an
// arbitrary config
func (cc *ConnClient) TestConnectionWithConfig(ctx context.Context, cfg json.RawMessage) error {
	cmdRaw, err := cc.rawInvokeWithConfig("std:test-connection", []byte("{}"), cfg, nil)
	if err != nil {
		return err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return newResponseError(resp)
	}
	return nil
}

// TestConnection runs the std:test-connection command
func (cc *ConnClient) TestConnection(ctx context.Context) (rawResponse []byte, err error) {
	cmdRaw, err := cc.rawInvoke("std:test-connection", []byte("{}"))
	if err != nil {
		return nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, newResponseError(resp)
	}

	rawResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rawResp, err := parseResponse(rawResponse)
	if err != nil {
		return nil, err
	}

	return rawResp.Data, nil
}

// ChangePassword runs the std:change-password command
func (cc *ConnClient) ChangePassword(ctx context.Context, identity string, uniqueID string, password string) (rawResponse []byte, err error) {

	var key Key
	if uniqueID == "" {
		key = NewSimpleKey(identity)
	} else {
		key = NewCompoundKey(identity, uniqueID)
	}

	input, err := json.Marshal(map[string]interface{}{
		"identity": identity,
		"key":      key,
		"password": password,
	})
	if err != nil {
		return nil, err
	}

	maskedInput, err := json.Marshal(map[string]interface{}{
		"identity": identity,
		"key":      key,
		"password": maskedPassword,
	})
	if err != nil {
		return nil, err
	}

	cmdRaw, err := cc.rawInvokeWithConfig("std:change-password", input, cc.config, maskedInput)
	if err != nil {
		return nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, newResponseError(resp)
	}

	rawResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rawResp, err := parseResponse(rawResponse)
	if err != nil {
		return nil, err
	}

	return rawResp.Data, nil
}

type SimpleKey struct {
	ID string `json:"id"`
}

type CompoundKey struct {
	LookupID string `json:"lookupId"`
	UniqueID string `json:"uniqueId"`
}

type Key struct {
	Simple   *SimpleKey   `json:"simple,omitempty"`
	Compound *CompoundKey `json:"compound,omitempty"`
}

func NewSimpleKey(id string) Key {
	return Key{
		Simple: &SimpleKey{
			ID: id,
		},
	}
}

func NewCompoundKey(lookupID string, uniqueID string) Key {
	return Key{
		Compound: &CompoundKey{
			LookupID: lookupID,
			UniqueID: uniqueID,
		},
	}
}

// Account is an sail connect account. The is used for AccountList, AccountRead
// and AccountUpdate commands.
type Account struct {
	Identity   string                 `json:"identity"`
	UUID       string                 `json:"uuid"`
	Key        Key                    `json:"key"`
	Attributes map[string]interface{} `json:"attributes"`
}

func (a *Account) ID() string {
	if a.Key.Simple != nil {
		return a.Key.Simple.ID
	}
	if a.Key.Compound != nil {
		return a.Key.Compound.LookupID
	}
	return a.Identity
}

func (a *Account) UniqueID() string {
	if a.Key.Compound != nil {
		return a.Key.Compound.UniqueID
	}
	if a.UUID != "" {
		return a.UUID
	}
	return ""
}

type accountListInput struct {
	Stateful *bool                  `json:"stateful,omitempty"`
	StateID  *string                `json:"stateId,omitempty"`
	Schema   map[string]interface{} `json:"schema,omitempty"`
}

// AccountList lists all accounts
func (cc *ConnClient) AccountList(ctx context.Context, stateful *bool, stateId *string, schema map[string]interface{}) (accounts []Account, state json.RawMessage, printable []byte, err error) {
	inputRaw, err := json.Marshal(accountListInput{
		Stateful: stateful,
		StateID:  stateId,
		Schema:   schema,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:account:list", inputRaw)
	if err != nil {
		return nil, nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, nil, newResponseError(resp)
	}

	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	rawResps, s, err := parseResponseList(rawResponse)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, r := range rawResps {
		acct := &Account{}
		err := json.Unmarshal(r.Data, acct)
		if err != nil {
			return nil, nil, nil, err
		}
		accounts = append(accounts, *acct)

		if len(printable) != 0 {
			printable = append(printable, []byte("\n")...)
		}
		printable = append(printable, r.Data...)
	}

	if s != nil {
		state = s.Data
	}

	return accounts, state, printable, nil
}

type readInput struct {
	Identity string                 `json:"identity"`
	Key      Key                    `json:"key"`
	Type     string                 `json:"type,omitempty"`
	Schema   map[string]interface{} `json:"schema,omitempty"`
}

// AccountRead reads a specific account
func (cc *ConnClient) AccountRead(ctx context.Context, id string, uniqueID string, schema map[string]interface{}) (account *Account, rawResponse []byte, err error) {
	input := readInput{
		Identity: id,
		Schema:   schema,
	}
	if uniqueID == "" {
		input.Key = NewSimpleKey(id)
	} else {
		input.Key = NewCompoundKey(id, uniqueID)
	}

	inRaw, err := json.Marshal(input)
	if err != nil {
		return nil, nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:account:read", inRaw)
	if err != nil {
		return nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, newResponseError(resp)
	}

	rawResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	rawResp, err := parseResponse(rawResponse)
	if err != nil {
		return nil, nil, err
	}

	acct := &Account{}
	err = json.Unmarshal(rawResp.Data, acct)
	if err != nil {
		return nil, nil, err
	}

	return acct, rawResp.Data, nil
}

type accountCreateInput struct {
	Identity   *string                `json:"identity"`
	Attributes map[string]interface{} `json:"attributes"`
	Schema     map[string]interface{} `json:"schema,omitempty"`
}

// AccountCreate creats an account
func (cc *ConnClient) AccountCreate(ctx context.Context, identity *string, attributes map[string]interface{}, schema map[string]interface{}) (account *Account, raw []byte, err error) {
	input, err := json.Marshal(accountCreateInput{
		Identity:   identity,
		Attributes: attributes,
		Schema:     schema,
	})
	if err != nil {
		return nil, nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:account:create", input)
	if err != nil {
		return nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, newResponseError(resp)
	}

	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	rawResp, err := parseResponse(raw)
	if err != nil {
		return nil, nil, err
	}

	acct := &Account{}
	err = json.Unmarshal(rawResp.Data, acct)
	if err != nil {
		return nil, nil, err
	}

	return acct, rawResp.Data, nil
}

// AccountDelete deletes an account
func (cc *ConnClient) AccountDelete(ctx context.Context, id string, uniqueID string, schema map[string]interface{}) (raw []byte, err error) {
	input := readInput{
		Identity: id,
		Schema:   schema,
	}
	if uniqueID == "" {
		input.Key = NewSimpleKey(id)
	} else {
		input.Key = NewCompoundKey(id, uniqueID)
	}

	inRaw, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:account:delete", inRaw)
	if err != nil {
		return nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, newResponseError(resp)
	}

	return nil, nil
}

// AttributeChange describes a change to a specific attribute
type AttributeChange struct {
	Op        string      `json:"op"`
	Attribute string      `json:"attribute"`
	Value     interface{} `json:"value"`
}

// AccountUpdate updates an account
func (cc *ConnClient) AccountUpdate(ctx context.Context, id string, uniqueID string, changes []AttributeChange, schema map[string]interface{}) (account *Account, rawResponse []byte, err error) {
	type accountUpdate struct {
		readInput
		Changes []AttributeChange `json:"changes"`
	}

	input := readInput{
		Identity: id,
		Schema:   schema,
	}
	if uniqueID == "" {
		input.Key = NewSimpleKey(id)
	} else {
		input.Key = NewCompoundKey(id, uniqueID)
	}

	inRaw, err := json.Marshal(accountUpdate{
		readInput: input,
		Changes:   changes,
	})
	if err != nil {
		return nil, nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:account:update", inRaw)
	if err != nil {
		return nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, newResponseError(resp)
	}

	rawResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	rawResp, err := parseResponse(rawResponse)
	if err != nil {
		return nil, nil, err
	}

	acct := &Account{}
	err = json.Unmarshal(rawResp.Data, acct)
	if err != nil {
		return nil, nil, err
	}

	return acct, rawResp.Data, nil
}

// AccountDiscoverSchema discovers schema for accounts
func (cc *ConnClient) AccountDiscoverSchema(ctx context.Context) (accountSchema *AccountSchema, rawResponse []byte, err error) {
	cmdRaw, err := cc.rawInvoke("std:account:discover-schema", []byte("{}"))
	if err != nil {
		return nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, newResponseError(resp)
	}

	rawResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	rawResp, err := parseResponse(rawResponse)
	if err != nil {
		return nil, nil, err
	}

	schema := &AccountSchema{}
	err = json.Unmarshal(rawResp.Data, schema)
	if err != nil {
		return nil, nil, err
	}

	return schema, rawResponse, nil
}

// Entitlement is an sail connect entitlement, used for EntitlementList and
// EntitlementRead
type Entitlement struct {
	Identity   string                 `json:"identity"`
	UUID       string                 `json:"uuid"`
	Key        Key                    `json:"key"`
	Attributes map[string]interface{} `json:"attributes"`
}

func (a *Entitlement) ID() string {
	if a.Key.Simple != nil {
		return a.Key.Simple.ID
	}
	if a.Key.Compound != nil {
		return a.Key.Compound.LookupID
	}
	return a.Identity
}

func (a *Entitlement) UniqueID() string {
	if a.Key.Compound != nil {
		return a.Key.Compound.UniqueID
	}
	if a.UUID != "" {
		return a.UUID
	}
	return ""
}

type entitlementListInput struct {
	accountListInput
	Type string `json:"type"`
}

// EntitlementList lists all entitlements
func (cc *ConnClient) EntitlementList(ctx context.Context, t string, stateful *bool, stateId *string, schema map[string]interface{}) (entitlements []Entitlement, state json.RawMessage, printable []byte, err error) {
	inputRaw, err := json.Marshal(entitlementListInput{
		Type: t,
		accountListInput: accountListInput{
			Stateful: stateful,
			StateID:  stateId,
			Schema:   schema,
		},
	})
	if err != nil {
		return nil, nil, nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:entitlement:list", inputRaw)
	if err != nil {
		return nil, nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, nil, newResponseError(resp)
	}

	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	rawResps, s, err := parseResponseList(rawResponse)
	if err != nil {
		return nil, nil, nil, err
	}

	for _, r := range rawResps {
		e := &Entitlement{}
		err := json.Unmarshal(r.Data, e)
		if err != nil {
			return nil, nil, nil, err
		}
		entitlements = append(entitlements, *e)

		if len(printable) != 0 {
			printable = append(printable, []byte("\n")...)
		}
		printable = append(printable, r.Data...)
	}

	if s != nil {
		state = s.Data
	}

	return entitlements, state, printable, nil
}

// EntitlementRead reads all entitlements
func (cc *ConnClient) EntitlementRead(ctx context.Context, id string, uniqueID string, t string, schema map[string]interface{}) (entitlement *Entitlement, rawResponse []byte, err error) {
	input := readInput{
		Identity: id,
		Type:     t,
		Schema:   schema,
	}

	if uniqueID == "" {
		input.Key = NewSimpleKey(id)
	} else {
		input.Key = NewCompoundKey(id, uniqueID)
	}

	inRaw, err := json.Marshal(input)
	if err != nil {
		return nil, nil, err
	}
	cmdRaw, err := cc.rawInvoke("std:entitlement:read", inRaw)
	if err != nil {
		return nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, newResponseError(resp)
	}

	rawResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	rawResp, err := parseResponse(rawResponse)
	if err != nil {
		return nil, nil, err
	}

	e := &Entitlement{}
	err = json.Unmarshal(rawResp.Data, e)
	if err != nil {
		return nil, nil, err
	}

	return e, rawResp.Data, nil
}

type ReadSpecOutput struct {
	Specification *ConnSpec `json:"specification"`
}

// SpecRead issues a custom:config command which is expected to return the
// connector specification. This is an experimental command used by the
// validation suite.
func (cc *ConnClient) SpecRead(ctx context.Context) (connSpec *ConnSpec, err error) {
	cmdRaw, err := cc.rawInvoke("std:spec:read", []byte(`{}`))
	if err != nil {
		return nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, newResponseError(resp)
	}

	rawResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rawResp, err := parseResponse(rawResponse)
	if err != nil {
		return nil, err
	}

	cfg := &ReadSpecOutput{}
	err = json.Unmarshal(rawResp.Data, cfg)
	if err != nil {
		return nil, err
	}

	return cfg.Specification, nil
}

// Invoke allows you to send an arbitrary json payload as a command
func (cc *ConnClient) Invoke(ctx context.Context, cmdType string, input json.RawMessage) (rawResponse []byte, err error) {
	cmdRaw, err := cc.rawInvoke(cmdType, input)
	if err != nil {
		return nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, newResponseError(resp)
	}

	rawResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return rawResponse, nil
}

func newResponseError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var errorPayload interface{}
	err := json.Unmarshal(body, &errorPayload)
	if err != nil {
		return fmt.Errorf("non-200 response: %s (body %s)", resp.Status, string(body))
	} else {
		pretty, err := json.MarshalIndent(errorPayload, "", "\t")
		if err != nil {
			return fmt.Errorf("non-200 response: %s (body %s)", resp.Status, string(body))
		} else {
			return fmt.Errorf("non-200 response: %s (body %s)", resp.Status, string(pretty))
		}
	}
}

type AccountCreateTemplateField struct {
	// Deprecated
	Name string `json:"name"`

	Key          string               `json:"key"`
	Type         string               `json:"type"`
	Required     bool                 `json:"required"`
	InitialValue TemplateInitialValue `json:"initialValue"`
}

type TemplateInitialValue struct {
	Type       string             `json:"type"`
	Attributes TemplateAttributes `json:"attributes"`
}

type TemplateAttributes struct {
	Name     string      `json:"name"`
	Value    interface{} `json:"value"`
	Template string      `json:"template"`
}

type AccountCreateTemplate struct {
	Fields []AccountCreateTemplateField `json:"fields"`
}

type AccountSchema struct {
	DisplayAttribute  string                   `json:"displayAttribute"`
	GroupAttribute    string                   `json:"groupAttribute"`
	IdentityAttribute string                   `json:"identityAttribute"`
	Attributes        []AccountSchemaAttribute `json:"attributes"`
}

type EntitlementSchema struct {
	Type               string                       `json:"type"`
	DisplayAttribute   string                       `json:"displayAttribute"`
	IdentityAttribute  string                       `json:"identityAttribute"`
	HierarchyAttribute string                       `json:"hierarchyAttribute"`
	Attributes         []EntitlementSchemaAttribute `json:"attributes"`
}

type AccountSchemaAttribute struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`

	Entitlement bool `json:"entitlement"`
	Managed     bool `json:"managed"`
	Multi       bool `json:"multi"`

	// Writable is not a standard spec field, yet
	Writable bool `json:"writable"`
}

type EntitlementSchemaAttribute struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`

	Multi    bool `json:"multi"`
	Required bool `json:"required"`
}

// ConnSpec is a connector config. See ConnConfig method.
type ConnSpec struct {
	Name                  string                `json:"name"`
	Commands              []string              `json:"commands"`
	AccountCreateTemplate AccountCreateTemplate `json:"accountCreateTemplate"`
	AccountSchema         AccountSchema         `json:"accountSchema"`
	EntitlementSchemas    []EntitlementSchema   `json:"entitlementSchemas"`
}

func (cc *ConnClient) rawInvoke(cmdType string, input json.RawMessage) (json.RawMessage, error) {
	return cc.rawInvokeWithConfig(cmdType, input, cc.config, nil)
}

func (cc *ConnClient) rawInvokeWithConfig(cmdType string, input json.RawMessage, cfg json.RawMessage, maskedInput []byte) (json.RawMessage, error) {

	// if input contains sensitive information, log the masked input to console
	if maskedInput == nil {
		log.Printf("Running %q with %q", cmdType, input)
	} else {
		log.Printf("Running %q with %q", cmdType, maskedInput)
	}

	invokeCmd := invokeCommand{
		ConnectorRef: cc.connectorRef,
		Type:         cmdType,
		Config:       cfg,
		Input:        input,
	}

	if cc.version == nil {
		invokeCmd.Tag = "latest"
	} else {
		invokeCmd.Version = cc.version
	}

	return json.Marshal(invokeCmd)
}

func connResourceUrl(endpoint string, resourceParts ...string) string {
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalf("invalid endpoint: %s (%q)", err, endpoint)
	}
	u.Path = path.Join(append([]string{u.Path}, resourceParts...)...)
	return u.String()
}

type invokeCommand struct {
	ConnectorRef string          `json:"connectorRef"`
	Version      *int            `json:"version,omitempty"`
	Tag          string          `json:"tag,omitempty"`
	Type         string          `json:"type"`
	Config       json.RawMessage `json:"config"`
	Input        json.RawMessage `json:"input"`
}

type sourceDataDiscoverInput struct {
	Query map[string]any `json:"queryInput"`
}

type sourceDataReadInput struct {
	SourceDataKey string         `json:"sourceDataKey"`
	Query         map[string]any `json:"queryInput"`
}

type sourceData struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	SubLabel string `json:"sublabel"`
}

func (cc *ConnClient) SourceDataDiscover(ctx context.Context, queryInput map[string]any) (sData []sourceData, raw []byte, err error) {
	input, err := json.Marshal(sourceDataDiscoverInput{
		Query: queryInput,
	})
	if err != nil {
		return nil, nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:source-data:discover", input)
	if err != nil {
		return nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, newResponseError(resp)
	}

	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	rawResp, err := parseResponse(raw)
	if err != nil {
		return nil, nil, err
	}

	data := make([]sourceData, 0)
	err = json.Unmarshal(rawResp.Data, &data)
	if err != nil {
		return nil, nil, err
	}

	return data, rawResp.Data, nil
}

func (cc *ConnClient) SourceDataRead(ctx context.Context, sourceDataKey string, queryInput map[string]any) (sData []sourceData, raw []byte, err error) {
	input, err := json.Marshal(sourceDataReadInput{
		SourceDataKey: sourceDataKey,
		Query:         queryInput,
	})
	if err != nil {
		return nil, nil, err
	}

	cmdRaw, err := cc.rawInvoke("std:source-data:read", input)
	if err != nil {
		return nil, nil, err
	}

	resp, err := cc.client.Post(ctx, connResourceUrl(cc.endpoint, cc.connectorRef, "invoke-direct"), "application/json", bytes.NewReader(cmdRaw))
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return nil, nil, newResponseError(resp)
	}

	raw, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	rawResp, err := parseResponse(raw)
	if err != nil {
		return nil, nil, err
	}

	data := make([]sourceData, 0)
	err = json.Unmarshal(rawResp.Data, &data)
	if err != nil {
		return nil, nil, err
	}

	return data, rawResp.Data, nil
}
