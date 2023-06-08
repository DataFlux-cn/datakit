// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.
// Some code modified from project Datadog (https://www.datadoghq.com/).

package snmputil

import (
	"fmt"
	"testing"

	"github.com/gosnmp/gosnmp"
	"github.com/stretchr/testify/assert"
)

//------------------------------------------------------------------------------

func Test_getValueFromPDU(t *testing.T) {
	tests := []struct {
		caseName          string
		pduVariable       gosnmp.SnmpPDU
		expectedName      string
		expectedSnmpValue ResultValue
		expectedErr       error
	}{
		{
			"Name",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.Integer,
				Value: 141,
			},
			"1.2.3",
			ResultValue{Value: float64(141)},
			nil,
		},
		{
			"Integer",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.Integer,
				Value: 141,
			},
			"1.2.3",
			ResultValue{Value: float64(141)},
			nil,
		},
		{
			"OctetString",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.OctetString,
				Value: []byte(`myVal`),
			},
			"1.2.3",
			ResultValue{Value: []byte(`myVal`)},
			nil,
		},
		{
			"BitString",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.BitString,
				Value: []byte(`myVal`),
			},
			"1.2.3",
			ResultValue{Value: []byte(`myVal`)},
			nil,
		},
		{
			"ObjectIdentifier",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.ObjectIdentifier,
				Value: "1.2.2",
			},
			"1.2.3",
			ResultValue{Value: "1.2.2"},
			nil,
		},
		{
			"ObjectIdentifier need trim",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.ObjectIdentifier,
				Value: ".1.2.2",
			},
			"1.2.3",
			ResultValue{Value: "1.2.2"},
			nil,
		},
		{
			"IPAddress",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.IPAddress,
				Value: "1.2.3.4",
			},
			"1.2.3",
			ResultValue{Value: "1.2.3.4"},
			nil,
		},
		{
			"IPAddress invalid value",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.IPAddress,
				Value: nil,
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: IPAddress should be string type but got type `<nil>` and value `<nil>`"),
		},
		{
			"Null",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.Null,
				Value: nil,
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: invalid type: Null"),
		},
		{
			"Counter32",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.Counter32,
				Value: uint(10),
			},
			"1.2.3",
			ResultValue{SubmissionType: "counter", Value: float64(10)},
			nil,
		},
		{
			"Gauge32",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.Gauge32,
				Value: uint(10),
			},
			"1.2.3",
			ResultValue{Value: float64(10)},
			nil,
		},
		{
			"TimeTicks",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.TimeTicks,
				Value: uint32(10),
			},
			"1.2.3",
			ResultValue{Value: float64(10)},
			nil,
		},
		{
			"Counter64",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.Counter64,
				Value: uint64(10),
			},
			"1.2.3",
			ResultValue{SubmissionType: "counter", Value: float64(10)},
			nil,
		},
		{
			"Uinteger32",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.Uinteger32,
				Value: uint32(10),
			},
			"1.2.3",
			ResultValue{Value: float64(10)},
			nil,
		},
		{
			"OpaqueFloat",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.OpaqueFloat,
				Value: float32(10),
			},
			"1.2.3",
			ResultValue{Value: float64(10)},
			nil,
		},
		{
			"OpaqueDouble",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.OpaqueDouble,
				Value: float64(10),
			},
			"1.2.3",
			ResultValue{Value: float64(10)},
			nil,
		},
		{
			"NoSuchObject",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.NoSuchObject,
				Value: nil,
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: invalid type: NoSuchObject"),
		},
		{
			"NoSuchInstance",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.NoSuchInstance,
				Value: nil,
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: invalid type: NoSuchInstance"),
		},
		{
			"gosnmp.OctetString with wrong type",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.OctetString,
				Value: 1.0,
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: OctetString/BitString should be []byte type but got type `float64` and value `1`"),
		},
		{
			"gosnmp.OpaqueFloat with wrong type",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.OpaqueFloat,
				Value: "abc",
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: OpaqueFloat should be float32 type but got type `string` and value `abc`"),
		},
		{
			"gosnmp.OpaqueDouble with wrong type",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.OpaqueDouble,
				Value: "abc",
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: OpaqueDouble should be float64 type but got type `string` and value `abc`"),
		},
		{
			"gosnmp.ObjectIdentifier with wrong type",
			gosnmp.SnmpPDU{
				Name:  ".1.2.3",
				Type:  gosnmp.ObjectIdentifier,
				Value: 1,
			},
			"1.2.3",
			ResultValue{},
			fmt.Errorf("oid .1.2.3: ObjectIdentifier should be string type but got type `int` and value `1`"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.caseName, func(t *testing.T) {
			name, value, err := GetResultValueFromPDU(tt.pduVariable)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedSnmpValue, value)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func Test_resultToColumnValues(t *testing.T) {
	tests := []struct {
		name                string
		columnOids          []string
		snmpPacket          *gosnmp.SnmpPacket
		expectedValues      ColumnResultValuesType
		expectedNextOidsMap map[string]string
	}{
		{
			"simple nominal case",
			[]string{"1.3.6.1.2.1.2.2.1.14", "1.3.6.1.2.1.2.2.1.2", "1.3.6.1.2.1.2.2.1.20"},
			&gosnmp.SnmpPacket{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.1",
						Type:  gosnmp.Integer,
						Value: 141,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.2.1",
						Type:  gosnmp.OctetString,
						Value: []byte("desc1"),
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.20.1",
						Type:  gosnmp.Integer,
						Value: 201,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.2",
						Type:  gosnmp.Integer,
						Value: 142,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.2.2",
						Type:  gosnmp.OctetString,
						Value: []byte("desc2"),
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.20.2",
						Type:  gosnmp.Integer,
						Value: 202,
					},
				},
			},
			ColumnResultValuesType{
				"1.3.6.1.2.1.2.2.1.14": {
					"1": ResultValue{
						Value: float64(141),
					},
					"2": ResultValue{
						Value: float64(142),
					},
				},
				"1.3.6.1.2.1.2.2.1.2": {
					"1": ResultValue{
						Value: []byte("desc1"),
					},
					"2": ResultValue{
						Value: []byte("desc2"),
					},
				},
				"1.3.6.1.2.1.2.2.1.20": {
					"1": ResultValue{
						Value: float64(201),
					},
					"2": ResultValue{
						Value: float64(202),
					},
				},
			},
			map[string]string{
				"1.3.6.1.2.1.2.2.1.14": "1.3.6.1.2.1.2.2.1.14.2",
				"1.3.6.1.2.1.2.2.1.2":  "1.3.6.1.2.1.2.2.1.2.2",
				"1.3.6.1.2.1.2.2.1.20": "1.3.6.1.2.1.2.2.1.20.2",
			},
		},
		{
			"no such object is skipped",
			[]string{"1.3.6.1.2.1.2.2.1.14", "1.3.6.1.2.1.2.2.1.2"},
			&gosnmp.SnmpPacket{
				Variables: []gosnmp.SnmpPDU{
					{
						Name: "1.3.6.1.2.1.2.2.1.14.1",
						Type: gosnmp.NoSuchObject,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.2.1",
						Type:  gosnmp.OctetString,
						Value: []byte("desc1"),
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.2",
						Type:  gosnmp.Integer,
						Value: 142,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.2.2",
						Type:  gosnmp.OctetString,
						Value: []byte("desc2"),
					},
				},
			},
			ColumnResultValuesType{
				"1.3.6.1.2.1.2.2.1.14": {
					// index 1 not fetched because of gosnmp.NoSuchObject error
					"2": ResultValue{
						Value: float64(142),
					},
				},
				"1.3.6.1.2.1.2.2.1.2": {
					"1": ResultValue{
						Value: []byte("desc1"),
					},
					"2": ResultValue{
						Value: []byte("desc2"),
					},
				},
			},
			map[string]string{
				"1.3.6.1.2.1.2.2.1.14": "1.3.6.1.2.1.2.2.1.14.2",
				"1.3.6.1.2.1.2.2.1.2":  "1.3.6.1.2.1.2.2.1.2.2",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, nextOidsMap := ResultToColumnValues(tt.columnOids, tt.snmpPacket)
			assert.Equal(t, tt.expectedValues, values)
			assert.Equal(t, tt.expectedNextOidsMap, nextOidsMap)
		})
	}
}

