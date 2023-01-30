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

// IdentitySyncJob struct for IdentitySyncJob
type IdentitySyncJob struct {
	// Job ID.
	Id string `json:"id"`
	// The job status.
	Status string `json:"status"`
	Payload IdentitySyncPayload `json:"payload"`
	AdditionalProperties map[string]interface{}
}

type _IdentitySyncJob IdentitySyncJob

// NewIdentitySyncJob instantiates a new IdentitySyncJob object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewIdentitySyncJob(id string, status string, payload IdentitySyncPayload) *IdentitySyncJob {
	this := IdentitySyncJob{}
	this.Id = id
	this.Status = status
	this.Payload = payload
	return &this
}

// NewIdentitySyncJobWithDefaults instantiates a new IdentitySyncJob object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewIdentitySyncJobWithDefaults() *IdentitySyncJob {
	this := IdentitySyncJob{}
	return &this
}

// GetId returns the Id field value
func (o *IdentitySyncJob) GetId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Id
}

// GetIdOk returns a tuple with the Id field value
// and a boolean to check if the value has been set.
func (o *IdentitySyncJob) GetIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Id, true
}

// SetId sets field value
func (o *IdentitySyncJob) SetId(v string) {
	o.Id = v
}

// GetStatus returns the Status field value
func (o *IdentitySyncJob) GetStatus() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Status
}

// GetStatusOk returns a tuple with the Status field value
// and a boolean to check if the value has been set.
func (o *IdentitySyncJob) GetStatusOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Status, true
}

// SetStatus sets field value
func (o *IdentitySyncJob) SetStatus(v string) {
	o.Status = v
}

// GetPayload returns the Payload field value
func (o *IdentitySyncJob) GetPayload() IdentitySyncPayload {
	if o == nil {
		var ret IdentitySyncPayload
		return ret
	}

	return o.Payload
}

// GetPayloadOk returns a tuple with the Payload field value
// and a boolean to check if the value has been set.
func (o *IdentitySyncJob) GetPayloadOk() (*IdentitySyncPayload, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Payload, true
}

// SetPayload sets field value
func (o *IdentitySyncJob) SetPayload(v IdentitySyncPayload) {
	o.Payload = v
}

func (o IdentitySyncJob) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["id"] = o.Id
	}
	if true {
		toSerialize["status"] = o.Status
	}
	if true {
		toSerialize["payload"] = o.Payload
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *IdentitySyncJob) UnmarshalJSON(bytes []byte) (err error) {
	varIdentitySyncJob := _IdentitySyncJob{}

	if err = json.Unmarshal(bytes, &varIdentitySyncJob); err == nil {
		*o = IdentitySyncJob(varIdentitySyncJob)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "id")
		delete(additionalProperties, "status")
		delete(additionalProperties, "payload")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableIdentitySyncJob struct {
	value *IdentitySyncJob
	isSet bool
}

func (v NullableIdentitySyncJob) Get() *IdentitySyncJob {
	return v.value
}

func (v *NullableIdentitySyncJob) Set(val *IdentitySyncJob) {
	v.value = val
	v.isSet = true
}

func (v NullableIdentitySyncJob) IsSet() bool {
	return v.isSet
}

func (v *NullableIdentitySyncJob) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableIdentitySyncJob(val *IdentitySyncJob) *NullableIdentitySyncJob {
	return &NullableIdentitySyncJob{value: val, isSet: true}
}

func (v NullableIdentitySyncJob) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableIdentitySyncJob) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

