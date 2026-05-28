// Copyright Project Harbor Authors
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

package rbac

import (
	"context"
	"sync"

	"github.com/casbin/casbin/v3"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

var _ evaluator.Evaluator = &Evaluator{}

// Evaluator the permission evaluator for rbac user
type Evaluator struct {
	rbacUser types.RBACUser
	enforcer *casbin.Enforcer
	err      error
	once     sync.Once
}

// HasPermission returns true when the rbac user has action permission for the resource
func (e *Evaluator) HasPermission(_ context.Context, resource types.Resource, action types.Action) bool {
	e.once.Do(func() {
		e.enforcer, e.err = makeEnforcer(e.rbacUser)
	})
	if e.err != nil {
		log.Errorf("failed to initialise RBAC enforcer: %v", e.err)
		return false
	}

	allowed, err := e.enforcer.Enforce(e.rbacUser.GetUserName(), resource.String(), action.String())
	if err != nil {
		log.Errorf("failed to evaluate RBAC permission: %v", err)
		return false
	}
	return allowed
}

// New returns evaluator.Evaluator for the RBACUser
func New(rbacUser types.RBACUser) *Evaluator {
	return &Evaluator{
		rbacUser: rbacUser,
	}
}