func Test_resultToScalarValues(t *testing.T) {
	tests := []struct {
		name           string
		snmpPacket     *gosnmp.SnmpPacket
		expectedValues ScalarResultValuesType
	}{
		{
			"simple case",
			&gosnmp.SnmpPacket{
				Variables: []gosnmp.SnmpPDU{
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.1",
						Type:  gosnmp.Integer,
						Value: 142,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.2",
						Type:  gosnmp.Counter32,
						Value: 142,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.3",
						Type:  gosnmp.NoSuchInstance,
						Value: 142,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.4",
						Type:  gosnmp.NoSuchObject,
						Value: 142,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.5",
						Type:  gosnmp.EndOfContents,
						Value: 142,
					},
					{
						Name:  "1.3.6.1.2.1.2.2.1.14.6",
						Type:  gosnmp.EndOfMibView,
						Value: 142,
					},
				},
			},
			ScalarResultValuesType{
				"1.3.6.1.2.1.2.2.1.14.1": {
					Value: float64(142),
				},
				"1.3.6.1.2.1.2.2.1.14.2": {
					SubmissionType: "counter",
					Value:          float64(142),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values := ResultToScalarValues(tt.snmpPacket)
			assert.Equal(t, tt.expectedValues, values)
		})
	}
}

//------------------------------------------------------------------------------

