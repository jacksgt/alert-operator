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

// checks if the AlertGroup type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AlertGroup{}

// AlertGroup struct for AlertGroup
type AlertGroup struct {
	Labels map[string]string `json:"labels"`
	Receiver Receiver `json:"receiver"`
	Alerts []GettableAlert `json:"alerts"`
}

type _AlertGroup AlertGroup

// NewAlertGroup instantiates a new AlertGroup object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAlertGroup(labels map[string]string, receiver Receiver, alerts []GettableAlert) *AlertGroup {
	this := AlertGroup{}
	this.Labels = labels
	this.Receiver = receiver
	this.Alerts = alerts
	return &this
}

// NewAlertGroupWithDefaults instantiates a new AlertGroup object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAlertGroupWithDefaults() *AlertGroup {
	this := AlertGroup{}
	return &this
}

// GetLabels returns the Labels field value
func (o *AlertGroup) GetLabels() map[string]string {
	if o == nil {
		var ret map[string]string
		return ret
	}

	return o.Labels
}

// GetLabelsOk returns a tuple with the Labels field value
// and a boolean to check if the value has been set.
func (o *AlertGroup) GetLabelsOk() (*map[string]string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Labels, true
}

// SetLabels sets field value
func (o *AlertGroup) SetLabels(v map[string]string) {
	o.Labels = v
}

// GetReceiver returns the Receiver field value
func (o *AlertGroup) GetReceiver() Receiver {
	if o == nil {
		var ret Receiver
		return ret
	}

	return o.Receiver
}

// GetReceiverOk returns a tuple with the Receiver field value
// and a boolean to check if the value has been set.
func (o *AlertGroup) GetReceiverOk() (*Receiver, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Receiver, true
}

// SetReceiver sets field value
func (o *AlertGroup) SetReceiver(v Receiver) {
	o.Receiver = v
}

// GetAlerts returns the Alerts field value
func (o *AlertGroup) GetAlerts() []GettableAlert {
	if o == nil {
		var ret []GettableAlert
		return ret
	}

	return o.Alerts
}

// GetAlertsOk returns a tuple with the Alerts field value
// and a boolean to check if the value has been set.
func (o *AlertGroup) GetAlertsOk() ([]GettableAlert, bool) {
	if o == nil {
		return nil, false
	}
	return o.Alerts, true
}

// SetAlerts sets field value
func (o *AlertGroup) SetAlerts(v []GettableAlert) {
	o.Alerts = v
}

func (o AlertGroup) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AlertGroup) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["labels"] = o.Labels
	toSerialize["receiver"] = o.Receiver
	toSerialize["alerts"] = o.Alerts
	return toSerialize, nil
}

func (o *AlertGroup) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"labels",
		"receiver",
		"alerts",
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

	varAlertGroup := _AlertGroup{}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&varAlertGroup)

	if err != nil {
		return err
	}

	*o = AlertGroup(varAlertGroup)

	return err
}

type NullableAlertGroup struct {
	value *AlertGroup
	isSet bool
}

func (v NullableAlertGroup) Get() *AlertGroup {
	return v.value
}

func (v *NullableAlertGroup) Set(val *AlertGroup) {
	v.value = val
	v.isSet = true
}

func (v NullableAlertGroup) IsSet() bool {
	return v.isSet
}

func (v *NullableAlertGroup) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAlertGroup(val *AlertGroup) *NullableAlertGroup {
	return &NullableAlertGroup{value: val, isSet: true}
}

func (v NullableAlertGroup) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAlertGroup) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


