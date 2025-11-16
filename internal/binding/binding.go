// Copyright 2025 coregx. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package binding provides request body binding functionality.
//
// This package is internal and not part of the public API.
// External code should use Context.Bind() method instead of calling this package directly.
package binding

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

// Common binding errors.
var (
	// ErrUnsupportedMediaType is returned when Content-Type is not supported.
	ErrUnsupportedMediaType = errors.New("unsupported media type")

	// ErrEmptyRequestBody is returned when request body is required but empty.
	ErrEmptyRequestBody = errors.New("request body is empty")

	// ErrInvalidContentType is returned when Content-Type header is malformed.
	ErrInvalidContentType = errors.New("invalid content-type header")
)

// Binder is the interface for request body binding.
type Binder interface {
	// Bind binds the request body to the given struct pointer.
	Bind(*http.Request, any) error
}

// JSON binder for application/json.
type jsonBinder struct{}

func (jsonBinder) Bind(req *http.Request, obj any) error {
	if req.Body == nil || req.ContentLength == 0 {
		return ErrEmptyRequestBody
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyRequestBody
		}
		return fmt.Errorf("json decode error: %w", err)
	}

	return nil
}

// XML binder for application/xml.
type xmlBinder struct{}

func (xmlBinder) Bind(req *http.Request, obj any) error {
	if req.Body == nil || req.ContentLength == 0 {
		return ErrEmptyRequestBody
	}

	decoder := xml.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEmptyRequestBody
		}
		return fmt.Errorf("xml decode error: %w", err)
	}

	return nil
}

// Form binder for application/x-www-form-urlencoded.
type formBinder struct{}

func (formBinder) Bind(req *http.Request, obj any) error {
	if err := req.ParseForm(); err != nil {
		return fmt.Errorf("parse form error: %w", err)
	}

	if len(req.Form) == 0 {
		return ErrEmptyRequestBody
	}

	return mapForm(obj, req.Form)
}

// Multipart form binder for multipart/form-data.
type multipartBinder struct{}

func (multipartBinder) Bind(req *http.Request, obj any) error {
	if err := req.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
		return fmt.Errorf("parse multipart form error: %w", err)
	}

	if req.MultipartForm == nil || len(req.MultipartForm.Value) == 0 {
		return ErrEmptyRequestBody
	}

	return mapForm(obj, req.MultipartForm.Value)
}

// mapForm maps form values to struct fields.
func mapForm(ptr any, form map[string][]string) error {
	val := reflect.ValueOf(ptr)
	if val.Kind() != reflect.Ptr {
		return errors.New("binding element must be a pointer")
	}

	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		structField := typ.Field(i)

		// Get form tag or use field name
		formTag := structField.Tag.Get("form")
		if formTag == "" {
			formTag = structField.Name
		}

		// Skip if tag is "-"
		if formTag == "-" {
			continue
		}

		// Get value from form
		values, ok := form[formTag]
		if !ok || len(values) == 0 {
			continue
		}

		// Set field value
		if err := setField(field, values[0]); err != nil {
			return fmt.Errorf("set field %s error: %w", structField.Name, err)
		}
	}

	return nil
}

// setField sets a struct field value from string.
func setField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
		return nil

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
		return nil

	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
		return nil

	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
}

// Default binders for each content type.
var (
	jsonBinding      = jsonBinder{}
	xmlBinding       = xmlBinder{}
	formBinding      = formBinder{}
	multipartBinding = multipartBinder{}
)

// GetBinder returns the appropriate binder for the given Content-Type.
func GetBinder(contentType string) (Binder, error) {
	// Extract base content type (ignore charset, boundary, etc.)
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	contentType = strings.TrimSpace(contentType)

	switch contentType {
	case "", "application/json":
		return jsonBinding, nil
	case "application/xml", "text/xml":
		return xmlBinding, nil
	case "application/x-www-form-urlencoded":
		return formBinding, nil
	case "multipart/form-data":
		return multipartBinding, nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedMediaType, contentType)
	}
}

// Bind binds the request body to obj based on Content-Type.
// This is a convenience function that gets the appropriate binder and calls Bind.
func Bind(req *http.Request, obj any) error {
	contentType := req.Header.Get("Content-Type")

	binder, err := GetBinder(contentType)
	if err != nil {
		return err
	}

	return binder.Bind(req, obj)
}
