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
	"testing"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()

	ref := types.ManagedObjectReference{Type: "Test", Value: "Test"}
	f := &mo.Folder{}
	f.Self = ref
	r.PutEntity(nil, f)

	e := r.Get(ref)

	if e.Reference() != ref {
		t.Fail()
	}

	r.Remove(ref)

	if r.Get(ref) != nil {
		t.Fail()
	}

	r.Put(e)
	e = r.Get(ref)

	if e.Reference() != ref {
		t.Fail()
	}
}
