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

package framework

import (
	"fmt"

	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	v1 "github.com/kiegroup/rhpam-kogito-operator/api/v1"

	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/test/config"
	kogitoFramework "github.com/kiegroup/kogito-operator/test/framework"
	bddtypes "github.com/kiegroup/kogito-operator/test/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeployKogitoBuild deploy a KogitoBuild
func DeployKogitoBuild(namespace string, buildHolder *bddtypes.KogitoBuildHolder) error {
	kogitoFramework.GetLogger(namespace).Info(fmt.Sprintf("Deploy %s example %s with name %s and native %v", buildHolder.KogitoBuild.GetSpec().GetRuntime(), buildHolder.KogitoBuild.GetSpec().GetGitSource().GetContextDir(), buildHolder.KogitoBuild.GetName(), buildHolder.KogitoBuild.GetSpec().IsNative()))

	if err := kogitoFramework.CreateObject(buildHolder.KogitoBuild); err != nil {
		return fmt.Errorf("Error creating example build %s: %v", buildHolder.KogitoBuild.GetName(), err)
	}
	if err := kogitoFramework.CreateObject(buildHolder.KogitoService); err != nil {
		return fmt.Errorf("Error creating example service %s: %v", buildHolder.KogitoService.GetName(), err)
	}
	return nil
}

// GetKogitoBuildStub Get basic KogitoBuild stub with all needed fields initialized
func GetKogitoBuildStub(namespace, runtimeType, name string) *v1.KogitoBuild {
	kogitoBuild := &v1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1.KogitoBuildSpec{
			Runtime:        api.RuntimeType(runtimeType),
			MavenMirrorURL: config.GetMavenMirrorURL(),
		},
	}

	if len(config.GetCustomMavenRepoURL()) > 0 {
		kogitoBuild.Spec.Env = framework.EnvOverride(kogitoBuild.Spec.Env, corev1.EnvVar{Name: "MAVEN_REPO_URL", Value: config.GetCustomMavenRepoURL()})
	}

	if config.IsMavenIgnoreSelfSignedCertificate() {
		kogitoBuild.Spec.Env = framework.EnvOverride(kogitoBuild.Spec.Env, corev1.EnvVar{Name: "MAVEN_IGNORE_SELF_SIGNED_CERTIFICATE", Value: "true"})
	}

	return kogitoBuild
}

// SetupKogitoBuildImageStreams sets the correct images for the KogitoBuild
func SetupKogitoBuildImageStreams(kogitoBuild *v1.KogitoBuild) {
	kogitoBuild.Spec.BuildImage = getKogitoBuildS2IImage()
	kogitoBuild.Spec.RuntimeImage = getKogitoBuildRuntimeImage(kogitoBuild)
}

func getKogitoBuildS2IImage() string {
	if len(config.GetBuildS2IImageStreamTag()) > 0 {
		return config.GetBuildS2IImageStreamTag()
	}

	return getKogitoBuildImage(infrastructure.KogitoBuilderImage)
}

func getKogitoBuildRuntimeImage(kogitoBuild *v1.KogitoBuild) string {
	var imageName string
	if len(config.GetBuildRuntimeImageStreamTag()) > 0 {
		return config.GetBuildRuntimeImageStreamTag()
	}
	if kogitoBuild.Spec.Native {
		imageName = infrastructure.KogitoRuntimeNative
	} else {
		imageName = infrastructure.KogitoRuntimeJVM
	}
	return getKogitoBuildImage(imageName)
}

// getKogitoBuildImage returns a build image with defaults set
func getKogitoBuildImage(imageName string) string {
	image := api.Image{
		Domain:    config.GetBuildImageRegistry(),
		Namespace: config.GetBuildImageNamespace(),
		Tag:       config.GetBuildImageVersion(),
	}

	// Update image name with suffix if provided
	if len(config.GetBuildImageNameSuffix()) > 0 {
		image.Name = fmt.Sprintf("%s-%s", imageName, config.GetBuildImageNameSuffix())
	}
	return framework.ConvertImageToImageTag(image)
}
