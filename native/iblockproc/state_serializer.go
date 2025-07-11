package iblockproc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"unicode"

	"github.com/unicornultrafoundation/go-helios/hash"
	"github.com/unicornultrafoundation/go-helios/native/idx"
	"github.com/unicornultrafoundation/go-helios/native/pos"

	"github.com/unicornultrafoundation/go-u2u/native"
)

func (es EpochState) String() string {
	// Create a more compact, single-line format that's readable
	s := fmt.Sprintf("EpochState{Epoch:%d, Start:%d, PrevStart:%d, StateRoot:%s",
		es.Epoch, es.EpochStart, es.PrevEpochStart, es.EpochStateRoot.String())

	if es.Validators != nil {
		s += fmt.Sprintf(", ValidatorsCount:%d", es.Validators.Len())
	}

	if len(es.ValidatorStates) > 0 {
		s += ", ValidatorStates:["
		for i, vs := range es.ValidatorStates {
			if i > 0 {
				s += ", "
			}
			s += fmt.Sprintf("{GasRefund:%d, PrevEvent:%s, Time:%d}",
				vs.GasRefund, vs.PrevEpochEvent.ID.String(), vs.PrevEpochEvent.Time)
		}
		s += "]"
	}

	if len(es.ValidatorProfiles) > 0 {
		s += ", ValidatorProfiles:{"
		first := true
		for id, profile := range es.ValidatorProfiles {
			if !first {
				s += ", "
			}
			first = false
			s += fmt.Sprintf("%d:{Weight:%s}", id, profile.Weight.String())
		}
		s += "}"
	}

	s += fmt.Sprintf(", Rules:{Name:%s, NetworkID:%d, MaxEpochGas:%d}}",
		es.Rules.Name, es.Rules.NetworkID, es.Rules.Epochs.MaxEpochGas)

	return s
}

// toCamelCase converts a field name to camelCase (lowercase first character)
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	// Special case for "ID" - should be "id" not "iD"
	if s == "ID" {
		return "id"
	}

	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

// fromCamelCase converts a camelCase field name back to PascalCase
func fromCamelCase(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// structToMap converts a struct to map[string]interface{} using camelCase field names
func structToMap(v interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %T", v)
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		var fieldName string
		// Check if field has existing json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			// Use existing json tag name
			fieldName = strings.Split(jsonTag, ",")[0]
		} else {
			// Convert field name to camelCase
			fieldName = toCamelCase(field.Name)
		}

		// Recursively convert the field value
		convertedValue, err := convertValue(fieldValue.Interface())
		if err != nil {
			return nil, fmt.Errorf("failed to convert field %s: %w", field.Name, err)
		}

		result[fieldName] = convertedValue
	}

	return result, nil
}

// convertValue recursively converts values to ensure all nested structs use camelCase
func convertValue(v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	val := reflect.ValueOf(v)

	// Handle special types BEFORE pointer handling to catch *pos.Validators
	switch typed := v.(type) {
	case *pos.Validators:
		if typed == nil {
			return nil, nil
		}
		// Serialize pos.Validators to match original struct format with just values
		values := make(map[string]string)
		for i := 0; i < int(typed.Len()); i++ {
			validatorIdx := idx.Validator(i)
			validatorID := typed.GetID(validatorIdx)
			weight := typed.GetWeightByIdx(validatorIdx)
			values[fmt.Sprintf("%d", validatorID)] = fmt.Sprintf("%d", weight)
		}

		result := make(map[string]interface{})
		result["values"] = values
		return result, nil
	}

	// Handle pointers (for other types)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, nil
		}
		return convertValue(val.Elem().Interface())
	}

	// Handle other special types - convert to appropriate JSON representations
	switch typed := v.(type) {
	case hash.Hash:
		return typed.String(), nil
	case hash.Event:
		return typed.String(), nil
	case *big.Int:
		if typed == nil {
			return nil, nil
		}
		return typed.String(), nil
	case big.Int:
		return typed.String(), nil
	}

	switch val.Kind() {
	case reflect.Slice:
		// Handle byte slices specially - convert to hex string
		if val.Type().Elem().Kind() == reflect.Uint8 {
			bytes := val.Bytes()
			return fmt.Sprintf("0x%x", bytes), nil
		}

		// Convert other slice/array elements
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			converted, err := convertValue(val.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			result[i] = converted
		}
		return result, nil

	case reflect.Array:
		// Handle byte arrays specially - convert to hex string
		if val.Type().Elem().Kind() == reflect.Uint8 {
			bytes := make([]byte, val.Len())
			for i := 0; i < val.Len(); i++ {
				bytes[i] = val.Index(i).Interface().(byte)
			}
			return fmt.Sprintf("0x%x", bytes), nil
		}

		// Convert other array elements
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			converted, err := convertValue(val.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			result[i] = converted
		}
		return result, nil

	case reflect.Struct:
		// Convert struct to map with camelCase field names
		return structToMap(v)

	case reflect.Map:
		// Convert map values (keeping keys as-is)
		result := make(map[string]interface{})
		for _, key := range val.MapKeys() {
			keyStr := fmt.Sprintf("%v", key.Interface())
			converted, err := convertValue(val.MapIndex(key).Interface())
			if err != nil {
				return nil, err
			}
			result[keyStr] = converted
		}
		return result, nil

	default:
		// Return primitive types as-is
		return v, nil
	}
}

