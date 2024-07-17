/*
Alertmanager API

API of the Prometheus Alertmanager (https://github.com/prometheus/alertmanager)

API version: 0.0.1
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package alertmanagerapi

import (
	"encoding/json"
	"bytes"
	"fmt"
)

// checks if the AlertStatus type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AlertStatus{}

// AlertStatus struct for AlertStatus
type AlertStatus struct {
	State string `json:"state"`
	SilencedBy []string `json:"silencedBy"`
	InhibitedBy []string `json:"inhibitedBy"`
}

type _AlertStatus AlertStatus

// NewAlertStatus instantiates a new AlertStatus object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAlertStatus(state string, silencedBy []string, inhibitedBy []string) *AlertStatus {
	this := AlertStatus{}
	this.State = state
	this.SilencedBy = silencedBy
	this.InhibitedBy = inhibitedBy
	return &this
}

// NewAlertStatusWithDefaults instantiates a new AlertStatus object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAlertStatusWithDefaults() *AlertStatus {
	this := AlertStatus{}
	return &this
}

// GetState returns the State field value
func (o *AlertStatus) GetState() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.State
}

// GetStateOk returns a tuple with the State field value
// and a boolean to check if the value has been set.
func (o *AlertStatus) GetStateOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.State, true
}

// SetState sets field value
func (o *AlertStatus) SetState(v string) {
	o.State = v
}

// GetSilencedBy returns the SilencedBy field value
func (o *AlertStatus) GetSilencedBy() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.SilencedBy
}

// GetSilencedByOk returns a tuple with the SilencedBy field value
// and a boolean to check if the value has been set.
func (o *AlertStatus) GetSilencedByOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.SilencedBy, true
}

// SetSilencedBy sets field value
func (o *AlertStatus) SetSilencedBy(v []string) {
	o.SilencedBy = v
}

// GetInhibitedBy returns the InhibitedBy field value
func (o *AlertStatus) GetInhibitedBy() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.InhibitedBy
}

// GetInhibitedByOk returns a tuple with the InhibitedBy field value
// and a boolean to check if the value has been set.
func (o *AlertStatus) GetInhibitedByOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.InhibitedBy, true
}

// SetInhibitedBy sets field value
func (o *AlertStatus) SetInhibitedBy(v []string) {
	o.InhibitedBy = v
}

func (o AlertStatus) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AlertStatus) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["state"] = o.State
	toSerialize["silencedBy"] = o.SilencedBy
	toSerialize["inhibitedBy"] = o.InhibitedBy
	return toSerialize, nil
}

func (o *AlertStatus) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"state",
		"silencedBy",
		"inhibitedBy",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err;
	}

	for _, requiredProperty := range(requiredProperties) {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varAlertStatus := _AlertStatus{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varAlertStatus)

	if err != nil {
		return err
	}

	*o = AlertStatus(varAlertStatus)

	return err
}

type NullableAlertStatus struct {
	value *AlertStatus
	isSet bool
}

func (v NullableAlertStatus) Get() *AlertStatus {
	return v.value
}

func (v *NullableAlertStatus) Set(val *AlertStatus) {
	v.value = val
	v.isSet = true
}

func (v NullableAlertStatus) IsSet() bool {
	return v.isSet
}

func (v *NullableAlertStatus) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAlertStatus(val *AlertStatus) *NullableAlertStatus {
	return &NullableAlertStatus{value: val, isSet: true}
}

func (v NullableAlertStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAlertStatus) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


