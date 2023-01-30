/*
IdentityNow Beta API

Use these APIs to interact with the IdentityNow platform to achieve repeatable, automated processes with greater scalability. These APIs are in beta and are subject to change. We encourage you to join the SailPoint Developer Community forum at https://developer.sailpoint.com/discuss to connect with other developers using our APIs.

API version: 3.1.0-beta
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package sailpointbetasdk

import (
	"encoding/json"
)

// NonEmployeeApprovalItemAllOf struct for NonEmployeeApprovalItemAllOf
type NonEmployeeApprovalItemAllOf struct {
	NonEmployeeRequest *NonEmployeeRequestLite `json:"nonEmployeeRequest,omitempty"`
	AdditionalProperties map[string]interface{}
}

type _NonEmployeeApprovalItemAllOf NonEmployeeApprovalItemAllOf

// NewNonEmployeeApprovalItemAllOf instantiates a new NonEmployeeApprovalItemAllOf object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewNonEmployeeApprovalItemAllOf() *NonEmployeeApprovalItemAllOf {
	this := NonEmployeeApprovalItemAllOf{}
	return &this
}

// NewNonEmployeeApprovalItemAllOfWithDefaults instantiates a new NonEmployeeApprovalItemAllOf object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewNonEmployeeApprovalItemAllOfWithDefaults() *NonEmployeeApprovalItemAllOf {
	this := NonEmployeeApprovalItemAllOf{}
	return &this
}

// GetNonEmployeeRequest returns the NonEmployeeRequest field value if set, zero value otherwise.
func (o *NonEmployeeApprovalItemAllOf) GetNonEmployeeRequest() NonEmployeeRequestLite {
	if o == nil || isNil(o.NonEmployeeRequest) {
		var ret NonEmployeeRequestLite
		return ret
	}
	return *o.NonEmployeeRequest
}

// GetNonEmployeeRequestOk returns a tuple with the NonEmployeeRequest field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *NonEmployeeApprovalItemAllOf) GetNonEmployeeRequestOk() (*NonEmployeeRequestLite, bool) {
	if o == nil || isNil(o.NonEmployeeRequest) {
		return nil, false
	}
	return o.NonEmployeeRequest, true
}

// HasNonEmployeeRequest returns a boolean if a field has been set.
func (o *NonEmployeeApprovalItemAllOf) HasNonEmployeeRequest() bool {
	if o != nil && !isNil(o.NonEmployeeRequest) {
		return true
	}

	return false
}

// SetNonEmployeeRequest gets a reference to the given NonEmployeeRequestLite and assigns it to the NonEmployeeRequest field.
func (o *NonEmployeeApprovalItemAllOf) SetNonEmployeeRequest(v NonEmployeeRequestLite) {
	o.NonEmployeeRequest = &v
}

func (o NonEmployeeApprovalItemAllOf) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if !isNil(o.NonEmployeeRequest) {
		toSerialize["nonEmployeeRequest"] = o.NonEmployeeRequest
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *NonEmployeeApprovalItemAllOf) UnmarshalJSON(bytes []byte) (err error) {
	varNonEmployeeApprovalItemAllOf := _NonEmployeeApprovalItemAllOf{}

	if err = json.Unmarshal(bytes, &varNonEmployeeApprovalItemAllOf); err == nil {
		*o = NonEmployeeApprovalItemAllOf(varNonEmployeeApprovalItemAllOf)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "nonEmployeeRequest")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableNonEmployeeApprovalItemAllOf struct {
	value *NonEmployeeApprovalItemAllOf
	isSet bool
}

func (v NullableNonEmployeeApprovalItemAllOf) Get() *NonEmployeeApprovalItemAllOf {
	return v.value
}

func (v *NullableNonEmployeeApprovalItemAllOf) Set(val *NonEmployeeApprovalItemAllOf) {
	v.value = val
	v.isSet = true
}

func (v NullableNonEmployeeApprovalItemAllOf) IsSet() bool {
	return v.isSet
}

func (v *NullableNonEmployeeApprovalItemAllOf) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableNonEmployeeApprovalItemAllOf(val *NonEmployeeApprovalItemAllOf) *NullableNonEmployeeApprovalItemAllOf {
	return &NullableNonEmployeeApprovalItemAllOf{value: val, isSet: true}
}

func (v NullableNonEmployeeApprovalItemAllOf) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableNonEmployeeApprovalItemAllOf) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

