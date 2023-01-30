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

// TestWorkflowRequest struct for TestWorkflowRequest
type TestWorkflowRequest struct {
	// The test input for the workflow.
	Input map[string]interface{} `json:"input"`
	AdditionalProperties map[string]interface{}
}

type _TestWorkflowRequest TestWorkflowRequest

// NewTestWorkflowRequest instantiates a new TestWorkflowRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTestWorkflowRequest(input map[string]interface{}) *TestWorkflowRequest {
	this := TestWorkflowRequest{}
	this.Input = input
	return &this
}

// NewTestWorkflowRequestWithDefaults instantiates a new TestWorkflowRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTestWorkflowRequestWithDefaults() *TestWorkflowRequest {
	this := TestWorkflowRequest{}
	return &this
}

// GetInput returns the Input field value
func (o *TestWorkflowRequest) GetInput() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Input
}

// GetInputOk returns a tuple with the Input field value
// and a boolean to check if the value has been set.
func (o *TestWorkflowRequest) GetInputOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Input, true
}

// SetInput sets field value
func (o *TestWorkflowRequest) SetInput(v map[string]interface{}) {
	o.Input = v
}

func (o TestWorkflowRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["input"] = o.Input
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *TestWorkflowRequest) UnmarshalJSON(bytes []byte) (err error) {
	varTestWorkflowRequest := _TestWorkflowRequest{}

	if err = json.Unmarshal(bytes, &varTestWorkflowRequest); err == nil {
		*o = TestWorkflowRequest(varTestWorkflowRequest)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "input")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableTestWorkflowRequest struct {
	value *TestWorkflowRequest
	isSet bool
}

func (v NullableTestWorkflowRequest) Get() *TestWorkflowRequest {
	return v.value
}

func (v *NullableTestWorkflowRequest) Set(val *TestWorkflowRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableTestWorkflowRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableTestWorkflowRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTestWorkflowRequest(val *TestWorkflowRequest) *NullableTestWorkflowRequest {
	return &NullableTestWorkflowRequest{value: val, isSet: true}
}

func (v NullableTestWorkflowRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTestWorkflowRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

