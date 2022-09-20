package client

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
)

// ConnClient is an sp connect client for a specific connector
type ConnClient struct {
	client       Client
	version      *int
	config       json.RawMessage
	connectorRef string
	endpoint     string
}

// NewConnClient returns a client for the provided (connectorID, version, config)
func NewConnClient(client Client, version *int, config json.RawMessage, connectorRef string, endpoint string) *ConnClient {
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
	cmdRaw, err := cc.rawInvokeWithConfig("std:test-connection", []byte("{}"), cfg)
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

	return io.ReadAll(resp.Body)
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

// Account is an sp connect account. The is used for AccountList, AccountRead
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

// AccountList lists all accounts
func (cc *ConnClient) AccountList(ctx context.Context) (accounts []Account, rawResponse []byte, err error) {
	cmdRaw, err := cc.rawInvoke("std:account:list", []byte("{}"))
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

	decoder := json.NewDecoder(bytes.NewReader(rawResponse))
	for {
		acct := &Account{}
		err := decoder.Decode(acct)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		accounts = append(accounts, *acct)
	}

	return accounts, rawResponse, nil
}

type readInput struct {
	Identity string `json:"identity"`
	Key      Key    `json:"key"`
	Type     string `json:"type,omitempty"`
}

// AccountRead reads a specific account
func (cc *ConnClient) AccountRead(ctx context.Context, id string, uniqueID string) (account *Account, rawResponse []byte, err error) {
	input := readInput{
		Identity: id,
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

	decoder := json.NewDecoder(bytes.NewReader(rawResponse))
	acct := &Account{}
	err = decoder.Decode(acct)
	if err != nil {
		return nil, nil, err
	}

	return acct, rawResponse, nil
}

// AccountCreate creats an account
func (cc *ConnClient) AccountCreate(ctx context.Context, identity *string, attributes map[string]interface{}) (account *Account, raw []byte, err error) {
	input, err := json.Marshal(map[string]interface{}{
		"identity":   identity,
		"attributes": attributes,
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

	acct := &Account{}
	err = json.Unmarshal(raw, acct)
	if err != nil {
		return nil, nil, err
	}

	return acct, raw, nil
}

// AccountDelete deletes an account
func (cc *ConnClient) AccountDelete(ctx context.Context, id string, uniqueID string) (raw []byte, err error) {
	input := readInput{
		Identity: id,
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
func (cc *ConnClient) AccountUpdate(ctx context.Context, id string, uniqueID string, changes []AttributeChange) (account *Account, rawResponse []byte, err error) {
	type accountUpdate struct {
		readInput
		Changes []AttributeChange `json:"changes"`
	}

	input := readInput{
		Identity: id,
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

	decoder := json.NewDecoder(bytes.NewReader(rawResponse))
	acct := &Account{}
	err = decoder.Decode(acct)
	if err != nil {
		return nil, nil, err
	}

	return account, rawResponse, nil
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

	decoder := json.NewDecoder(bytes.NewReader(rawResponse))
	schema := &AccountSchema{}
	err = decoder.Decode(schema)
	if err != nil {
		return nil, nil, err
	}

	return schema, rawResponse, nil
}

// Entitlement is an sp connect entitlement, used for EntitlementList and
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

// EntitlementList lists all entitlements
func (cc *ConnClient) EntitlementList(ctx context.Context, t string) (entitlements []Entitlement, rawResponse []byte, err error) {
	cmdRaw, err := cc.rawInvoke("std:entitlement:list", []byte(fmt.Sprintf(`{"type": %q}`, t)))
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

	decoder := json.NewDecoder(bytes.NewReader(rawResponse))
	for {
		e := &Entitlement{}
		err := decoder.Decode(e)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}
		entitlements = append(entitlements, *e)
	}

	return entitlements, rawResponse, nil
}

// EntitlementRead reads all entitlements
func (cc *ConnClient) EntitlementRead(ctx context.Context, id string, uniqueID string, t string) (entitlement *Entitlement, rawResponse []byte, err error) {
	input := readInput{
		Identity: id,
		Type:     t,
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

	decoder := json.NewDecoder(bytes.NewReader(rawResponse))
	e := &Entitlement{}
	err = decoder.Decode(e)
	if err != nil {
		return nil, nil, err
	}

	return e, rawResponse, nil
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

	decoder := json.NewDecoder(resp.Body)
	cfg := &ReadSpecOutput{}
	err = decoder.Decode(cfg)
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
	Type              string                       `json:"type"`
	DisplayName       string                       `json:"displayName"`
	IdentityAttribute string                       `json:"identityAttribute"`
	Attributes        []EntitlementSchemaAttribute `json:"attributes"`
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

	Multi bool `json:"multi"`
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
	return cc.rawInvokeWithConfig(cmdType, input, cc.config)
}

func (cc *ConnClient) rawInvokeWithConfig(cmdType string, input json.RawMessage, cfg json.RawMessage) (json.RawMessage, error) {
	log.Printf("Running %q with %q", cmdType, input)
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

const connectorsEndpoint = "/beta/platform-connectors"

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
