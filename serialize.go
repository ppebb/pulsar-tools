package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
)

// Returns the number of bytes written to out
func UnmarshalPulsarType(b []byte, out any) error {
	ret := reflect.ValueOf(out)
	if ret.Kind() == reflect.Ptr && !ret.IsNil() {
		ret = ret.Elem()
	}

	offset := 0
	return unmarshalPulsarTypeInner(b, &offset, &ret)
}

func unmarshalPulsarTypeInner(bytes []byte, offset *int, out *reflect.Value) error {
	pulsarType := out.Type()
	if pulsarType.Kind() == reflect.Ptr {
		pulsarType = pulsarType.Elem()
	}

	fields := reflect.VisibleFields(pulsarType)

	for i, field := range fields {
		fieldType := field.Type

		switch fieldType.Kind() {
		case reflect.Struct:
			{
				retInner := reflect.New(fieldType).Elem()
				err := unmarshalPulsarTypeInner(bytes, offset, &retInner)
				if err != nil {
					return err
				}

				out.Field(i).Set(retInner)
			}
		case reflect.Array:
			{
				arrayLen := fieldType.Len()
				retInner := reflect.New(reflect.ArrayOf(arrayLen, fieldType.Elem())).Elem()

				for j := range arrayLen {
					prim := reflect.New(fieldType.Elem()).Elem()
					err := unmarshalPulsarTypeInner(bytes, offset, &prim)
					if err != nil {
						return err
					}
					retInner.Index(j).Set(prim)
				}

				out.Field(i).Set(retInner)
			}
		default:
			prim := reflect.New(fieldType).Elem()
			err := unmarshalPrimitive(bytes, offset, &prim)
			if err != nil {
				return err
			}

			out.Field(i).Set(prim)
		}
	}

	return nil
}

func unmarshalPrimitive(bytes []byte, offset *int, out *reflect.Value) error {
	size := int(out.Type().Size())
	bslice := bytes[*offset : *offset+size]
	*offset = *offset + size

	switch out.Type().Kind() {
	case reflect.Bool:
		out.SetBool(bslice[0] != 0)

	case reflect.Int8:
		out.SetInt(int64(bslice[0]))
	case reflect.Int16:
		out.SetInt(int64(binary.BigEndian.Uint16(bslice)))
	case reflect.Int32:
		out.SetInt(int64(binary.BigEndian.Uint32(bslice)))
	case reflect.Int64:
		out.SetInt(int64(binary.BigEndian.Uint64(bslice)))

	case reflect.Uint8:
		out.SetUint(uint64(bslice[0]))
	case reflect.Uint16:
		out.SetUint(uint64(binary.BigEndian.Uint16(bslice)))
	case reflect.Uint32:
		out.SetUint(uint64(binary.BigEndian.Uint32(bslice)))
	case reflect.Uint64:
		out.SetUint(uint64(binary.BigEndian.Uint64(bslice)))

	case reflect.Float32:
		out.SetFloat(float64(math.Float32frombits(binary.BigEndian.Uint32(bslice))))
	case reflect.Float64:
		out.SetFloat(float64(math.Float64frombits(binary.BigEndian.Uint64(bslice))))
	default:
		return fmt.Errorf("Unsupported type '%s' found in struct!", out.String())
	}

	return nil
}
