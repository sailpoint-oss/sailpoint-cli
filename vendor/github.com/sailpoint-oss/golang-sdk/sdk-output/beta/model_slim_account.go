/*
IdentityNow Beta API

Use these APIs to interact with the IdentityNow platform to achieve repeatable, automated processes with greater scalability. These APIs are in beta and are subject to change. We encourage you to join the SailPoint Developer Community forum at https://developer.sailpoint.com/discuss to connect with other developers using our APIs.

API version: 3.1.0-beta
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package sailpointbetasdk

import (
	"time"
	"encoding/json"
)

// SlimAccount struct for SlimAccount
type SlimAccount struct {
	// System-generated unique ID of the Object
	Id *string `json:"id,omitempty"`
	// Name of the Object
	Name string `json:"name"`
	// Creation date of the Object
	Created *time.Time `json:"created,omitempty"`
	// Last modification date of the Object
	Modified *time.Time `json:"modified,omitempty"`
	// Unique ID from the owning source
	Uuid *string `json:"uuid,omitempty"`
	// The native identifier of the account
	NativeIdentity *string `json:"nativeIdentity,omitempty"`
	// The description for the account
	Description *string `json:"description,omitempty"`
	// Whether the account is disabled
	Disabled *bool `json:"disabled,omitempty"`
	// Whether the account is locked
	Locked *bool `json:"locked,omitempty"`
	// Whether the account was manually correlated
	ManuallyCorrelated *bool `json:"manuallyCorrelated,omitempty"`
	// Whether the account has any entitlements associated with it
	HasEntitlements *bool `json:"hasEntitlements,omitempty"`
	// The ID of the source for which this account belongs
	SourceId *string `json:"sourceId,omitempty"`
	// The ID of the identity for which this account is correlated to if not uncorrelated
	IdentityId *string `json:"identityId,omitempty"`
	// A map containing attributes associated with the account
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _SlimAccount SlimAccount

// NewSlimAccount instantiates a new SlimAccount object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSlimAccount(name string) *SlimAccount {
	this := SlimAccount{}
	this.Name = name
	return &this
}

// NewSlimAccountWithDefaults instantiates a new SlimAccount object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSlimAccountWithDefaults() *SlimAccount {
	this := SlimAccount{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *SlimAccount) GetId() string {
	if o == nil || isNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetIdOk() (*string, bool) {
	if o == nil || isNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *SlimAccount) HasId() bool {
	if o != nil && !isNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *SlimAccount) SetId(v string) {
	o.Id = &v
}

// GetName returns the Name field value
func (o *SlimAccount) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *SlimAccount) SetName(v string) {
	o.Name = v
}

// GetCreated returns the Created field value if set, zero value otherwise.
func (o *SlimAccount) GetCreated() time.Time {
	if o == nil || isNil(o.Created) {
		var ret time.Time
		return ret
	}
	return *o.Created
}

// GetCreatedOk returns a tuple with the Created field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetCreatedOk() (*time.Time, bool) {
	if o == nil || isNil(o.Created) {
		return nil, false
	}
	return o.Created, true
}

// HasCreated returns a boolean if a field has been set.
func (o *SlimAccount) HasCreated() bool {
	if o != nil && !isNil(o.Created) {
		return true
	}

	return false
}

// SetCreated gets a reference to the given time.Time and assigns it to the Created field.
func (o *SlimAccount) SetCreated(v time.Time) {
	o.Created = &v
}

// GetModified returns the Modified field value if set, zero value otherwise.
func (o *SlimAccount) GetModified() time.Time {
	if o == nil || isNil(o.Modified) {
		var ret time.Time
		return ret
	}
	return *o.Modified
}

// GetModifiedOk returns a tuple with the Modified field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetModifiedOk() (*time.Time, bool) {
	if o == nil || isNil(o.Modified) {
		return nil, false
	}
	return o.Modified, true
}

// HasModified returns a boolean if a field has been set.
func (o *SlimAccount) HasModified() bool {
	if o != nil && !isNil(o.Modified) {
		return true
	}

	return false
}

// SetModified gets a reference to the given time.Time and assigns it to the Modified field.
func (o *SlimAccount) SetModified(v time.Time) {
	o.Modified = &v
}

// GetUuid returns the Uuid field value if set, zero value otherwise.
func (o *SlimAccount) GetUuid() string {
	if o == nil || isNil(o.Uuid) {
		var ret string
		return ret
	}
	return *o.Uuid
}

// GetUuidOk returns a tuple with the Uuid field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetUuidOk() (*string, bool) {
	if o == nil || isNil(o.Uuid) {
		return nil, false
	}
	return o.Uuid, true
}

// HasUuid returns a boolean if a field has been set.
func (o *SlimAccount) HasUuid() bool {
	if o != nil && !isNil(o.Uuid) {
		return true
	}

	return false
}

// SetUuid gets a reference to the given string and assigns it to the Uuid field.
func (o *SlimAccount) SetUuid(v string) {
	o.Uuid = &v
}

// GetNativeIdentity returns the NativeIdentity field value if set, zero value otherwise.
func (o *SlimAccount) GetNativeIdentity() string {
	if o == nil || isNil(o.NativeIdentity) {
		var ret string
		return ret
	}
	return *o.NativeIdentity
}

// GetNativeIdentityOk returns a tuple with the NativeIdentity field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetNativeIdentityOk() (*string, bool) {
	if o == nil || isNil(o.NativeIdentity) {
		return nil, false
	}
	return o.NativeIdentity, true
}

// HasNativeIdentity returns a boolean if a field has been set.
func (o *SlimAccount) HasNativeIdentity() bool {
	if o != nil && !isNil(o.NativeIdentity) {
		return true
	}

	return false
}

// SetNativeIdentity gets a reference to the given string and assigns it to the NativeIdentity field.
func (o *SlimAccount) SetNativeIdentity(v string) {
	o.NativeIdentity = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *SlimAccount) GetDescription() string {
	if o == nil || isNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetDescriptionOk() (*string, bool) {
	if o == nil || isNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *SlimAccount) HasDescription() bool {
	if o != nil && !isNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *SlimAccount) SetDescription(v string) {
	o.Description = &v
}

// GetDisabled returns the Disabled field value if set, zero value otherwise.
func (o *SlimAccount) GetDisabled() bool {
	if o == nil || isNil(o.Disabled) {
		var ret bool
		return ret
	}
	return *o.Disabled
}

// GetDisabledOk returns a tuple with the Disabled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetDisabledOk() (*bool, bool) {
	if o == nil || isNil(o.Disabled) {
		return nil, false
	}
	return o.Disabled, true
}

// HasDisabled returns a boolean if a field has been set.
func (o *SlimAccount) HasDisabled() bool {
	if o != nil && !isNil(o.Disabled) {
		return true
	}

	return false
}

// SetDisabled gets a reference to the given bool and assigns it to the Disabled field.
func (o *SlimAccount) SetDisabled(v bool) {
	o.Disabled = &v
}

// GetLocked returns the Locked field value if set, zero value otherwise.
func (o *SlimAccount) GetLocked() bool {
	if o == nil || isNil(o.Locked) {
		var ret bool
		return ret
	}
	return *o.Locked
}

// GetLockedOk returns a tuple with the Locked field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetLockedOk() (*bool, bool) {
	if o == nil || isNil(o.Locked) {
		return nil, false
	}
	return o.Locked, true
}

// HasLocked returns a boolean if a field has been set.
func (o *SlimAccount) HasLocked() bool {
	if o != nil && !isNil(o.Locked) {
		return true
	}

	return false
}

// SetLocked gets a reference to the given bool and assigns it to the Locked field.
func (o *SlimAccount) SetLocked(v bool) {
	o.Locked = &v
}

// GetManuallyCorrelated returns the ManuallyCorrelated field value if set, zero value otherwise.
func (o *SlimAccount) GetManuallyCorrelated() bool {
	if o == nil || isNil(o.ManuallyCorrelated) {
		var ret bool
		return ret
	}
	return *o.ManuallyCorrelated
}

// GetManuallyCorrelatedOk returns a tuple with the ManuallyCorrelated field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetManuallyCorrelatedOk() (*bool, bool) {
	if o == nil || isNil(o.ManuallyCorrelated) {
		return nil, false
	}
	return o.ManuallyCorrelated, true
}

// HasManuallyCorrelated returns a boolean if a field has been set.
func (o *SlimAccount) HasManuallyCorrelated() bool {
	if o != nil && !isNil(o.ManuallyCorrelated) {
		return true
	}

	return false
}

// SetManuallyCorrelated gets a reference to the given bool and assigns it to the ManuallyCorrelated field.
func (o *SlimAccount) SetManuallyCorrelated(v bool) {
	o.ManuallyCorrelated = &v
}

// GetHasEntitlements returns the HasEntitlements field value if set, zero value otherwise.
func (o *SlimAccount) GetHasEntitlements() bool {
	if o == nil || isNil(o.HasEntitlements) {
		var ret bool
		return ret
	}
	return *o.HasEntitlements
}

// GetHasEntitlementsOk returns a tuple with the HasEntitlements field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetHasEntitlementsOk() (*bool, bool) {
	if o == nil || isNil(o.HasEntitlements) {
		return nil, false
	}
	return o.HasEntitlements, true
}

// HasHasEntitlements returns a boolean if a field has been set.
func (o *SlimAccount) HasHasEntitlements() bool {
	if o != nil && !isNil(o.HasEntitlements) {
		return true
	}

	return false
}

// SetHasEntitlements gets a reference to the given bool and assigns it to the HasEntitlements field.
func (o *SlimAccount) SetHasEntitlements(v bool) {
	o.HasEntitlements = &v
}

// GetSourceId returns the SourceId field value if set, zero value otherwise.
func (o *SlimAccount) GetSourceId() string {
	if o == nil || isNil(o.SourceId) {
		var ret string
		return ret
	}
	return *o.SourceId
}

// GetSourceIdOk returns a tuple with the SourceId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetSourceIdOk() (*string, bool) {
	if o == nil || isNil(o.SourceId) {
		return nil, false
	}
	return o.SourceId, true
}

// HasSourceId returns a boolean if a field has been set.
func (o *SlimAccount) HasSourceId() bool {
	if o != nil && !isNil(o.SourceId) {
		return true
	}

	return false
}

// SetSourceId gets a reference to the given string and assigns it to the SourceId field.
func (o *SlimAccount) SetSourceId(v string) {
	o.SourceId = &v
}

// GetIdentityId returns the IdentityId field value if set, zero value otherwise.
func (o *SlimAccount) GetIdentityId() string {
	if o == nil || isNil(o.IdentityId) {
		var ret string
		return ret
	}
	return *o.IdentityId
}

// GetIdentityIdOk returns a tuple with the IdentityId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetIdentityIdOk() (*string, bool) {
	if o == nil || isNil(o.IdentityId) {
		return nil, false
	}
	return o.IdentityId, true
}

// HasIdentityId returns a boolean if a field has been set.
func (o *SlimAccount) HasIdentityId() bool {
	if o != nil && !isNil(o.IdentityId) {
		return true
	}

	return false
}

// SetIdentityId gets a reference to the given string and assigns it to the IdentityId field.
func (o *SlimAccount) SetIdentityId(v string) {
	o.IdentityId = &v
}

// GetAttributes returns the Attributes field value if set, zero value otherwise.
func (o *SlimAccount) GetAttributes() map[string]interface{} {
	if o == nil || isNil(o.Attributes) {
		var ret map[string]interface{}
		return ret
	}
	return o.Attributes
}

// GetAttributesOk returns a tuple with the Attributes field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *SlimAccount) GetAttributesOk() (map[string]interface{}, bool) {
	if o == nil || isNil(o.Attributes) {
		return map[string]interface{}{}, false
	}
	return o.Attributes, true
}

// HasAttributes returns a boolean if a field has been set.
func (o *SlimAccount) HasAttributes() bool {
	if o != nil && !isNil(o.Attributes) {
		return true
	}

	return false
}

// SetAttributes gets a reference to the given map[string]interface{} and assigns it to the Attributes field.
func (o *SlimAccount) SetAttributes(v map[string]interface{}) {
	o.Attributes = v
}

func (o SlimAccount) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	if true {
		toSerialize["name"] = o.Name
	}
	if !isNil(o.Created) {
		toSerialize["created"] = o.Created
	}
	if !isNil(o.Modified) {
		toSerialize["modified"] = o.Modified
	}
	if !isNil(o.Uuid) {
		toSerialize["uuid"] = o.Uuid
	}
	if !isNil(o.NativeIdentity) {
		toSerialize["nativeIdentity"] = o.NativeIdentity
	}
	if !isNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !isNil(o.Disabled) {
		toSerialize["disabled"] = o.Disabled
	}
	if !isNil(o.Locked) {
		toSerialize["locked"] = o.Locked
	}
	if !isNil(o.ManuallyCorrelated) {
		toSerialize["manuallyCorrelated"] = o.ManuallyCorrelated
	}
	if !isNil(o.HasEntitlements) {
		toSerialize["hasEntitlements"] = o.HasEntitlements
	}
	if !isNil(o.SourceId) {
		toSerialize["sourceId"] = o.SourceId
	}
	if !isNil(o.IdentityId) {
		toSerialize["identityId"] = o.IdentityId
	}
	if !isNil(o.Attributes) {
		toSerialize["attributes"] = o.Attributes
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *SlimAccount) UnmarshalJSON(bytes []byte) (err error) {
	varSlimAccount := _SlimAccount{}

	if err = json.Unmarshal(bytes, &varSlimAccount); err == nil {
		*o = SlimAccount(varSlimAccount)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "id")
		delete(additionalProperties, "name")
		delete(additionalProperties, "created")
		delete(additionalProperties, "modified")
		delete(additionalProperties, "uuid")
		delete(additionalProperties, "nativeIdentity")
		delete(additionalProperties, "description")
		delete(additionalProperties, "disabled")
		delete(additionalProperties, "locked")
		delete(additionalProperties, "manuallyCorrelated")
		delete(additionalProperties, "hasEntitlements")
		delete(additionalProperties, "sourceId")
		delete(additionalProperties, "identityId")
		delete(additionalProperties, "attributes")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableSlimAccount struct {
	value *SlimAccount
	isSet bool
}

func (v NullableSlimAccount) Get() *SlimAccount {
	return v.value
}

func (v *NullableSlimAccount) Set(val *SlimAccount) {
	v.value = val
	v.isSet = true
}

func (v NullableSlimAccount) IsSet() bool {
	return v.isSet
}

func (v *NullableSlimAccount) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSlimAccount(val *SlimAccount) *NullableSlimAccount {
	return &NullableSlimAccount{value: val, isSet: true}
}

func (v NullableSlimAccount) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSlimAccount) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

