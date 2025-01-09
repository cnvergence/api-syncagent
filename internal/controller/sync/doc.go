/*
Copyright 2025 The KCP Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Package sync contains a controller that watches the APIExport we manage
in the platform cluster. Once the virtual workspace URL for said APIExport
is ready, the controller will begin to synchronize resources back and forth
between the platform cluster (i.e. all relevant workspaces) and the service
cluster.
*/
package sync
