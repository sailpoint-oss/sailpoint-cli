/*
IdentityNow Beta API

Use these APIs to interact with the IdentityNow platform to achieve repeatable, automated processes with greater scalability. These APIs are in beta and are subject to change. We encourage you to join the SailPoint Developer Community forum at https://developer.sailpoint.com/discuss to connect with other developers using our APIs.

API version: 3.1.0-beta
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package sailpointbetasdk

import (
	"encoding/json"
	"time"
)

// VAClusterStatusChangeEvent struct for VAClusterStatusChangeEvent
type VAClusterStatusChangeEvent struct {
	// The date and time the status change occurred.
	Created time.Time `json:"created"`
	// The type of the object that initiated this event.
	Type map[string]interface{} `json:"type"`
	Application TriggerInputVAClusterStatusChangeEventApplication `json:"application"`
	HealthCheckResult TriggerInputVAClusterStatusChangeEventHealthCheckResult `json:"healthCheckResult"`
	PreviousHealthCheckResult TriggerInputVAClusterStatusChangeEventPreviousHealthCheckResult `json:"previousHealthCheckResult"`
	AdditionalProperties map[string]interface{}
}

type _VAClusterStatusChangeEvent VAClusterStatusChangeEvent

// NewVAClusterStatusChangeEvent instantiates a new VAClusterStatusChangeEvent object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewVAClusterStatusChangeEvent(created time.Time, type_ map[string]interface{}, application TriggerInputVAClusterStatusChangeEventApplication, healthCheckResult TriggerInputVAClusterStatusChangeEventHealthCheckResult, previousHealthCheckResult TriggerInputVAClusterStatusChangeEventPreviousHealthCheckResult) *VAClusterStatusChangeEvent {
	this := VAClusterStatusChangeEvent{}
	this.Created = created
	this.Type = type_
	this.Application = application
	this.HealthCheckResult = healthCheckResult
	this.PreviousHealthCheckResult = previousHealthCheckResult
	return &this
}

// NewVAClusterStatusChangeEventWithDefaults instantiates a new VAClusterStatusChangeEvent object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewVAClusterStatusChangeEventWithDefaults() *VAClusterStatusChangeEvent {
	this := VAClusterStatusChangeEvent{}
	return &this
}

// GetCreated returns the Created field value
func (o *VAClusterStatusChangeEvent) GetCreated() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.Created
}

// GetCreatedOk returns a tuple with the Created field value
// and a boolean to check if the value has been set.
func (o *VAClusterStatusChangeEvent) GetCreatedOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Created, true
}

// SetCreated sets field value
func (o *VAClusterStatusChangeEvent) SetCreated(v time.Time) {
	o.Created = v
}

// GetType returns the Type field value
func (o *VAClusterStatusChangeEvent) GetType() map[string]interface{} {
	if o == nil {
		var ret map[string]interface{}
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *VAClusterStatusChangeEvent) GetTypeOk() (map[string]interface{}, bool) {
	if o == nil {
		return map[string]interface{}{}, false
	}
	return o.Type, true
}

// SetType sets field value
func (o *VAClusterStatusChangeEvent) SetType(v map[string]interface{}) {
	o.Type = v
}

// GetApplication returns the Application field value
func (o *VAClusterStatusChangeEvent) GetApplication() TriggerInputVAClusterStatusChangeEventApplication {
	if o == nil {
		var ret TriggerInputVAClusterStatusChangeEventApplication
		return ret
	}

	return o.Application
}

// GetApplicationOk returns a tuple with the Application field value
// and a boolean to check if the value has been set.
func (o *VAClusterStatusChangeEvent) GetApplicationOk() (*TriggerInputVAClusterStatusChangeEventApplication, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Application, true
}

// SetApplication sets field value
func (o *VAClusterStatusChangeEvent) SetApplication(v TriggerInputVAClusterStatusChangeEventApplication) {
	o.Application = v
}

// GetHealthCheckResult returns the HealthCheckResult field value
func (o *VAClusterStatusChangeEvent) GetHealthCheckResult() TriggerInputVAClusterStatusChangeEventHealthCheckResult {
	if o == nil {
		var ret TriggerInputVAClusterStatusChangeEventHealthCheckResult
		return ret
	}

	return o.HealthCheckResult
}

// GetHealthCheckResultOk returns a tuple with the HealthCheckResult field value
// and a boolean to check if the value has been set.
func (o *VAClusterStatusChangeEvent) GetHealthCheckResultOk() (*TriggerInputVAClusterStatusChangeEventHealthCheckResult, bool) {
	if o == nil {
		return nil, false
	}
	return &o.HealthCheckResult, true
}

// SetHealthCheckResult sets field value
func (o *VAClusterStatusChangeEvent) SetHealthCheckResult(v TriggerInputVAClusterStatusChangeEventHealthCheckResult) {
	o.HealthCheckResult = v
}

// GetPreviousHealthCheckResult returns the PreviousHealthCheckResult field value
func (o *VAClusterStatusChangeEvent) GetPreviousHealthCheckResult() TriggerInputVAClusterStatusChangeEventPreviousHealthCheckResult {
	if o == nil {
		var ret TriggerInputVAClusterStatusChangeEventPreviousHealthCheckResult
		return ret
	}

	return o.PreviousHealthCheckResult
}

// GetPreviousHealthCheckResultOk returns a tuple with the PreviousHealthCheckResult field value
// and a boolean to check if the value has been set.
func (o *VAClusterStatusChangeEvent) GetPreviousHealthCheckResultOk() (*TriggerInputVAClusterStatusChangeEventPreviousHealthCheckResult, bool) {
	if o == nil {
		return nil, false
	}
	return &o.PreviousHealthCheckResult, true
}

// SetPreviousHealthCheckResult sets field value
func (o *VAClusterStatusChangeEvent) SetPreviousHealthCheckResult(v TriggerInputVAClusterStatusChangeEventPreviousHealthCheckResult) {
	o.PreviousHealthCheckResult = v
}

func (o VAClusterStatusChangeEvent) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["created"] = o.Created
	}
	if true {
		toSerialize["type"] = o.Type
	}
	if true {
		toSerialize["application"] = o.Application
	}
	if true {
		toSerialize["healthCheckResult"] = o.HealthCheckResult
	}
	if true {
		toSerialize["previousHealthCheckResult"] = o.PreviousHealthCheckResult
	}

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return json.Marshal(toSerialize)
}

func (o *VAClusterStatusChangeEvent) UnmarshalJSON(bytes []byte) (err error) {
	varVAClusterStatusChangeEvent := _VAClusterStatusChangeEvent{}

	if err = json.Unmarshal(bytes, &varVAClusterStatusChangeEvent); err == nil {
		*o = VAClusterStatusChangeEvent(varVAClusterStatusChangeEvent)
	}

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(bytes, &additionalProperties); err == nil {
		delete(additionalProperties, "created")
		delete(additionalProperties, "type")
		delete(additionalProperties, "application")
		delete(additionalProperties, "healthCheckResult")
		delete(additionalProperties, "previousHealthCheckResult")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableVAClusterStatusChangeEvent struct {
	value *VAClusterStatusChangeEvent
	isSet bool
}

func (v NullableVAClusterStatusChangeEvent) Get() *VAClusterStatusChangeEvent {
	return v.value
}

func (v *NullableVAClusterStatusChangeEvent) Set(val *VAClusterStatusChangeEvent) {
	v.value = val
	v.isSet = true
}

func (v NullableVAClusterStatusChangeEvent) IsSet() bool {
	return v.isSet
}

func (v *NullableVAClusterStatusChangeEvent) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableVAClusterStatusChangeEvent(val *VAClusterStatusChangeEvent) *NullableVAClusterStatusChangeEvent {
	return &NullableVAClusterStatusChangeEvent{value: val, isSet: true}
}

func (v NullableVAClusterStatusChangeEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableVAClusterStatusChangeEvent) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

