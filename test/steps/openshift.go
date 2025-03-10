// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package steps

import (
	"fmt"

	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-operator/test/framework"
	v1 "github.com/kiegroup/rhpam-kogito-operator/api/v1"
	"k8s.io/apimachinery/pkg/types"
)

func registerOpenShiftSteps(ctx *godog.ScenarioContext, data *Data) {
	// BuildConfig steps
	ctx.Step(`^BuildConfig "([^"]*)" is created with webhooks within (\d+) minutes$`, data.buildConfigHasWebhooksWithinMinutes)
}

func (data *Data) buildConfigHasWebhooksWithinMinutes(buildConfigName string, timeoutInMin int) error {
	kogitoBuild := &v1.KogitoBuild{}
	exists, err := framework.GetObjectWithKey(types.NamespacedName{Namespace: data.Namespace, Name: buildConfigName}, kogitoBuild)

	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("KogitoBuild with name %s doesn't exist", buildConfigName)
	}

	return framework.WaitForBuildConfigCreatedWithWebhooks(data.Namespace, buildConfigName, kogitoBuild.Spec.GetWebHooks(), timeoutInMin)
}
