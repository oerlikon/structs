package structs

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errNotExported = errors.New("field is not exported")
	errNotSettable = errors.New("field is not settable")
)

// Field represents a single struct field that encapsulates high level
// functions around the field.
type Field struct {
	value      reflect.Value
	field      reflect.StructField
	defaultTag string
}

// Tag returns the value associated with key in the tag string. If there is no
// such key in the tag, Tag returns the empty string.
func (f *Field) Tag(key string) string {
	return f.field.Tag.Get(key)
}

// Value returns the underlying value of the field. It panics if the field
// is not exported.
func (f *Field) Value() interface{} {
	return f.value.Interface()
}

// IsEmbedded returns true if the given field is an anonymous field (embedded).
func (f *Field) IsEmbedded() bool {
	return f.field.Anonymous
}

// IsExported returns true if the given field is exported.
func (f *Field) IsExported() bool {
	return f.field.PkgPath == ""
}

// IsZero returns true if the given field is not initialized (has a zero value).
// It panics if the field is not exported.
func (f *Field) IsZero() bool {
	zero := reflect.Zero(f.value.Type()).Interface()
	current := f.Value()
	return reflect.DeepEqual(current, zero)
}

// Name returns the name of the given field.
func (f *Field) Name() string {
	return f.field.Name
}

// Kind returns the fields kind, such as "string", "map", "bool", etc ..
func (f *Field) Kind() reflect.Kind {
	return f.value.Kind()
}

// Set sets the field to given value v. It returns an error if the field is not
// settable (not addressable or not exported) or if the given value's type
// is not assignable to the field's type.
func (f *Field) Set(val interface{}) error {
	// we can't set unexported fields, so be sure this field is exported
	if !f.IsExported() {
		return errNotExported
	}
	if !f.value.CanSet() {
		return errNotSettable
	}
	value := reflect.ValueOf(val)
	if !value.Type().AssignableTo(f.value.Type()) {
		return fmt.Errorf("can't assign %s to %s", value.Type(), f.value.Type())
	}
	f.value.Set(value)
	return nil
}

// Zero sets the field to its zero value. It returns an error if the field is not
// settable (not addressable or not exported).
func (f *Field) Zero() error {
	zero := reflect.Zero(f.value.Type()).Interface()
	return f.Set(zero)
}

// Fields returns a slice of Fields. This is particular handy to get the fields
// of a nested struct . A struct tag with the content of "-" ignores the
// checking of that particular field. Example:
//
//	// Field is ignored by this package.
//	Field *http.Request `structs:"-"`
//
// It panics if field is not exported or if field's kind is not struct.
func (f *Field) Fields() []*Field {
	return getFields(f.value, f.defaultTag)
}

// Field returns the field from a nested struct or nil if not found.
func (f *Field) Field(name string) *Field {
	value := &f.value
	// value must be settable so we need to make sure it holds the address of the
	// variable and not a copy, so we can pass the pointer to structVal instead of a
	// copy (which is not assigned to any variable, hence not settable).
	// see "https://blog.golang.org/laws-of-reflection#TOC_8."
	if f.value.Kind() != reflect.Ptr {
		a := f.value.Addr()
		value = &a
	}
	v := structVal(value.Interface())

	field, ok := v.Type().FieldByName(name)
	if !ok {
		return nil
	}

	return &Field{
		field: field,
		value: v.FieldByName(name),
	}
}
