// Copyright 2016 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package simulator

import (
	"errors"
	"path"
	"reflect"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

type PropertyCollector struct {
	mo.PropertyCollector
}

func NewPropertyCollector(ref types.ManagedObjectReference) object.Reference {
	s := &PropertyCollector{}
	s.Self = ref
	return s
}

var errMissingField = errors.New("missing field")
var errEmptyField = errors.New("empty field")

func fieldValueInterface(f reflect.StructField, rval reflect.Value) interface{} {
	pval := rval.Interface()

	if rval.Kind() == reflect.Slice {
		// Convert slice to types.ArrayOf*
		switch v := pval.(type) {
		case []string:
			pval = &types.ArrayOfString{
				String: v,
			}
		case []int32:
			pval = &types.ArrayOfInt{
				Int: v,
			}
		default:
			kind := f.Type.Elem().Name()
			akind, _ := typeFunc("ArrayOf" + kind)
			a := reflect.New(akind)
			a.Elem().FieldByName(kind).Set(rval)
			pval = a.Interface()
		}
	}

	return pval
}

func fieldValue(rval reflect.Value, p string) (interface{}, error) {
	if rval.Type().Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	var value interface{}

	fields := strings.Split(p, ".")

	for i, name := range fields {
		x := ucFirst(name)
		val := rval.FieldByName(x)
		if !val.IsValid() {
			return nil, errMissingField
		}

		if isEmpty(val) {
			return nil, errEmptyField
		}

		if i == len(fields)-1 {
			ftype, _ := rval.Type().FieldByName(x)
			value = fieldValueInterface(ftype, val)
			break
		}

		rval = val
	}

	return value, nil
}

func fieldRefs(f interface{}) []types.ManagedObjectReference {
	switch fv := f.(type) {
	case types.ManagedObjectReference:
		return []types.ManagedObjectReference{fv}
	case *types.ManagedObjectReference:
		return []types.ManagedObjectReference{*fv}
	case *types.ArrayOfManagedObjectReference:
		return fv.ManagedObjectReference
	case nil:
		// empty field
	}

	return nil
}

func isEmpty(rval reflect.Value) bool {
	switch rval.Kind() {
	case reflect.Ptr:
		return rval.IsNil()
	case reflect.String, reflect.Slice:
		return rval.Len() == 0
	}

	return false
}

func isTrue(v *bool) bool {
	return v != nil && *v
}

func isFalse(v *bool) bool {
	return v == nil || *v == false
}

func lcFirst(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func ucFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}

type retrieveResult struct {
	*types.RetrieveResult
	req       *types.RetrievePropertiesEx
	recurse   map[string]bool
	collected map[types.ManagedObjectReference]bool
}

func (rr *retrieveResult) collectAll(rval reflect.Value, rtype reflect.Type, content *types.ObjectContent) {
	for i := 0; i < rval.NumField(); i++ {
		val := rval.Field(i)

		if isEmpty(val) {
			continue
		}

		f := rtype.Field(i)

		if f.Anonymous {
			// recurse into embedded field
			rr.collectAll(val, f.Type, content)
			continue
		}

		content.PropSet = append(content.PropSet, types.DynamicProperty{
			Name: lcFirst(f.Name),
			Val:  fieldValueInterface(f, val),
		})
	}
}

func (rr *retrieveResult) collectFields(rval reflect.Value, fields []string, content *types.ObjectContent) []types.ManagedObjectReference {
	var refs []types.ManagedObjectReference

	for _, name := range fields {
		val, err := fieldValue(rval, name)
		if err == nil {
			if rr.recurse[name] {
				refs = append(refs, fieldRefs(val)...)
			}

			prop := types.DynamicProperty{
				Name: name,
				Val:  val,
			}

			content.PropSet = append(content.PropSet, prop)
			continue
		}

		switch err {
		case errEmptyField:
			// ok
		case errMissingField:
			content.MissingSet = append(content.MissingSet, types.MissingProperty{
				Path: name,
				Fault: types.LocalizedMethodFault{Fault: &types.InvalidProperty{
					Name: name,
				}},
			})
		}
	}

	return refs
}

func (rr *retrieveResult) collect(ref types.ManagedObjectReference) {
	if rr.collected[ref] {
		return
	}

	obj := Map.Get(ref)

	content := types.ObjectContent{
		Obj: ref,
	}

	rval := reflect.ValueOf(obj).Elem()
	rtype := rval.Type()

	// PropertyCollector is for Managed Object types only (package mo).
	// If the registry object is not in the mo package, assume it is a wrapper
	// type where the first field is an embedded mo type.
	// Otherwise, PropSet.All will not work as expected.
	if path.Base(rtype.PkgPath()) != "mo" {
		rval = rval.Field(0)
		rtype = rval.Type()
	}

	var refs []types.ManagedObjectReference

	for _, spec := range rr.req.SpecSet {
		for _, p := range spec.PropSet {
			if p.Type != ref.Type {
				// e.g. ManagedEntity, ComputeResource
				field, ok := rtype.FieldByName(p.Type)

				if !(ok && field.Anonymous) {
					continue
				}
			}

			if isTrue(p.All) {
				rr.collectAll(rval, rtype, &content)
				continue
			}

			refs = append(refs, rr.collectFields(rval, p.PathSet, &content)...)
		}
	}

	rr.Objects = append(rr.Objects, content)
	rr.collected[ref] = true

	for _, rref := range refs {
		rr.collect(rref)
	}
}

func (pc *PropertyCollector) collect(r *types.RetrievePropertiesEx) (*types.RetrieveResult, types.BaseMethodFault) {
	var refs []types.ManagedObjectReference

	rr := &retrieveResult{
		RetrieveResult: &types.RetrieveResult{},
		req:            r,
		recurse:        make(map[string]bool),
		collected:      make(map[types.ManagedObjectReference]bool),
	}

	// Select object references
	for _, spec := range r.SpecSet {
		for _, o := range spec.ObjectSet {
			obj := Map.Get(o.Obj)
			if obj == nil {
				if isFalse(spec.ReportMissingObjectsInResults) {
					return nil, &types.ManagedObjectNotFound{Obj: o.Obj}
				}
				continue
			}

			if o.SelectSet == nil || isFalse(o.Skip) {
				refs = append(refs, o.Obj)
			}

			for _, ss := range o.SelectSet {
				ts := ss.(*types.TraversalSpec)

				if ts.SelectSet != nil {
					rr.recurse[ts.Path] = true
				}

				rval := reflect.ValueOf(obj)

				f, _ := fieldValue(rval, ts.Path)

				refs = append(refs, fieldRefs(f)...)
			}
		}
	}

	for _, ref := range refs {
		rr.collect(ref)
	}

	return rr.RetrieveResult, nil
}

func (pc *PropertyCollector) RetrievePropertiesEx(r *types.RetrievePropertiesEx) soap.HasFault {
	body := &methods.RetrievePropertiesExBody{}

	res, fault := pc.collect(r)

	if fault != nil {
		body.Fault_ = Fault("", fault)
	} else {
		body.Res = &types.RetrievePropertiesExResponse{
			Returnval: res,
		}
	}

	return body
}

// RetrieveProperties is deprecated, but govmomi is still using it at the moment.
func (pc *PropertyCollector) RetrieveProperties(r *types.RetrieveProperties) soap.HasFault {
	body := &methods.RetrievePropertiesBody{}

	res := pc.RetrievePropertiesEx(&types.RetrievePropertiesEx{
		This:    r.This,
		SpecSet: r.SpecSet,
	})

	if res.Fault() != nil {
		body.Fault_ = res.Fault()
	} else {
		body.Res = &types.RetrievePropertiesResponse{
			Returnval: res.(*methods.RetrievePropertiesExBody).Res.Returnval.Objects,
		}
	}

	return body
}
