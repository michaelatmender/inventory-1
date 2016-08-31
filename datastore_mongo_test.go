// Copyright 2016 Mender Software AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.
package main_test

import (
	"errors"
	. "github.com/mendersoftware/inventory"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMongoGetDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoGetDevice in short mode.")
	}

	testCases := map[string]struct {
		InputID     DeviceID
		InputDevice *Device
		OutputError error
	}{
		"no device and no ID given": {
			InputID:     DeviceID(""),
			InputDevice: nil,
		},
		"device with given ID not exists": {
			InputID:     DeviceID("123"),
			InputDevice: nil,
		},
		"device with given ID exists, no error": {
			InputID: DeviceID("0002"),
			InputDevice: &Device{
				ID: DeviceID("0002"),
				Attributes: DeviceAttributes{
					"mac": DeviceAttribute{Name: "mac", Value: "0002-mac"},
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		session := db.Session()
		store := NewDataStoreMongoWithSession(session)

		if testCase.InputDevice != nil {
			session.DB(DbName).C(DbDevicesColl).Insert(testCase.InputDevice)
		}

		dbdev, err := store.GetDevice(testCase.InputID)

		if testCase.InputDevice != nil {
			assert.NotNil(t, dbdev, "expected to device of ID %s to be found", testCase.InputDevice.ID)
			assert.Equal(t, testCase.InputID, dbdev.ID)
		} else {
			assert.Nil(t, dbdev, "expected no device to be found")
		}

		assert.NoError(t, err, "expected no error")

		// Need to close all sessions to be able to call wipe at next test case
		session.Close()
	}
}

func TestMongoAddDevice(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoAddDevice in short mode.")
	}

	testCases := map[string]struct {
		InputDevice *Device
		OutputError error
	}{
		"no device given": {
			InputDevice: nil,
			OutputError: errors.New("failed to store device: error parsing element 0 of field documents :: caused by :: wrong type for '0' field, expected object, found 0: null"),
		},
		"valid device with one attribute, no error": {
			InputDevice: &Device{
				ID: DeviceID("0002"),
				Attributes: DeviceAttributes{
					"mac": DeviceAttribute{Name: "mac", Value: "0002-mac"},
				},
			},
			OutputError: nil,
		},
		"valid device with two attributes, no error": {
			InputDevice: &Device{
				ID: DeviceID("0003"),
				Attributes: DeviceAttributes{
					"mac": DeviceAttribute{Name: "mac", Value: "0002-mac"},
					"sn":  DeviceAttribute{Name: "sn", Value: "0002-sn"},
				},
			},
			OutputError: nil,
		},
		"valid device with attribute without value, no error": {
			InputDevice: &Device{
				ID: DeviceID("0004"),
				Attributes: DeviceAttributes{
					"mac": DeviceAttribute{Name: "mac"},
				},
			},
			OutputError: nil,
		},
		"valid device with array in attribute value, no error": {
			InputDevice: &Device{
				ID: DeviceID("0005"),
				Attributes: DeviceAttributes{
					"mac": DeviceAttribute{Name: "mac", Value: []interface{}{123, 456}},
				},
			},
			OutputError: nil,
		},
		"valid device without attributes, no error": {
			InputDevice: &Device{
				ID: DeviceID("0007"),
				Attributes: DeviceAttributes{
					"mac": DeviceAttribute{Name: "mac"},
				},
			},
			OutputError: nil,
		},
	}

	for name, testCase := range testCases {
		t.Logf("test case: %s", name)

		// Make sure we start test with empty database
		db.Wipe()

		session := db.Session()
		store := NewDataStoreMongoWithSession(session)

		err := store.AddDevice(testCase.InputDevice)

		if testCase.OutputError != nil {
			assert.EqualError(t, err, testCase.OutputError.Error())
		} else {
			assert.NoError(t, err, "expected no error inserting to data store")

			var dbdev *Device
			devsColl := session.DB(DbName).C(DbDevicesColl)
			err := devsColl.Find(nil).One(&dbdev)

			assert.NoError(t, err, "expected no error")

			assert.NotNil(t, dbdev, "expected to device of ID %s to be found", testCase.InputDevice.ID)

			assert.Equal(t, testCase.InputDevice.ID, dbdev.ID)
		}

		// Need to close all sessions to be able to call wipe at next test case
		session.Close()
	}
}

func TestNewDataStoreMongo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestNewDataStoreMongo in short mode.")
	}

	ds, err := NewDataStoreMongo("illegal url")

	assert.Nil(t, ds)
	assert.EqualError(t, err, "failed to open mgo session")
}