// mapToStruct converts a map[string]interface{} back to struct using camelCase field names
func mapToStruct(m map[string]interface{}, target interface{}) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		if !field.IsExported() || !fieldValue.CanSet() {
			continue
		}

		var mapKey string
		// Check if field has existing json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			mapKey = strings.Split(jsonTag, ",")[0]
		} else {
			mapKey = toCamelCase(field.Name)
		}

		if value, exists := m[mapKey]; exists && value != nil {
			// Convert value to proper type recursively
			if err := setFieldValue(fieldValue, value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// setFieldValue sets a reflect.Value with proper type conversion
func setFieldValue(fieldValue reflect.Value, value interface{}) error {
	if value == nil {
		return nil
	}

	valueReflect := reflect.ValueOf(value)
	fieldType := fieldValue.Type()

	// Handle special hash types - convert from hex strings
	if str, ok := value.(string); ok && strings.HasPrefix(str, "0x") {
		switch fieldType {
		case reflect.TypeOf(hash.Hash{}):
			fieldValue.Set(reflect.ValueOf(hash.HexToHash(str)))
			return nil
		case reflect.TypeOf(hash.Event{}):
			fieldValue.Set(reflect.ValueOf(hash.Event(hash.HexToHash(str))))
			return nil
		}

		// Handle byte slices/arrays from hex strings
		if fieldType.Kind() == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8 {
			// Remove 0x prefix and decode hex
			hexStr := strings.TrimPrefix(str, "0x")
			bytes := make([]byte, len(hexStr)/2)
			for i := 0; i < len(hexStr); i += 2 {
				if i+1 < len(hexStr) {
					high := hexCharToByte(hexStr[i])
					low := hexCharToByte(hexStr[i+1])
					bytes[i/2] = (high << 4) | low
				}
			}
			fieldValue.Set(reflect.ValueOf(bytes))
			return nil
		}

		if fieldType.Kind() == reflect.Array && fieldType.Elem().Kind() == reflect.Uint8 {
			// Remove 0x prefix and decode hex
			hexStr := strings.TrimPrefix(str, "0x")
			arrayLen := fieldType.Len()
			bytes := make([]byte, len(hexStr)/2)
			for i := 0; i < len(hexStr); i += 2 {
				if i+1 < len(hexStr) {
					high := hexCharToByte(hexStr[i])
					low := hexCharToByte(hexStr[i+1])
					bytes[i/2] = (high << 4) | low
				}
			}

			// Copy to array
			arr := reflect.New(fieldType).Elem()
			for i := 0; i < arrayLen && i < len(bytes); i++ {
				arr.Index(i).Set(reflect.ValueOf(bytes[i]))
			}
			fieldValue.Set(arr)
			return nil
		}
	}

	// Handle direct assignment if types are compatible
	if valueReflect.Type().AssignableTo(fieldType) {
		fieldValue.Set(valueReflect)
		return nil
	}

	// Handle convertible types
	if valueReflect.Type().ConvertibleTo(fieldType) {
		fieldValue.Set(valueReflect.Convert(fieldType))
		return nil
	}

	// Handle pointer fields
	if fieldType.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			fieldValue.Set(reflect.New(fieldType.Elem()))
		}
		return setFieldValue(fieldValue.Elem(), value)
	}

	// Handle struct fields
	if fieldType.Kind() == reflect.Struct {
		if valueMap, ok := value.(map[string]interface{}); ok {
			return mapToStruct(valueMap, fieldValue.Addr().Interface())
		}
	}

	// Handle slice fields
	if fieldType.Kind() == reflect.Slice {
		if valueSlice, ok := value.([]interface{}); ok {
			slice := reflect.MakeSlice(fieldType, len(valueSlice), len(valueSlice))
			for i, item := range valueSlice {
				if err := setFieldValue(slice.Index(i), item); err != nil {
					return err
				}
			}
			fieldValue.Set(slice)
			return nil
		}
	}

	// Handle map fields
	if fieldType.Kind() == reflect.Map {
		if valueMap, ok := value.(map[string]interface{}); ok {
			mapValue := reflect.MakeMap(fieldType)
			for k, v := range valueMap {
				keyValue := reflect.ValueOf(k)
				elemValue := reflect.New(fieldType.Elem()).Elem()
				if err := setFieldValue(elemValue, v); err != nil {
					return err
				}
				mapValue.SetMapIndex(keyValue, elemValue)
			}
			fieldValue.Set(mapValue)
			return nil
		}
	}

	return fmt.Errorf("cannot convert %T to %s", value, fieldType)
}

