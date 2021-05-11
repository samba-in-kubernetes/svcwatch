/*
Copyright 2021 John Mulligan <phlogistonjohn@asynchrono.us>

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

package statefile

import (
	"encoding/json"
	"os"
)

// HostInfo holds values pertaining to a single address/name pair.
type HostInfo struct {
	Name        string `json:"name"`
	IPv4Address string `json:"ipv4"`
	Target      string `json:"target"`
}

// HostState represents the overall state of the watched addresses.
type HostState struct {
	Reference string     `json:"ref"`
	Items     []HostInfo `json:"items"`
}

// Save a HostState to the file indicated by path.
func (hs HostState) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(hs)
}

// Differs returns true if the other HostState is not equivalent to
// the current HostState.
func (hs HostState) Differs(other HostState) bool {
	if hs.Reference != other.Reference {
		return true
	}
	if len(hs.Items) != len(other.Items) {
		return true
	}
	for i := range hs.Items {
		if hs.Items[i] != other.Items[i] {
			return true
		}
	}
	return false
}
