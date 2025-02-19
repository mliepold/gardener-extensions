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

package controlplane

import (
	"github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/aws"
	extensionswebhook "github.com/gardener/gardener-extensions/pkg/webhook"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane"
	"github.com/gardener/gardener-extensions/pkg/webhook/controlplane/genericmutator"

	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var logger = log.Log.WithName("aws-controlplane-webhook")

// AddToManager creates a webhook and adds it to the manager.
func AddToManager(mgr manager.Manager) (webhook.Webhook, error) {
	logger.Info("Adding webhook to manager")
	fciCodec := controlplane.NewFileContentInlineCodec()
	return controlplane.Add(mgr, controlplane.AddArgs{
		Kind:     extensionswebhook.ShootKind,
		Provider: aws.Type,
		Types:    []runtime.Object{&appsv1.Deployment{}, &extensionsv1alpha1.OperatingSystemConfig{}},
		Mutator: genericmutator.NewMutator(NewEnsurer(logger), controlplane.NewUnitSerializer(),
			controlplane.NewKubeletConfigCodec(fciCodec), fciCodec, logger),
	})
}
