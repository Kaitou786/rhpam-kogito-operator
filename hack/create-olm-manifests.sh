#!/bin/sh
# Copyright 2021 Red Hat, Inc. and/or its affiliates
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.



source ./hack/env.sh

OLM_DIR="build/_output/olm/"
VERSION=$(getOperatorVersion)
MANIFESTS="${OLM_DIR}/${VERSION}"
#cleanup OLM_DIR
rm -rf "${OLM_DIR}"

#create the target version directory
mkdir -p "${MANIFESTS}"

#copy bundle manifests and Dockerfile
cp -r bundle/*  "${MANIFESTS}/"
cp bundle.Dockerfile "${MANIFESTS}/Dockerfile"

#Correct the relative path in Dockerfile
sed -i "s|bundle/manifests|manifests|g"  "${MANIFESTS}/Dockerfile"
sed -i "s|bundle/metadata|metadata|g"    "${MANIFESTS}/Dockerfile"
sed -i "s|bundle/tests|tests|g"          "${MANIFESTS}/Dockerfile"

#removing metrics files, due to KOGITO-4547, once the issue or the upstream issue https://github.com/operator-framework/operator-lifecycle-manager/issues/2019 is resolved we can remove the lines below
rm "${MANIFESTS}/manifests/rhpam-kogito-operator-controller-manager-metrics-service_v1_service.yaml"
rm "${MANIFESTS}/manifests/rhpam-kogito-operator-metrics-reader_rbac.authorization.k8s.io_v1_clusterrole.yaml"