// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package traps

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/gosnmp/gosnmp"
	"github.com/pkg/errors"
)

const ddsource string = "snmp-traps"

// Formatter is an interface to extract and format raw SNMP Traps
type Formatter interface {
	FormatPacket(packet *SnmpPacket) ([]byte, error)
}

// JSONFormatter is a Formatter implementation that transforms Traps into JSON
type JSONFormatter struct {
	oidResolver OIDResolver
	namespace   string
}

type trapVariable struct {
	OID     string      `json:"oid"`
	VarType string      `json:"type"`
	Value   interface{} `json:"value"`
}

const (
	sysUpTimeInstanceOID = "1.3.6.1.2.1.1.3.0"
	snmpTrapOID          = "1.3.6.1.6.3.1.1.4.1.0"
)

// NewJSONFormatter creates a new JSONFormatter instance with an optional OIDResolver variable.
func NewJSONFormatter(oidResolver OIDResolver, namespace string) (JSONFormatter, error) {
	if oidResolver == nil {
		return JSONFormatter{}, fmt.Errorf("NewJSONFormatter called with a nil OIDResolver")
	}
	return JSONFormatter{oidResolver, namespace}, nil
}

// FormatPacket converts a raw SNMP trap packet to a FormattedSnmpPacket containing the JSON data and the tags to attach
// {
//	"trap": {
//    "ddsource": "snmp-traps",
//    "ddtags": "namespace:default,snmp_device:10.0.0.2,...",
//    "timestamp": 123456789,
//    "snmpTrapName": "...",
//    "snmpTrapOID": "1.3.6.1.5.3.....",
//    "snmpTrapMIB": "...",
//    "uptime": "12345",
//    "genericTrap": "5", # v1 only
//    "specificTrap": "0",  # v1 only
//    "variables": [
//      {
//        "oid": "1.3.4.1....",
//        "type": "integer",
//        "value": 12
//      },
//      ...
//    ],
//   }
// }
func (f JSONFormatter) FormatPacket(packet *SnmpPacket) ([]byte, error) {
	payload := make(map[string]interface{})
	var formattedTrap map[string]interface{}
	var err error
	if packet.Content.Version == gosnmp.Version1 {
		formattedTrap = f.formatV1Trap(packet.Content)
	} else {
		formattedTrap, err = f.formatTrap(packet.Content)
		if err != nil {
			return nil, err
		}
	}
	formattedTrap["ddsource"] = ddsource
	formattedTrap["ddtags"] = strings.Join(f.getTags(packet), ",")
	formattedTrap["timestamp"] = packet.Timestamp
	payload["trap"] = formattedTrap
	return json.Marshal(payload)
}

// GetTags returns a list of tags associated to an SNMP trap packet.
func (f JSONFormatter) getTags(packet *SnmpPacket) []string {
	return []string{
		"snmp_version:" + formatVersion(packet.Content),
		"device_namespace:" + f.namespace,
		"snmp_device:" + packet.Addr.IP.String(),
	}
}

func (f JSONFormatter) formatV1Trap(packet *gosnmp.SnmpPacket) map[string]interface{} {
	data := make(map[string]interface{})
	data["uptime"] = uint32(packet.Timestamp)
	enterpriseOid := NormalizeOID(packet.Enterprise)
	genericTrap := packet.GenericTrap
	specificTrap := packet.SpecificTrap
	var trapOID string
	if genericTrap == 6 {
		// Vendor-specific trap
		trapOID = fmt.Sprintf("%s.0.%d", enterpriseOid, specificTrap)
	} else {
		// Generic trap
		trapOID = fmt.Sprintf("%s.%d", genericTrapOid, genericTrap+1)
	}
	data["snmpTrapOID"] = trapOID
	trapMetadata, err := f.oidResolver.GetTrapMetadata(trapOID)
	if err != nil {
		log.Debugf("unable to resolve OID: %s", err)
	} else {
		data["snmpTrapName"] = trapMetadata.Name
		data["snmpTrapMIB"] = trapMetadata.MIBName
	}
	data["enterpriseOID"] = enterpriseOid
	data["genericTrap"] = genericTrap
	data["specificTrap"] = specificTrap
	variables := parseVariables(packet.Variables)
	data["variables"] = variables
	for _, variable := range variables {
		varMetadata, err := f.oidResolver.GetVariableMetadata(trapOID, variable.OID)
		if err != nil {
			log.Debugf("unable to enrich variable: %s", err)
			continue
		}
		data[varMetadata.Name] = parseValue(variable, varMetadata)
	}
	return data
}

func (f JSONFormatter) formatTrap(packet *gosnmp.SnmpPacket) (map[string]interface{}, error) {
	/*
		An SNMP v2 or v3 trap packet consists in the following variables (PDUs):
		{sysUpTime.0, snmpTrapOID.0, additionalDataVariables...}
		See: https://tools.ietf.org/html/rfc3416#section-4.2.6
	*/
	variables := packet.Variables
	if len(variables) < 2 {
		return nil, fmt.Errorf("expected at least 2 variables, got %d", len(variables))
	}

	data := make(map[string]interface{})

	uptime, err := parseSysUpTime(variables[0])
	if err != nil {
		return nil, err
	}
	data["uptime"] = uptime

	trapOID, err := parseSnmpTrapOID(variables[1])
	if err != nil {
		return nil, err
	}
	data["snmpTrapOID"] = trapOID

	trapMetadata, err := f.oidResolver.GetTrapMetadata(trapOID)
	if err != nil {
		log.Debugf("unable to resolve OID: %s", err)
	} else {
		data["snmpTrapName"] = trapMetadata.Name
		data["snmpTrapMIB"] = trapMetadata.MIBName
	}

	parsedVariables := parseVariables(variables[2:])
	data["variables"] = parsedVariables
	for _, variable := range parsedVariables {
		varMetadata, err := f.oidResolver.GetVariableMetadata(trapOID, variable.OID)
		if err != nil {
			log.Debugf("unable to enrich variable: %s", err)
			continue
		}
		data[varMetadata.Name] = parseValue(variable, varMetadata)
	}
	return data, nil
}

