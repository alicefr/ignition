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
	"fmt"
	"net/url"
	"strings"

	"github.com/coreos/ignition/v2/config/v3_6_experimental/types"
	"github.com/coreos/ignition/v2/internal/log"
	"github.com/coreos/ignition/v2/internal/resource"
)

// resolveTemplates processes template variables in the configuration by fetching
// remote values and replacing all occurrences of the template tags throughout the config.
func resolveTemplates(cfg types.Config, fetcher *resource.Fetcher, logger *log.Logger) (types.Config, error) {
	if len(cfg.Template) == 0 {
		return cfg, nil
	}

	logger.Info("processing %d template variables", len(cfg.Template))

	// Fetch template values
	templateValues := make(map[string]string)
	for _, template := range cfg.Template {
		logger.Info("fetching template variable %q from %q", template.Tag, template.Remote.URL)

		parsedURL, err := url.Parse(template.Remote.URL)
		if err != nil {
			return cfg, fmt.Errorf("invalid template URL %q: %v", template.Remote.URL, err)
		}

		// Create basic fetch options for HTTP GET
		opts := resource.FetchOptions{}

		// Fetch the value
		data, err := fetcher.FetchToBuffer(*parsedURL, opts)
		if err != nil {
			return cfg, fmt.Errorf("failed to fetch template variable %q from %q: %v", template.Tag, template.Remote.URL, err)
		}

		value := strings.TrimSpace(string(data))
		templateValues[template.Tag] = value
		logger.Info("template variable %q resolved to %q", template.Tag, value)
	}

	resolvedConfig, err := replaceTemplateVariables(cfg, templateValues)
	if err != nil {
		return cfg, fmt.Errorf("failed to replace template variables: %v", err)
	}

	resolvedConfig.Template = nil

	logger.Info("template processing completed")
	return resolvedConfig, nil
}

// replaceTemplateVariables performs string replacement of template variables throughout
// the entire configuration by marshaling to JSON, replacing, and unmarshaling back.
func replaceTemplateVariables(cfg types.Config, templateValues map[string]string) (types.Config, error) {
	// Marshal the config to JSON
	configJSON, err := json.Marshal(cfg)
	if err != nil {
		return cfg, fmt.Errorf("failed to marshal config: %v", err)
	}

	configStr := string(configJSON)

	for tag, value := range templateValues {
		// Replace ${tag} format
		configStr = strings.ReplaceAll(configStr, "${"+tag+"}", value)
	}

	var resolvedConfig types.Config
	if err := json.Unmarshal([]byte(configStr), &resolvedConfig); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal resolved config: %v", err)
	}

	return resolvedConfig, nil
}
