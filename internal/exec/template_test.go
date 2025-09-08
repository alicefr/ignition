// Copyright 2025 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"encoding/json"
	"testing"

	"github.com/coreos/ignition/v2/config/v3_6_experimental/types"
)

func TestReplaceTemplateVariables(t *testing.T) {
	tests := []struct {
		name           string
		config         types.Config
		templateValues map[string]string
		expected       string
	}{
		{
			name: "Replace template variable in file contents",
			config: types.Config{
				Storage: types.Storage{
					Files: []types.File{
						{
							Node: types.Node{
								Path: "/test/file",
							},
							FileEmbedded1: types.FileEmbedded1{
								Contents: types.Resource{
									Source: stringPtr("data:,node-${id}"),
								},
							},
						},
					},
				},
			},
			templateValues: map[string]string{
				"id": "worker-001",
			},
			expected: "node-worker-001",
		},
		{
			name: "Replace multiple template variables",
			config: types.Config{
				Storage: types.Storage{
					Files: []types.File{
						{
							Node: types.Node{
								Path: "/test/${env}/config",
							},
							FileEmbedded1: types.FileEmbedded1{
								Contents: types.Resource{
									Source: stringPtr("data:,${env}-${id}"),
								},
							},
						},
					},
				},
			},
			templateValues: map[string]string{
				"id":  "001",
				"env": "production",
			},
			expected: "production-001",
		},
		{
			name: "No template variables",
			config: types.Config{
				Storage: types.Storage{
					Files: []types.File{
						{
							Node: types.Node{
								Path: "/test/file",
							},
							FileEmbedded1: types.FileEmbedded1{
								Contents: types.Resource{
									Source: stringPtr("data:,static-content"),
								},
							},
						},
					},
				},
			},
			templateValues: map[string]string{},
			expected:       "static-content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := replaceTemplateVariables(tt.config, tt.templateValues)
			if err != nil {
				t.Fatalf("replaceTemplateVariables failed: %v", err)
			}

			if len(result.Storage.Files) == 0 {
				t.Fatal("Expected at least one file in result")
			}

			file := result.Storage.Files[0]
			if file.Contents.Source == nil {
				t.Fatal("Expected file contents to be present")
			}

			// Extract the data URL content
			source := *file.Contents.Source
			if source[:5] != "data:" {
				t.Fatalf("Expected data URL, got: %s", source)
			}

			// Extract content after "data:,"
			content := source[6:] // Skip "data:,"
			if content != tt.expected {
				t.Errorf("Expected content %q, got %q", tt.expected, content)
			}
		})
	}
}

func TestConfigMarshaling(t *testing.T) {
	// Test that Template types can be marshaled and unmarshaled
	config := types.Config{
		Ignition: types.Ignition{
			Version: "3.6.0-experimental",
		},
		Template: []types.Template{
			{
				Tag: "id",
				Remote: types.TemplateRemote{
					URL: "http://example.com/node-id",
				},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	// Unmarshal back
	var unmarshaled types.Config
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	// Verify template was preserved
	if len(unmarshaled.Template) != 1 {
		t.Fatalf("Expected 1 template, got %d", len(unmarshaled.Template))
	}

	template := unmarshaled.Template[0]
	if template.Tag != "id" {
		t.Errorf("Expected tag 'id', got %q", template.Tag)
	}

	if template.Remote.URL != "http://example.com/node-id" {
		t.Errorf("Expected URL 'http://example.com/node-id', got %q", template.Remote.URL)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}