func TestMongoUpsertAttributes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestMongoUpsertAttributes in short mode.")
	}

	testCases := map[string]struct {
		devs []Device

		inDevId DeviceID
		inAttrs DeviceAttributes

		outAttrs DeviceAttributes
	}{
		"dev exists, attributes exist, update both attrs (descr + val)": {
			devs: []Device{
				{
					ID: DeviceID("0003"),
					Attributes: map[string]DeviceAttribute{
						"mac": {
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
						},
						"sn": {
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
						},
					},
				},
			},
			inDevId: DeviceID("0003"),
			inAttrs: map[string]DeviceAttribute{
				"mac": DeviceAttribute{
					Description: strPtr("mac description"),
					Value:       "0003-newmac",
				},
				"sn": DeviceAttribute{
					Description: strPtr("sn description"),
					Value:       "0003-newsn",
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"mac": DeviceAttribute{
					Description: strPtr("mac description"),
					Value:       "0003-newmac",
				},
				"sn": DeviceAttribute{
					Description: strPtr("sn description"),
					Value:       "0003-newsn",
				},
			},
		},
		"dev exists, attributes exist, update one attr (descr + val)": {
			devs: []Device{
				{
					ID: DeviceID("0003"),
					Attributes: map[string]DeviceAttribute{
						"mac": {
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
						},
						"sn": {
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
						},
					},
				},
			},
			inDevId: DeviceID("0003"),
			inAttrs: map[string]DeviceAttribute{
				"sn": DeviceAttribute{
					Description: strPtr("sn description"),
					Value:       "0003-newsn",
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"mac": DeviceAttribute{
					Description: strPtr("descr"),
					Value:       "0003-mac",
				},
				"sn": DeviceAttribute{
					Description: strPtr("sn description"),
					Value:       "0003-newsn",
				},
			},
		},

		"dev exists, attributes exist, update one attr (descr only)": {
			devs: []Device{
				{
					ID: DeviceID("0003"),
					Attributes: map[string]DeviceAttribute{
						"mac": {
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
						},
						"sn": {
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
						},
					},
				},
			},
			inDevId: DeviceID("0003"),
			inAttrs: map[string]DeviceAttribute{
				"sn": DeviceAttribute{
					Description: strPtr("sn description"),
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"mac": DeviceAttribute{
					Description: strPtr("descr"),
					Value:       "0003-mac",
				},
				"sn": DeviceAttribute{
					Description: strPtr("sn description"),
					Value:       "0003-sn",
				},
			},
		},
		"dev exists, attributes exist, update one attr (value only)": {
			devs: []Device{
				{
					ID: DeviceID("0003"),
					Attributes: map[string]DeviceAttribute{
						"mac": {
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
						},
						"sn": {
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
						},
					},
				},
			},
			inDevId: DeviceID("0003"),
			inAttrs: map[string]DeviceAttribute{
				"sn": DeviceAttribute{
					Value: "0003-newsn",
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"mac": DeviceAttribute{
					Description: strPtr("descr"),
					Value:       "0003-mac",
				},
				"sn": DeviceAttribute{
					Description: strPtr("descr"),
					Value:       "0003-newsn",
				},
			},
		},
		"dev exists, attributes exist, update one attr (value only, change type)": {
			devs: []Device{
				{
					ID: DeviceID("0003"),
					Attributes: map[string]DeviceAttribute{
						"mac": {
							Name:        "mac",
							Value:       "0003-mac",
							Description: strPtr("descr"),
						},
						"sn": {
							Name:        "sn",
							Value:       "0003-sn",
							Description: strPtr("descr"),
						},
					},
				},
			},
			inDevId: DeviceID("0003"),
			inAttrs: map[string]DeviceAttribute{
				"sn": DeviceAttribute{
					Value: []string{"0003-sn-1", "0003-sn-2"},
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"mac": DeviceAttribute{
					Description: strPtr("descr"),
					Value:       "0003-mac",
				},
				"sn": DeviceAttribute{
					Description: strPtr("descr"),
					//[]interface{} instead of []string - otherwise DeepEquals fails where it really shouldn't
					Value: []interface{}{"0003-sn-1", "0003-sn-2"},
				},
			},
		},
		"dev exists, no attributes exist, upsert new attrs (val + descr)": {
			devs: []Device{
				{
					ID: DeviceID("0003"),
				},
			},
			inDevId: DeviceID("0003"),
			inAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Value:       []string{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				"mac": DeviceAttribute{
					Value:       "0006-mac",
					Description: strPtr("mac addr"),
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Value:       []interface{}{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				"mac": DeviceAttribute{
					Value:       "0006-mac",
					Description: strPtr("mac addr"),
				},
			},
		},
		"dev doesn't exist, upsert new attr (descr + val)": {
			devs:    []Device{},
			inDevId: DeviceID("0099"),
			inAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Description: strPtr("ip addr array"),
					Value:       []string{"1.2.3.4", "1.2.3.5"},
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Description: strPtr("ip addr array"),
					Value:       []interface{}{"1.2.3.4", "1.2.3.5"},
				},
			},
		},
		"dev doesn't exist, upsert new attr (val only)": {
			devs:    []Device{},
			inDevId: DeviceID("0099"),
			inAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Value: []string{"1.2.3.4", "1.2.3.5"},
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Value: []interface{}{"1.2.3.4", "1.2.3.5"},
				},
			},
		},
		"dev doesn't exist, upsert with new attrs (val + descr)": {
			inDevId: DeviceID("0099"),
			inAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Value:       []string{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				"mac": DeviceAttribute{
					Value:       "0099-mac",
					Description: strPtr("mac addr"),
				},
			},

			outAttrs: map[string]DeviceAttribute{
				"ip": DeviceAttribute{
					Value:       []interface{}{"1.2.3.4", "1.2.3.5"},
					Description: strPtr("ip addr array"),
				},
				"mac": DeviceAttribute{
					Value:       "0099-mac",
					Description: strPtr("mac addr"),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Logf("%s", name)
		//setup
		db.Wipe()

		s := db.Session()

		for _, d := range tc.devs {
			err := s.DB(DbName).C(DbDevicesColl).Insert(d)
			assert.NoError(t, err, "failed to setup input data")
		}

		//test
		d := NewDataStoreMongoWithSession(s)
		err := d.UpsertAttributes(tc.inDevId, tc.inAttrs)
		assert.NoError(t, err, "UpsertAttributes failed")

		//get the device back
		var dev Device
		err = s.DB(DbName).C(DbDevicesColl).FindId(tc.inDevId).One(&dev)
		assert.NoError(t, err, "error getting device")

		if !compare(dev.Attributes, tc.outAttrs) {
			t.Errorf("attributes mismatch, have: %v\nwant: %v", dev.Attributes, tc.outAttrs)
		}

		s.Close()
	}

	//wipe(d)
}

func strPtr(s string) *string {
	return &s
}

func compare(a, b DeviceAttributes) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		vb := b[k]

		if !reflect.DeepEqual(va.Value, vb.Value) {
			return false
		}

		if va.Description == nil &&
			vb.Description == nil {
			return true
		}

		if va.Description != nil &&
			vb.Description != nil &&
			*va.Description == *vb.Description {
			return true
		} else {
			return false
		}
	}

	return true
}