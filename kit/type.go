// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kit

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"reflect"
	"strings"
)

// Type provides JSON, XML marshal / unmarshal with encoding of underlying
// type name using kit.Types type name registry
type Type struct {
	T reflect.Type
}

// ShortTypeName returns short package-qualified name of the type: package dir + "." + type name
func (k Type) ShortTypeName() string {
	return Types.TypeName(k.T)
}

// String satisfies the stringer interface
func String(k Type) string {
	if k.T == nil {
		return "nil"
	}
	return k.ShortTypeName()
}

// MarshalJSON saves only the type name
func (k Type) MarshalJSON() ([]byte, error) {
	if k.T == nil {
		b := []byte("null")
		return b, nil
	}
	nm := "\"" + k.ShortTypeName() + "\""
	b := []byte(nm)
	return b, nil
}

// UnmarshalJSON loads the type name and looks it up in the Types registry of type names
func (k *Type) UnmarshalJSON(b []byte) error {
	if bytes.Equal(b, []byte("null")) {
		k.T = nil
		return nil
	}
	tn := string(bytes.Trim(bytes.TrimSpace(b), "\""))
	// fmt.Printf("loading type: %v", tn)
	typ := Types.Type(tn)
	if typ == nil {
		return fmt.Errorf("Type UnmarshalJSON: Types type name not found: %v", tn)
	}
	k.T = typ
	return nil
}

// todo: try to save info as an attribute within a single element instead of
// full start/end

// MarshalXML saves only the type name
func (k Type) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	tokens := []xml.Token{start}
	if k.T == nil {
		tokens = append(tokens, xml.CharData("null"))
	} else {
		tokens = append(tokens, xml.CharData(k.ShortTypeName()))
	}
	tokens = append(tokens, xml.EndElement{start.Name})
	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}
	err := e.Flush()
	if err != nil {
		return err
	}
	return nil
}

// UnmarshalXML loads the type name and looks it up in the Types registry of type names
func (k *Type) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	ct, ok := t.(xml.CharData)
	if ok {
		tn := string(bytes.TrimSpace([]byte(ct)))
		if tn == "null" {
			k.T = nil
		} else {
			// fmt.Printf("loading type: %v\n", tn)
			typ := Types.Type(tn)
			if typ == nil {
				return fmt.Errorf("Type UnmarshalXML: Types type name not found: %v", tn)
			}
			k.T = typ
		}
		t, err := d.Token()
		if err != nil {
			return err
		}
		et, ok := t.(xml.EndElement)
		if ok {
			if et.Name != start.Name {
				return fmt.Errorf("Type UnmarshalXML: EndElement: %v does not match StartElement: %v", et.Name, start.Name)
			}
			return nil
		}
		return fmt.Errorf("Type UnmarshalXML: Token: %+v is not expected EndElement", et)
	}
	return fmt.Errorf("Type UnmarshalXML: Token: %+v is not expected EndElement", ct)
}

// StructTags returns a map[string]string of the tag string from a reflect.StructTag value
// e.g., from StructField.Tag
func StructTags(tags reflect.StructTag) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	flds := strings.Fields(string(tags))
	smap := make(map[string]string, len(flds))
	for _, fld := range flds {
		cli := strings.Index(fld, ":")
		if cli < 0 || len(fld) < cli+3 {
			continue
		}
		vl := strings.TrimSuffix(fld[cli+2:], `"`)
		smap[fld[:cli]] = vl
	}
	return smap
}

// StringJSON returns a JSON representation of item, as a string
// e.g., for printing / debugging etc.
func StringJSON(it any) string {
	b, _ := json.MarshalIndent(it, "", "  ")
	return string(b)
}

// TypeFor returns the [reflect.Type] that represents the type argument T.
// It is a copy of [reflect.TypeFor], which will likely be added in Go 1.22
// (see https://github.com/golang/go/issues/60088)
func TypeFor[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}