// hexCharToByte converts a hex character to byte value
func hexCharToByte(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	default:
		return 0
	}
}

// MarshalJSON implements custom JSON marshaling with automatic camelCase conversion
func (es EpochState) MarshalJSON() ([]byte, error) {
	m, err := structToMap(es)
	if err != nil {
		return nil, err
	}

	return json.Marshal(m)
}

// UnmarshalJSON implements custom JSON unmarshaling with automatic camelCase conversion
func (es *EpochState) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	return mapToStruct(m, es)
}

// MarshalJSON implements custom JSON marshaling with automatic camelCase conversion for EventInfo
func (ei EventInfo) MarshalJSON() ([]byte, error) {
	// Create a custom map for EventInfo with ID as string
	result := make(map[string]interface{})

	// Handle ID field specially - convert to string
	result["id"] = ei.ID.String()

	// Handle other fields normally with camelCase conversion
	result["gasPowerLeft"] = ei.GasPowerLeft
	result["time"] = ei.Time

	return json.Marshal(result)
}

// UnmarshalJSON implements custom JSON unmarshaling with automatic camelCase conversion for EventInfo
func (ei *EventInfo) UnmarshalJSON(data []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	// Handle ID field specially - parse from string
	if idStr, exists := m["id"]; exists {
		if idString, ok := idStr.(string); ok {
			// Parse the string back to hash.Event
			ei.ID = hash.Event(hash.HexToHash(idString))
		} else {
			return fmt.Errorf("ID field must be a string")
		}
	}

	// Handle other fields normally
	if gasPowerLeft, exists := m["gasPowerLeft"]; exists {
		// Convert the map back to native.GasPowerLeft using reflection
		if gplMap, ok := gasPowerLeft.(map[string]interface{}); ok {
			var gpl native.GasPowerLeft
			if err := mapToStruct(gplMap, &gpl); err != nil {
				return fmt.Errorf("failed to unmarshal gasPowerLeft: %w", err)
			}
			ei.GasPowerLeft = gpl
		} else {
			return fmt.Errorf("gasPowerLeft field must be an object")
		}
	}

	if timeVal, exists := m["time"]; exists {
		switch t := timeVal.(type) {
		case float64:
			ei.Time = native.Timestamp(t)
		case int64:
			ei.Time = native.Timestamp(t)
		case int:
			ei.Time = native.Timestamp(t)
		default:
			return fmt.Errorf("time field must be a number")
		}
	}

	return nil
}

// ToMap converts EventInfo to a map[string]interface{} using JSON field names as keys
func (ei *EventInfo) ToMap() (map[string]interface{}, error) {
	// First marshal to JSON bytes
	jsonBytes, err := json.Marshal(ei)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal EventInfo to JSON: %w", err)
	}

	// Then unmarshal into a map
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return result, nil
}

// ToMap converts EpochState to a map[string]interface{} using JSON field names as keys
func (es *EpochState) ToMap() (map[string]interface{}, error) {
	// Use structToMap directly instead of JSON round-trip to preserve custom type handling
	result, err := structToMap(*es)
	if err != nil {
		return nil, fmt.Errorf("failed to convert EpochState to map: %w", err)
	}

	// Add extra computed fields
	result["hash"] = es.Hash().String()
	result["end"] = es.EpochStart
	result["start"] = es.PrevEpochStart

	// Remove the original timestamp fields to avoid duplication
	delete(result, "epochStart")
	delete(result, "prevEpochStart")

	return result, nil
}
