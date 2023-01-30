/*
IdentityNow V3 API

Use these APIs to interact with the IdentityNow platform to achieve repeatable, automated processes with greater scalability. We encourage you to join the SailPoint Developer Community forum at https://developer.sailpoint.com/discuss to connect with other developers using our APIs.

API version: 3.0.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package sailpointsdk

import (
	"time"
	"encoding/json"
)

// Account1 Account
type Account1 struct {
	Id string `json:"id"`
	Name string `json:"name"`
	Type DocumentType `json:"_type"`
	// The ID of the account
	AccountId *string `json:"accountId,omitempty"`
	Source *Source1 `json:"source,omitempty"`
	// Indicates if the account is disabled
	Disabled *bool `json:"disabled,omitempty"`
	// Indicates if the account is locked
	Locked *bool `json:"locked,omitempty"`
	Privileged *bool `json:"privileged,omitempty"`
	// Indicates if the account has been manually correlated to an identity
	ManuallyCorrelated *bool `json:"manuallyCorrelated,omitempty"`
	// A date-time in ISO-8601 format
	PasswordLastSet NullableTime `json:"passwordLastSet,omitempty"`
	// a map or dictionary of key/value pairs
	EntitlementAttributes map[string]interface{} `json:"entitlementAttributes,omitempty"`
	// A date-time in ISO-8601 format
	Created NullableTime `json:"created,omitempty"`
	// A date-time in ISO-8601 format
	Modified NullableTime `json:"modified,omitempty"`
	// a map or dictionary of key/value pairs
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Identity *DisplayReference `json:"identity,omitempty"`
	Access []Entitlement1 `json:"access,omitempty"`
	// The number of entitlements assigned to the account
	EntitlementCount *int32 `json:"entitlementCount,omitempty"`
	// Indicates if the account is not correlated to an identity
	Uncorrelated *bool `json:"uncorrelated,omitempty"`
	Tags []string `json:"tags,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _Account1 Account1

// NewAccount1 instantiates a new Account1 object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAccount1(id string, name string, type_ DocumentType) *Account1 {
	this := Account1{}
	this.Id = id
	this.Name = name
	this.Type = type_
	return &this
}

// NewAccount1WithDefaults instantiates a new Account1 object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAccount1WithDefaults() *Account1 {
	this := Account1{}
	return &this
}

// GetId returns the Id field value
func (o *Account1) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *Account1) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *Account1) SetId(v string) {
	o.Id = v
}

// GetName returns the Name field value
func (o *Account1) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *Account1) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *Account1) SetName(v string) {
	o.Name = v
}

// GetType returns the Type field value
func (o *Account1) GetType() DocumentType {
	if o == nil {
		var ret DocumentType
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *Account1) GetTypeOk() (*DocumentType, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *Account1) SetType(v DocumentType) {
	o.Type = v
}

// GetAccountId returns the AccountId field value if set, zero value otherwise.
func (o *Account1) GetAccountId() string {
	if o == nil || isNil(o.AccountId) {
		var ret string
		return ret
	}
	return *o.AccountId
}

// GetAccountIdOk returns a tuple with the AccountId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetAccountIdOk() (*string, bool) {
	if o == nil || isNil(o.AccountId) {
		return nil, false
	}
	return o.AccountId, true
}

// HasAccountId returns a boolean if a field has been set.
func (o *Account1) HasAccountId() bool {
	if o != nil && !isNil(o.AccountId) {
		return true
	}

	return false
}

// SetAccountId gets a reference to the given string and assigns it to the AccountId field.
func (o *Account1) SetAccountId(v string) {
	o.AccountId = &v
}

// GetSource returns the Source field value if set, zero value otherwise.
func (o *Account1) GetSource() Source1 {
	if o == nil || isNil(o.Source) {
		var ret Source1
		return ret
	}
	return *o.Source
}

// GetSourceOk returns a tuple with the Source field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetSourceOk() (*Source1, bool) {
	if o == nil || isNil(o.Source) {
		return nil, false
	}
	return o.Source, true
}

// HasSource returns a boolean if a field has been set.
func (o *Account1) HasSource() bool {
	if o != nil && !isNil(o.Source) {
		return true
	}

	return false
}

// SetSource gets a reference to the given Source1 and assigns it to the Source field.
func (o *Account1) SetSource(v Source1) {
	o.Source = &v
}

// GetDisabled returns the Disabled field value if set, zero value otherwise.
func (o *Account1) GetDisabled() bool {
	if o == nil || isNil(o.Disabled) {
		var ret bool
		return ret
	}
	return *o.Disabled
}

// GetDisabledOk returns a tuple with the Disabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetDisabledOk() (*bool, bool) {
	if o == nil || isNil(o.Disabled) {
		return nil, false
	}
	return o.Disabled, true
}

// HasDisabled returns a boolean if a field has been set.
func (o *Account1) HasDisabled() bool {
	if o != nil && !isNil(o.Disabled) {
		return true
	}

	return false
}

// SetDisabled gets a reference to the given bool and assigns it to the Disabled field.
func (o *Account1) SetDisabled(v bool) {
	o.Disabled = &v
}

// GetLocked returns the Locked field value if set, zero value otherwise.
func (o *Account1) GetLocked() bool {
	if o == nil || isNil(o.Locked) {
		var ret bool
		return ret
	}
	return *o.Locked
}

// GetLockedOk returns a tuple with the Locked field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetLockedOk() (*bool, bool) {
	if o == nil || isNil(o.Locked) {
		return nil, false
	}
	return o.Locked, true
}

// HasLocked returns a boolean if a field has been set.
func (o *Account1) HasLocked() bool {
	if o != nil && !isNil(o.Locked) {
		return true
	}

	return false
}

// SetLocked gets a reference to the given bool and assigns it to the Locked field.
func (o *Account1) SetLocked(v bool) {
	o.Locked = &v
}

// GetPrivileged returns the Privileged field value if set, zero value otherwise.
func (o *Account1) GetPrivileged() bool {
	if o == nil || isNil(o.Privileged) {
		var ret bool
		return ret
	}
	return *o.Privileged
}

// GetPrivilegedOk returns a tuple with the Privileged field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetPrivilegedOk() (*bool, bool) {
	if o == nil || isNil(o.Privileged) {
		return nil, false
	}
	return o.Privileged, true
}

// HasPrivileged returns a boolean if a field has been set.
func (o *Account1) HasPrivileged() bool {
	if o != nil && !isNil(o.Privileged) {
		return true
	}

	return false
}

// SetPrivileged gets a reference to the given bool and assigns it to the Privileged field.
func (o *Account1) SetPrivileged(v bool) {
	o.Privileged = &v
}

// GetManuallyCorrelated returns the ManuallyCorrelated field value if set, zero value otherwise.
func (o *Account1) GetManuallyCorrelated() bool {
	if o == nil || isNil(o.ManuallyCorrelated) {
		var ret bool
		return ret
	}
	return *o.ManuallyCorrelated
}

// GetManuallyCorrelatedOk returns a tuple with the ManuallyCorrelated field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetManuallyCorrelatedOk() (*bool, bool) {
	if o == nil || isNil(o.ManuallyCorrelated) {
		return nil, false
	}
	return o.ManuallyCorrelated, true
}

// HasManuallyCorrelated returns a boolean if a field has been set.
func (o *Account1) HasManuallyCorrelated() bool {
	if o != nil && !isNil(o.ManuallyCorrelated) {
		return true
	}

	return false
}

// SetManuallyCorrelated gets a reference to the given bool and assigns it to the ManuallyCorrelated field.
func (o *Account1) SetManuallyCorrelated(v bool) {
	o.ManuallyCorrelated = &v
}

// GetPasswordLastSet returns the PasswordLastSet field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Account1) GetPasswordLastSet() time.Time {
	if o == nil || isNil(o.PasswordLastSet.Get()) {
		var ret time.Time
		return ret
	}
	return *o.PasswordLastSet.Get()
}

// GetPasswordLastSetOk returns a tuple with the PasswordLastSet field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Account1) GetPasswordLastSetOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return o.PasswordLastSet.Get(), o.PasswordLastSet.IsSet()
}

// HasPasswordLastSet returns a boolean if a field has been set.
func (o *Account1) HasPasswordLastSet() bool {
	if o != nil && o.PasswordLastSet.IsSet() {
		return true
	}

	return false
}

// SetPasswordLastSet gets a reference to the given NullableTime and assigns it to the PasswordLastSet field.
func (o *Account1) SetPasswordLastSet(v time.Time) {
	o.PasswordLastSet.Set(&v)
}
// SetPasswordLastSetNil sets the value for PasswordLastSet to be an explicit nil
func (o *Account1) SetPasswordLastSetNil() {
	o.PasswordLastSet.Set(nil)
}

// UnsetPasswordLastSet ensures that no value is present for PasswordLastSet, not even an explicit nil
func (o *Account1) UnsetPasswordLastSet() {
	o.PasswordLastSet.Unset()
}

// GetEntitlementAttributes returns the EntitlementAttributes field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Account1) GetEntitlementAttributes() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}
	return o.EntitlementAttributes
}

// GetEntitlementAttributesOk returns a tuple with the EntitlementAttributes field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Account1) GetEntitlementAttributesOk() (map[string]interface{}, bool) {
	if o == nil || isNil(o.EntitlementAttributes) {
		return map[string]interface{}{}, false
	}
	return o.EntitlementAttributes, true
}

// HasEntitlementAttributes returns a boolean if a field has been set.
func (o *Account1) HasEntitlementAttributes() bool {
	if o != nil && isNil(o.EntitlementAttributes) {
		return true
	}

	return false
}

// SetEntitlementAttributes gets a reference to the given map[string]interface{} and assigns it to the EntitlementAttributes field.
func (o *Account1) SetEntitlementAttributes(v map[string]interface{}) {
	o.EntitlementAttributes = v
}

// GetCreated returns the Created field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Account1) GetCreated() time.Time {
	if o == nil || isNil(o.Created.Get()) {
		var ret time.Time
		return ret
	}
	return *o.Created.Get()
}

// GetCreatedOk returns a tuple with the Created field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Account1) GetCreatedOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return o.Created.Get(), o.Created.IsSet()
}

// HasCreated returns a boolean if a field has been set.
func (o *Account1) HasCreated() bool {
	if o != nil && o.Created.IsSet() {
		return true
	}

	return false
}

// SetCreated gets a reference to the given NullableTime and assigns it to the Created field.
func (o *Account1) SetCreated(v time.Time) {
	o.Created.Set(&v)
}
// SetCreatedNil sets the value for Created to be an explicit nil
func (o *Account1) SetCreatedNil() {
	o.Created.Set(nil)
}

// UnsetCreated ensures that no value is present for Created, not even an explicit nil
func (o *Account1) UnsetCreated() {
	o.Created.Unset()
}

// GetModified returns the Modified field value if set, zero value otherwise (both if not set or set to explicit null).
func (o *Account1) GetModified() time.Time {
	if o == nil || isNil(o.Modified.Get()) {
		var ret time.Time
		return ret
	}
	return *o.Modified.Get()
}

// GetModifiedOk returns a tuple with the Modified field value if set, nil otherwise
// and a boolean to check if the value has been set.
// NOTE: If the value is an explicit nil, `nil, true` will be returned
func (o *Account1) GetModifiedOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return o.Modified.Get(), o.Modified.IsSet()
}

// HasModified returns a boolean if a field has been set.
func (o *Account1) HasModified() bool {
	if o != nil && o.Modified.IsSet() {
		return true
	}

	return false
}

// SetModified gets a reference to the given NullableTime and assigns it to the Modified field.
func (o *Account1) SetModified(v time.Time) {
	o.Modified.Set(&v)
}
// SetModifiedNil sets the value for Modified to be an explicit nil
func (o *Account1) SetModifiedNil() {
	o.Modified.Set(nil)
}

// UnsetModified ensures that no value is present for Modified, not even an explicit nil
func (o *Account1) UnsetModified() {
	o.Modified.Unset()
}

// GetAttributes returns the Attributes field value if set, zero value otherwise.
func (o *Account1) GetAttributes() map[string]interface{} {
	if o == nil || isNil(o.Attributes) {
		var ret map[string]interface{}
		return ret
	}
	return o.Attributes
}

// GetAttributesOk returns a tuple with the Attributes field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetAttributesOk() (map[string]interface{}, bool) {
	if o == nil || isNil(o.Attributes) {
		return map[string]interface{}{}, false
	}
	return o.Attributes, true
}

// HasAttributes returns a boolean if a field has been set.
func (o *Account1) HasAttributes() bool {
	if o != nil && !isNil(o.Attributes) {
		return true
	}

	return false
}

// SetAttributes gets a reference to the given map[string]interface{} and assigns it to the Attributes field.
func (o *Account1) SetAttributes(v map[string]interface{}) {
	o.Attributes = v
}

// GetIdentity returns the Identity field value if set, zero value otherwise.
func (o *Account1) GetIdentity() DisplayReference {
	if o == nil || isNil(o.Identity) {
		var ret DisplayReference
		return ret
	}
	return *o.Identity
}

// GetIdentityOk returns a tuple with the Identity field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetIdentityOk() (*DisplayReference, bool) {
	if o == nil || isNil(o.Identity) {
		return nil, false
	}
	return o.Identity, true
}

// HasIdentity returns a boolean if a field has been set.
func (o *Account1) HasIdentity() bool {
	if o != nil && !isNil(o.Identity) {
		return true
	}

	return false
}

// SetIdentity gets a reference to the given DisplayReference and assigns it to the Identity field.
func (o *Account1) SetIdentity(v DisplayReference) {
	o.Identity = &v
}

// GetAccess returns the Access field value if set, zero value otherwise.
func (o *Account1) GetAccess() []Entitlement1 {
	if o == nil || isNil(o.Access) {
		var ret []Entitlement1
		return ret
	}
	return o.Access
}

// GetAccessOk returns a tuple with the Access field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetAccessOk() ([]Entitlement1, bool) {
	if o == nil || isNil(o.Access) {
		return nil, false
	}
	return o.Access, true
}

// HasAccess returns a boolean if a field has been set.
func (o *Account1) HasAccess() bool {
	if o != nil && !isNil(o.Access) {
		return true
	}

	return false
}

// SetAccess gets a reference to the given []Entitlement1 and assigns it to the Access field.
func (o *Account1) SetAccess(v []Entitlement1) {
	o.Access = v
}

// GetEntitlementCount returns the EntitlementCount field value if set, zero value otherwise.
func (o *Account1) GetEntitlementCount() int32 {
	if o == nil || isNil(o.EntitlementCount) {
		var ret int32
		return ret
	}
	return *o.EntitlementCount
}

// GetEntitlementCountOk returns a tuple with the EntitlementCount field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetEntitlementCountOk() (*int32, bool) {
	if o == nil || isNil(o.EntitlementCount) {
		return nil, false
	}
	return o.EntitlementCount, true
}

// HasEntitlementCount returns a boolean if a field has been set.
func (o *Account1) HasEntitlementCount() bool {
	if o != nil && !isNil(o.EntitlementCount) {
		return true
	}

	return false
}

// SetEntitlementCount gets a reference to the given int32 and assigns it to the EntitlementCount field.
func (o *Account1) SetEntitlementCount(v int32) {
	o.EntitlementCount = &v
}

// GetUncorrelated returns the Uncorrelated field value if set, zero value otherwise.
func (o *Account1) GetUncorrelated() bool {
	if o == nil || isNil(o.Uncorrelated) {
		var ret bool
		return ret
	}
	return *o.Uncorrelated
}

// GetUncorrelatedOk returns a tuple with the Uncorrelated field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetUncorrelatedOk() (*bool, bool) {
	if o == nil || isNil(o.Uncorrelated) {
		return nil, false
	}
	return o.Uncorrelated, true
}

// HasUncorrelated returns a boolean if a field has been set.
func (o *Account1) HasUncorrelated() bool {
	if o != nil && !isNil(o.Uncorrelated) {
		return true
	}

	return false
}

// SetUncorrelated gets a reference to the given bool and assigns it to the Uncorrelated field.
func (o *Account1) SetUncorrelated(v bool) {
	o.Uncorrelated = &v
}

// GetTags returns the Tags field value if set, zero value otherwise.
func (o *Account1) GetTags() []string {
	if o == nil || isNil(o.Tags) {
		var ret []string
		return ret
	}
	return o.Tags
}

// GetTagsOk returns a tuple with the Tags field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *Account1) GetTagsOk() ([]string, bool) {
	if o == nil || isNil(o.Tags) {
		return nil, false
	}
	return o.Tags, true
}

// HasTags returns a boolean if a field has been set.
func (o *Account1) HasTags() bool {
	if o != nil && !isNil(o.Tags) {
		return true
	}

	return false
}

// SetTags gets a reference to the given []string and assigns it to the Tags field.
func (o *Account1) SetTags(v []string) {
	o.Tags = v
}

func (o Account1) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["id"] = o.Id
	}
	if true {
		toSerialize["name"] = o.Name
	}
	if true {
		toSerialize["_type"] = o.Type
	}
	if !isNil(o.AccountId) {
		toSerialize["accountId"] = o.AccountId
	}
	if !isNil(o.Source) {
		toSerialize["source"] = o.Source
	}
	if !isNil(o.Disabled) {
		toSerialize["disabled"] = o.Disabled
	}
	if !isNil(o.Locked) {
		toSerialize["locked"] = o.Locked
	}
	if !isNil(o.Privileged) {
		toSerialize["privileged"] = o.Privileged
	}
	if !isNil(o.ManuallyCorrelated) {
		toSerialize["manuallyCorrelated"] = o.ManuallyCorrelated
	}
	if o.PasswordLastSet.IsSet() {
		toSerialize["passwordLastSet"] = o.PasswordLastSet.Get()
	}
	if o.EntitlementAttributes != nil {
		toSerialize["entitlementAttributes"] = o.EntitlementAttributes
	}
	if o.Created.IsSet() {
		toSerialize["created"] = o.Created.Get()
	}
	if o.Modified.IsSet() {
		toSerialize["modified"] = o.Modified.Get()
	}
	if !isNil(o.Attributes) {
		toSerialize["attributes"] = o.Attributes
	}
	if !isNil(o.Identity) {
		toSerialize["identity"] = o.Identity
	}
	if !isNil(o.Access) {
		toSerialize["access"] = o.Access
	}
	if !isNil(o.EntitlementCount) {
		toSerialize["entitlementCount"] = o.EntitlementCount
	}
	if !isNil(o.Uncorrelated) {
		toSerialize["uncorrelated"] = o.Uncorrelated
	}
	if !isNil(o.Tags) {
		toSerialize["tags"] = o.Tags
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *Account1) UnmarshalJSON(bytes []byte) (err error) {
	varAccount1 := _Account1{}

	if err = json.Unmarshal(bytes, &varAccount1); err == nil {
		*o = Account1(varAccount1)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "id")
		delete(additionalProperties, "name")
		delete(additionalProperties, "_type")
		delete(additionalProperties, "accountId")
		delete(additionalProperties, "source")
		delete(additionalProperties, "disabled")
		delete(additionalProperties, "locked")
		delete(additionalProperties, "privileged")
		delete(additionalProperties, "manuallyCorrelated")
		delete(additionalProperties, "passwordLastSet")
		delete(additionalProperties, "entitlementAttributes")
		delete(additionalProperties, "created")
		delete(additionalProperties, "modified")
		delete(additionalProperties, "attributes")
		delete(additionalProperties, "identity")
		delete(additionalProperties, "access")
		delete(additionalProperties, "entitlementCount")
		delete(additionalProperties, "uncorrelated")
		delete(additionalProperties, "tags")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableAccount1 struct {
	value *Account1
	isSet bool
}

func (v NullableAccount1) Get() *Account1 {
	return v.value
}

func (v *NullableAccount1) Set(val *Account1) {
	v.value = val
	v.isSet = true
}

func (v NullableAccount1) IsSet() bool {
	return v.isSet
}

func (v *NullableAccount1) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAccount1(val *Account1) *NullableAccount1 {
	return &NullableAccount1{value: val, isSet: true}
}

func (v NullableAccount1) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAccount1) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