func TestToFloat64FromString(t *testing.T) {
	snmpValue := &ResultValue{
		SubmissionType: "gauge",
		Value:          "255.745",
	}
	value, err := snmpValue.ToFloat64()
	assert.NoError(t, err)
	assert.Equal(t, float64(255.745), value)
}

func TestToFloat64FromFloat(t *testing.T) {
	snmpValue := &ResultValue{
		SubmissionType: "gauge",
		Value:          float64(255.745),
	}
	value, err := snmpValue.ToFloat64()
	assert.NoError(t, err)
	assert.Equal(t, float64(255.745), value)
}

func TestToFloat64FromInvalidType(t *testing.T) {
	snmpValue := &ResultValue{
		SubmissionType: "gauge",
		Value:          int64(255),
	}
	_, err := snmpValue.ToFloat64()
	assert.NotNil(t, err)
}

func TestResultValue_ToString(t *testing.T) {
	tests := []struct {
		name          string
		resultValue   ResultValue
		expectedStr   string
		expectedError string
	}{
		{
			name: "hexify",
			resultValue: ResultValue{
				Value: []byte{0xff, 0xaa, 0x00},
			},
			expectedStr:   "0xffaa00",
			expectedError: "",
		},
		{
			name: "do not hexify newline and tabs",
			resultValue: ResultValue{
				Value: []byte(`m\ny\rV\ta\n\r\tl`),
			},
			expectedStr:   "m\\ny\\rV\\ta\\n\\r\\tl",
			expectedError: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strValue, err := tt.resultValue.ToString()
			if tt.expectedError == "" {
				assert.Nil(t, err)
			} else {
				assert.Contains(t, err.Error(), tt.expectedError)
			}
			assert.Equal(t, tt.expectedStr, strValue)
		})
	}
}

//------------------------------------------------------------------------------

var storeMock = &ResultValueStore{
	ScalarValues: ScalarResultValuesType{
		"1.1.1.1.0": {Value: float64(10)},   // a float value
		"1.1.1.2.0": {Value: "a_str_value"}, // a string value
		"1.1.1.3.0": {Value: nil},           // invalid type value
	},
	ColumnValues: ColumnResultValuesType{
		"1.1.1": {
			"1": ResultValue{Value: float64(10)},   // a float value
			"2": ResultValue{Value: "a_str_value"}, // a string value
			"3": ResultValue{Value: nil},           // invalid type value
		},
		"1.1.2": {
			"1": ResultValue{Value: float64(21)},
			"2": ResultValue{Value: float64(22)},
		},
	},
}

func Test_resultValueStore_getColumnValueAsFloat(t *testing.T) {
	assert.Equal(t, float64(0), storeMock.GetColumnValueAsFloat("0.0", "1"))    // wrong column
	assert.Equal(t, float64(10), storeMock.GetColumnValueAsFloat("1.1.1", "1")) // ok float value
	assert.Equal(t, float64(0), storeMock.GetColumnValueAsFloat("1.1.1", "2"))  // cannot convert str to float
	assert.Equal(t, float64(0), storeMock.GetColumnValueAsFloat("1.1.1", "3"))  // wrong type
	assert.Equal(t, float64(0), storeMock.GetColumnValueAsFloat("1.1.1", "99")) // index not found
	assert.Equal(t, float64(21), storeMock.GetColumnValueAsFloat("1.1.2", "1")) // ok float value
}

func Test_resultValueStore_GetColumnIndexes(t *testing.T) {
	indexes, err := storeMock.GetColumnIndexes("0.0")
	assert.EqualError(t, err, "error getting column value oid=0.0: value for Column OID `0.0` not found in results")
	assert.Nil(t, indexes)

	indexes, err = storeMock.GetColumnIndexes("1.1.1")
	assert.NoError(t, err)
	assert.Equal(t, []string{"1", "2", "3"}, indexes)
}

func TestResultValueStoreAsString(t *testing.T) {
	store := &ResultValueStore{
		ScalarValues: ScalarResultValuesType{
			"1.1.1.1.0": {Value: float64(10)}, // a float value
		},
		ColumnValues: ColumnResultValuesType{
			"1.1.1": {
				"1": ResultValue{Value: float64(10)}, // a float value
			},
		},
	}
	str := ResultValueStoreAsString(store)
	assert.Equal(t, "{\"scalar_values\":{\"1.1.1.1.0\":{\"value\":10}},\"column_values\":{\"1.1.1\":{\"1\":{\"value\":10}}}}", str)

	str = ResultValueStoreAsString(nil)
	assert.Equal(t, "", str)
}

//------------------------------------------------------------------------------