// NormalizeOID convert an OID from the absolute form ".1.2.3..." to a relative form "1.2.3..."
func NormalizeOID(value string) string {
	// OIDs can be formatted as ".1.2.3..." ("absolute form") or "1.2.3..." ("relative form").
	// Convert everything to relative form, like we do in the Python check.
	return strings.TrimLeft(value, ".")
}

// parseValue checks to see if the variable has a mapping in an enum and
// returns the mapping if it exists, otherwise returns the value unchanged
func parseValue(variable trapVariable, varMetadata VariableMetadata) interface{} {
	if len(varMetadata.Enumeration) > 0 {
		switch variable.Value.(type) {
		case int:
			// if we find a mapping set it and return
			i := variable.Value.(int)
			if value, ok := varMetadata.Enumeration[i]; !ok || variable.VarType != "integer" {
				log.Debugf("unable to find enum mapping for value %d variable %q", i, varMetadata.Name)
			} else {
				return value
			}
		case string:
			// do bitwise search
			if variable.VarType != "string" {
				log.Warnf("received string for non-string variable: %+v", variable)
				return variable.Value
			}
			s := variable.Value.(string)
			enabledValues := make([]string, 0)
			sBytes := []byte(s)
			for i, b := range sBytes {
				for j := 0; j < 8; j++ {
					position := j + i*8 // position is the index in the current byte plus 8 * the position in the byte array
					enabled, err := isBitEnabled(uint8(b), j)
					if err != nil {
						log.Debugf("unable to determine status at position %d: %s", position, err.Error())
					}
					if enabled {
						value := varMetadata.Enumeration[position]
						if value == "" {
							log.Debugf("unable to find enum mapping for value %f variable %q", i, varMetadata.Name)
						} else {
							enabledValues = append(enabledValues, value)
						}
					}
				}
			}
			if len(enabledValues) > 0 {
				return enabledValues
			}
		default:
			log.Debugf("value is not an enum compatible type (i.e. int, string): %+v", variable.Value)
		}
	}

	// if no mapping is found or type is not integer
	return variable.Value
}

func parseSysUpTime(variable gosnmp.SnmpPDU) (uint32, error) {
	name := NormalizeOID(variable.Name)
	if name != sysUpTimeInstanceOID {
		return 0, fmt.Errorf("expected OID %s, got %s", sysUpTimeInstanceOID, name)
	}

	value, ok := variable.Value.(uint32)
	if !ok {
		return 0, fmt.Errorf("expected uptime to be uint32 (got %v of type %T)", variable.Value, variable.Value)
	}

	return value, nil
}

func parseSnmpTrapOID(variable gosnmp.SnmpPDU) (string, error) {
	name := NormalizeOID(variable.Name)
	if name != snmpTrapOID {
		return "", fmt.Errorf("expected OID %s, got %s", snmpTrapOID, name)
	}

	value := ""
	switch variable.Value.(type) {
	case string:
		value = variable.Value.(string)
	case []byte:
		value = string(variable.Value.([]byte))
	default:
		return "", fmt.Errorf("expected snmpTrapOID to be a string (got %v of type %T)", variable.Value, variable.Value)
	}

	return NormalizeOID(value), nil
}

func parseVariables(variables []gosnmp.SnmpPDU) []trapVariable {
	var parsedVariables []trapVariable

	for _, variable := range variables {
		varOID := NormalizeOID(variable.Name)
		varType := formatType(variable)
		varValue := formatValue(variable)
		parsedVariables = append(parsedVariables, trapVariable{OID: varOID, VarType: varType, Value: varValue})
	}

	return parsedVariables
}

func formatType(variable gosnmp.SnmpPDU) string {
	switch variable.Type {
	case gosnmp.Integer, gosnmp.Uinteger32:
		return "integer"
	case gosnmp.OctetString, gosnmp.BitString:
		return "string"
	case gosnmp.ObjectIdentifier:
		return "oid"
	case gosnmp.Counter32:
		return "counter32"
	case gosnmp.Counter64:
		return "counter64"
	case gosnmp.Gauge32:
		return "gauge32"
	default:
		return "other"
	}
}

func formatValue(variable gosnmp.SnmpPDU) interface{} {
	switch variable.Value.(type) {
	case []byte:
		return string(variable.Value.([]byte))
	default:
		return variable.Value
	}
}

func formatVersion(packet *gosnmp.SnmpPacket) string {
	switch packet.Version {
	case gosnmp.Version3:
		return "3"
	case gosnmp.Version2c:
		return "2"
	case gosnmp.Version1:
		return "1"
	default:
		return "unknown"
	}
}

// isBitEnabled takes in a uint8 and returns true if
// the bit at the passed position is 1.
// Each byte is little endian meaning if
// you have the binary 10000000, passing position 0
// would return true and 7 would return false
func isBitEnabled(n uint8, pos int) (bool, error) {
	if pos > 7 {
		return false, errors.Errorf("invalid position %d, must be 0-7.", pos)
	}
	val := n & uint8(1<<(7-pos))
	return val > 0, nil
}
