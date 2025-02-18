// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backupentry

import (
	"context"

	extensionscontroller "github.com/gardener/gardener-extensions/pkg/controller"
	extensions1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/gardener/gardener/pkg/operation/common"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type secretToBackupEntryMapper struct {
	client     client.Client
	predicates []predicate.Predicate
}

func (m *secretToBackupEntryMapper) Map(obj handler.MapObject) []reconcile.Request {
	if obj.Object == nil {
		return nil
	}

	secret, ok := obj.Object.(*corev1.Secret)
	if !ok {
		return nil
	}

	backupEntryList := &extensions1alpha1.BackupEntryList{}
	if err := m.client.List(context.TODO(), client.MatchingField("spec.secretRef.name", secret.Name).MatchingField("spec.secretRef.namespace", secret.Namespace), backupEntryList); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, backupEntry := range backupEntryList.Items {
		if !extensionscontroller.EvalGenericPredicate(&backupEntry, m.predicates...) {
			continue
		}
		requests = append(requests, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: backupEntry.Name,
			},
		})
	}
	return requests
}

// SecretToBackupEntryMapper returns a mapper that returns requests for BackupEntry whose
// referenced secrets have been modified.
func SecretToBackupEntryMapper(client client.Client, predicates []predicate.Predicate) handler.Mapper {
	return &secretToBackupEntryMapper{client, predicates}
}

type namespaceToBackupEntryMapper struct {
	client     client.Client
	predicates []predicate.Predicate
}

func (m *namespaceToBackupEntryMapper) Map(obj handler.MapObject) []reconcile.Request {
	if obj.Object == nil {
		return nil
	}

	namespace, ok := obj.Object.(*corev1.Namespace)
	if !ok {
		return nil
	}

	backupEntryList := &extensions1alpha1.BackupEntryList{}
	if err := m.client.List(context.TODO(), nil, backupEntryList); err != nil {
		return nil
	}

	shootUID := namespace.Annotations[common.ShootUID]
	var requests []reconcile.Request
	for _, backupEntry := range backupEntryList.Items {
		if !extensionscontroller.EvalGenericPredicate(&backupEntry, m.predicates...) {
			continue
		}

		expectedTechnicalID, expectedUID := ExtractShootDetailsFromBackupEntryName(backupEntry.Name)
		if namespace.Name == expectedTechnicalID && shootUID == expectedUID {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name: backupEntry.Name,
				},
			})
		}
	}
	return requests
}

// NamespaceToBackupEntryMapper returns a mapper that returns requests for BackupEntry whose
// associated Shoot's seed namespace have been modified.
func NamespaceToBackupEntryMapper(client client.Client, predicates []predicate.Predicate) handler.Mapper {
	return &namespaceToBackupEntryMapper{client, predicates}
}
