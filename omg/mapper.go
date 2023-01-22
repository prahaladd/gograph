package omg

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/prahaladd/gograph/core"
)

const ogmTagSuffix = "ogm"

// Mapper interface defines a contract for implementations that map arbitrary structs to
// graph entities (vertices and structs)
type Mapper interface {

	// ToVertex maps a specified struct to a graph vertex with the specified labels.
	//
	// The properties of the vertex are populated from the fields of the struct.
	//
	// Implementations must not expect the label to be always specified. In cases when label
	// is not specified the type of the value serves as the vertex label.
	//
	// The passed in value to be mapped must be a struct or a pointer to a struct.
	// Nested structs are not currently supported
	ToVertex(v any, labels []string) (*core.Vertex, error)

	// ToEdge maps a specified struct to a graph edge with the specified labels.
	//
	// The properties of the edge are populated from the fields of the struct
	//
	// Implementations must not expect the label to be always specified. In cases when label
	// is not specified the type of the struct serves as the edge label.
	//
	// The passed in value to be mapped must be a struct or a pointer to a struct.
	// Nested structs are not currently supported
	ToEdge(v any, label *string) (*core.Edge, error)

	// FromVertex maps a vertex properties to a user-defined struct.
	//
	// Vertex label information is not retained as a part of the conversion process.
	// Hence, if the vertex node has a label other than the struct type name, then the label information is lost.

	FromVertex(vertex *core.Vertex, v any) error

	// FromEdge maps an edge properties to a user-defined struct.
	//
	// Edge label information is not retained as a part of the conversion process.
	// Hence, if the label node has a label other than the struct type name, then the label information is lost.
	FromEdge(edge *core.Edge, v any) error
}

type ReflectionMapper struct {
}

func (rm *ReflectionMapper) ToVertex(v any, labels []string) (*core.Vertex, error) {

	typeOfV := reflect.TypeOf(v)

	kindOfV := typeOfV.Kind()

	var typeNameOfV string

	switch kindOfV {
	case reflect.Struct:

		typeNameOfV = typeOfV.Name()
		vertex := core.Vertex{}
		if len(labels) > 0 {
			vertex.Labels = labels
		} else {
			vertex.Labels = []string{typeNameOfV}
		}
		vertex.Properties = rm.performMap(typeOfV, reflect.ValueOf(v))
		return &vertex, nil

	case reflect.Ptr:
		typeNameOfV = typeOfV.Elem().Name()
		vertex := core.Vertex{}
		if len(labels) > 0 {
			vertex.Labels = labels
		} else {
			vertex.Labels = []string{typeNameOfV}
		}
		vertex.Properties = rm.performMap(typeOfV.Elem(), reflect.Indirect(reflect.ValueOf(v)))
		return &vertex, nil
	default:
		return nil, errors.New("passed in value must be a struct or pointer to a struct")
	}
}

func (rm *ReflectionMapper) ToEdge(v interface{}, label *string) (*core.Edge, error) {

	typeOfV := reflect.TypeOf(v)

	kindOfV := typeOfV.Kind()

	var typeNameOfV string

	switch kindOfV {
	case reflect.Struct:

		typeNameOfV = typeOfV.Name()
		edge := core.Edge{}
		if label == nil {
			edge.Type = typeNameOfV
		} else {
			if len(*label) > 0 {
				edge.Type = *label
			} else {
				edge.Type = typeNameOfV
			}
		}

		edge.Properties = rm.performMap(typeOfV, reflect.ValueOf(v))
		return &edge, nil

	case reflect.Ptr:
		typeNameOfV = typeOfV.Elem().Name()
		edge := core.Edge{}
		if label == nil {
			edge.Type = typeNameOfV
		} else {
			if len(*label) > 0 {
				edge.Type = *label
			} else {
				edge.Type = typeNameOfV
			}
		}
		edge.Properties = rm.performMap(typeOfV.Elem(), reflect.Indirect(reflect.ValueOf(v)))
		return &edge, nil
	default:
		return nil, errors.New("passed in value must be a struct or pointer to a struct")
	}
}

func (rm *ReflectionMapper) FromVertex(vertex *core.Vertex, v any) error {
	typeOfV := reflect.TypeOf(v)

	kindOfV := typeOfV.Kind()

	var finalValue reflect.Value
	switch kindOfV {
	case reflect.Ptr:
		if typeOfV.Elem().Kind() != reflect.Struct {
			return errors.New("passed in value must be a pointer to a struct type")
		}

		if reflect.ValueOf(v).IsNil() {
			finalValue = reflect.New(reflect.TypeOf(v))
		} else {
			finalValue = reflect.ValueOf(v)
		}
		//rm.performReverseMap(vertex.Properties, reflect.TypeOf(v), reflect.Indirect(finalValue))
		return rm.performDecode(vertex.Properties, reflect.TypeOf(v), reflect.Indirect(finalValue), v)
	default:
		return errors.New("passed in value must be a pointer to a struct type")
	}
}

