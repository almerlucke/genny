package template

import (
	"reflect"
)

type Parameter struct {
	Name  string
	Value any
}

func NewParameter(name string, val any) *Parameter {
	return &Parameter{Name: name, Value: val}
}

// intermediateValue is created in prototype to check if prototype needs to split if value is an array or slice
type intermediateValue struct {
	value        any
	reflectValue reflect.Value
	length       int
	indexAble    bool
}

func newIntermediateValue(v any) *intermediateValue {
	reflectValue := reflect.ValueOf(v)
	kind := reflectValue.Kind()
	length := 1
	indexAble := false

	if kind == reflect.Slice || kind == reflect.Array {
		length = reflectValue.Len()
		indexAble = true
	}

	return &intermediateValue{value: v, reflectValue: reflectValue, length: length, indexAble: indexAble}
}

func (iv *intermediateValue) index(i int) any {
	// return normal value if not indexAble
	if !iv.indexAble {
		return iv.value
	}

	// clip index to length
	if i >= iv.length {
		i = iv.length - 1
	}

	// return element at index
	return iv.reflectValue.Index(i).Interface()
}

// Template is a template of a map. When Generate() is called, a deep copy of the template is made with all Generator values
// in the template replaced with the Generate() from that Generator. In the deep copy all parameter entries are replaced
// with the matching replacement values. If the value from a Generator is a slice/array, the output of the template is split into multiple
// maps based on the longest slice/array value found
type Template map[string]any

func (t Template) intermediate() Template {
	tc := Template{}

	for k, v := range t {
		switch vt := v.(type) {
		case Template:
			tc[k] = vt.intermediate()
		case *Parameter:
			tc[k] = newIntermediateValue(vt.Value)
		case Parameter:
			tc[k] = newIntermediateValue(vt.Value)
		default:
			if g, ok := generate(v); ok {
				tc[k] = newIntermediateValue(g)
			} else {
				tc[k] = v
			}
		}
	}

	return tc
}

func (t Template) length() int {
	lmax := 1

	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			tmax := vt.length()
			if tmax > lmax {
				lmax = tmax
			}
		case *intermediateValue:
			if vt.length > lmax {
				lmax = vt.length
			}
		default:
			break
		}
	}

	return lmax
}

func (t Template) mapAtIndex(i int) map[string]any {
	m := map[string]any{}

	for k, v := range t {
		switch vt := v.(type) {
		case Template:
			m[k] = vt.mapAtIndex(i)
		case *intermediateValue:
			m[k] = vt.index(i)
		default:
			m[k] = v
		}
	}

	return m
}

func (t Template) SetParameter(name string, val any) {
	t.SetParameters([]*Parameter{NewParameter(name, val)})
}

func (t Template) SetParameters(parameters []*Parameter) {
	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			vt.SetParameters(parameters)
		case *Parameter:
			for _, parameter := range parameters {
				if vt.Name == parameter.Name {
					vt.Value = parameter.Value
					break
				}
			}
		case Parameter:
			for _, parameter := range parameters {
				if vt.Name == parameter.Name {
					vt.Value = parameter.Value
					break
				}
			}
		default:
			break
		}
	}
}

func (t Template) Generate() []map[string]any {
	intermediate := t.intermediate()
	length := intermediate.length()
	maps := make([]map[string]any, length)

	for i := 0; i < length; i++ {
		maps[i] = intermediate.mapAtIndex(i)
	}

	return maps
}

func (t Template) Continuous() bool {
	continuous := true

	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			continuous = vt.Continuous()
		default:
			continuous = boolForMethod(vt, "Continuous", true)
		}

		if !continuous {
			break
		}
	}

	return continuous
}

func (t Template) Reset() {
	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			vt.Reset()
		default:
			reset(vt)
		}
	}
}

func (t Template) Done() bool {
	for _, v := range t {
		switch vt := v.(type) {
		case Template:
			if vt.Done() {
				return true
			}
		default:
			if boolForMethod(vt, "Done", false) {
				return true
			}
		}
	}

	return false
}

/*
Reflective functions to call methods of Generator interface for any object that implements a type of
Generator[T]
*/
func generate(va any) (any, bool) {
	v := reflect.ValueOf(va)

	m := v.MethodByName("Generate")
	if !m.IsValid() {
		return va, false
	}

	mt := m.Type()
	if mt.NumIn() != 0 || mt.NumOut() != 1 {
		return va, false
	}

	return m.Call(nil)[0].Interface(), true
}

func boolForMethod(va any, method string, defaultValue bool) bool {
	v := reflect.ValueOf(va)

	m := v.MethodByName(method)
	if !m.IsValid() {
		return defaultValue
	}

	mt := m.Type()
	if mt.NumIn() != 0 || mt.NumOut() != 1 {
		return defaultValue
	}

	return m.Call(nil)[0].Interface().(bool)
}

func reset(va any) {
	v := reflect.ValueOf(va)

	m := v.MethodByName("Reset")
	if !m.IsValid() {
		return
	}

	mt := m.Type()
	if mt.NumIn() != 0 || mt.NumOut() != 0 {
		return
	}

	_ = m.Call(nil)
}
