// Package mcpx provides utilities for MCP extensions.
package mcpx

import (
	"fmt"
	"reflect"
	"strings"
)

// MapArguments copies properties from args into v.
//
// The target must be a pointer to a struct whose fields carry `param` tags.
// Tags have the form `param:"name[,required]"`.  If args contains a key
// with no matching field, if a required value is missing, or if a value’s
// type is not assignable to the field, an error is returned.
func MapArguments(args map[string]any, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be *struct")
	}
	rv = rv.Elem()
	rt := rv.Type()

	fields := make(map[string]int)
	required := make(map[string]bool)
	for i := range rt.NumField() {
		f := rt.Field(i)
		tag := f.Tag.Get("param")
		if tag == "" {
			continue
		}
		name, flags, _ := strings.Cut(tag, ",")
		if name == "" {
			name = strings.ToLower(f.Name)
		}
		fields[name] = i
		if flags == "require" {
			required[name] = true
		}
	}

	set := make(map[string]bool)
	for k, val := range args {
		idx, ok := fields[k]
		if !ok {
			return fmt.Errorf("unknown parameter %q", k)
		}
		fv := rv.Field(idx)
		if !fv.CanSet() {
			return fmt.Errorf("cannot set field %q", k)
		}
		vv := reflect.ValueOf(val)
		if !vv.Type().AssignableTo(fv.Type()) {
			return fmt.Errorf("parameter %q expects %s, got %s", k, fv.Type(), vv.Type())
		}
		fv.Set(vv)
		set[k] = true
	}

	for k := range required {
		if !set[k] {
			return fmt.Errorf("missing required parameter %q", k)
		}
	}
	return nil
}