func (rm *ReflectionMapper) FromEdge(edge *core.Edge, v any) error {
	typeOfV := reflect.TypeOf(v)

	kindOfV := typeOfV.Kind()

	var finalValue reflect.Value
	switch kindOfV {
	case reflect.Ptr:
		if typeOfV.Elem().Kind() != reflect.Struct {
			return errors.New("passed in value must be a pointer to a struct type")
		}

		if reflect.ValueOf(v).IsNil() {
			finalValue = reflect.New(reflect.TypeOf(v))
		} else {
			finalValue = reflect.ValueOf(v)
		}
		err := rm.performDecode(edge.Properties, reflect.TypeOf(v), reflect.Indirect(finalValue), v)
		return err
	default:
		return errors.New("passed in value must be a pointer to a struct type")
	}
}

func (rm *ReflectionMapper) performMap(t reflect.Type, val reflect.Value) core.KVMap {

	props := core.KVMap{}
	for i := 0; i < val.NumField(); i++ {
		var key string
		if t.Field(i).Tag != "" && t.Field(i).Tag.Get(ogmTagSuffix) != "" {
			key = t.Field(i).Tag.Get(ogmTagSuffix)
		} else {
			key = t.Field(i).Name
		}
		props[key] = val.Field(i).Interface()
	}
	return props
}

func (rm *ReflectionMapper) performReverseMap(properties core.KVMap, t reflect.Type, val reflect.Value) {
	t = t.Elem()
	for i := 0; i < val.NumField(); i++ {
		if t.Field(i).Tag != "" && t.Field(i).Tag.Get(ogmTagSuffix) != "" {
			val.Field(i).Set(reflect.ValueOf(properties[t.Field(i).Tag.Get(ogmTagSuffix)]))
		} else {
			// TODO: make property name lookup handling more streamlined and configurable
			// by the caller.
			// data stores like Agensgraph convert the property name to lower case
			// and return it as such. In these cases, the reverse lookup would need
			// to happen in case insensitive manner.
			if _, ok := properties[t.Field(i).Name]; ok {
				val.Field(i).Set(reflect.ValueOf(properties[t.Field(i).Name]))
			} else if _, ok = properties[strings.ToLower(t.Field(i).Name)]; ok {
				val.Field(i).Set(reflect.ValueOf(properties[strings.ToLower(t.Field(i).Name)]))
			} else if _, ok := properties[strings.ToUpper(t.Field(i).Name)]; ok {
				val.Field(i).Set(reflect.ValueOf(properties[strings.ToUpper(t.Field(i).Name)]))
			} else {
				// not found - simply continue
				continue
			}

		}
	}
}

func (rm *ReflectionMapper) performDecode(properties core.KVMap, t reflect.Type, val reflect.Value, v any) error {
	fieldTagMapping := make(map[string]string)
	fieldMappingByName := make(map[string]reflect.StructField)

	t = t.Elem()
	for i := 0; i < val.NumField(); i++ {
		if t.Field(i).Tag != "" && t.Field(i).Tag.Get(ogmTagSuffix) != "" {
			fieldTagMapping[t.Field(i).Tag.Get(ogmTagSuffix)] = t.Field(i).Name

		}
		fieldMappingByName[strings.ToLower(t.Field(i).Name)] = t.Field(i)
		fieldMappingByName[strings.ToUpper(t.Field(i).Name)] = t.Field(i)
		fieldMappingByName[t.Field(i).Name] = t.Field(i)
	}
	mapToDecode := make(map[string]interface{})
	for k, v := range properties {
		fieldToDecode, ok := fieldMappingByName[k]
		// if the field is not found by name, then check if the key is a tag on a field
		if !ok {
			originalFieldName, ok := fieldTagMapping[k]
			if !ok {
				return fmt.Errorf("unknown field %s", k)
			}
			fieldToDecode = fieldMappingByName[originalFieldName]
		}

		mapToDecode[fieldToDecode.Name] = v
	}
	err := mapstructure.Decode(mapToDecode, v)
	return err
}

func NewReflectionMapper() *ReflectionMapper {
	return &ReflectionMapper{}
}
