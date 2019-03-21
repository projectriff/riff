// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

// MediaType is an enumeration of the supported mime types that an element of an image might have.
type MediaType string

// The collection of known MediaType values.
const (
	OCIContentDescriptor           MediaType = "application/vnd.oci.descriptor.v1+json"
	OCIImageIndex                  MediaType = "application/vnd.oci.image.index.v1+json"
	OCIManifestSchema1             MediaType = "application/vnd.oci.image.manifest.v1+json"
	OCIConfigJSON                  MediaType = "application/vnd.oci.image.config.v1+json"
	OCILayer                       MediaType = "application/vnd.oci.image.layer.v1.tar+gzip"
	OCIRestrictedLayer             MediaType = "application/vnd.oci.image.layer.nondistributable.v1.tar+gzip"
	OCIUncompressedLayer           MediaType = "application/vnd.oci.image.layer.v1.tar"
	OCIUncompressedRestrictedLayer MediaType = "application/vnd.oci.image.layer.nondistributable.v1.tar"

	DockerManifestSchema1       MediaType = "application/vnd.docker.distribution.manifest.v1+json"
	DockerManifestSchema1Signed MediaType = "application/vnd.docker.distribution.manifest.v1+prettyjws"
	DockerManifestSchema2       MediaType = "application/vnd.docker.distribution.manifest.v2+json"
	DockerManifestList          MediaType = "application/vnd.docker.distribution.manifest.list.v2+json"
	DockerLayer                 MediaType = "application/vnd.docker.image.rootfs.diff.tar.gzip"
	DockerConfigJSON            MediaType = "application/vnd.docker.container.image.v1+json"
	DockerPluginConfig          MediaType = "application/vnd.docker.plugin.v1+json"
	DockerForeignLayer          MediaType = "application/vnd.docker.image.rootfs.foreign.diff.tar.gzip"
	DockerUncompressedLayer     MediaType = "application/vnd.docker.image.rootfs.diff.tar"
)
