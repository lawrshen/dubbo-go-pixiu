/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOmanage.goUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package filter

import (
	"github.com/apache/dubbo-go-pixiu/pkg/common/extension"
	"github.com/apache/dubbo-go-pixiu/pkg/common/yaml"
	"github.com/apache/dubbo-go-pixiu/pkg/model"
	"sync"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/apache/dubbo-go-pixiu/pkg/logger"
)


type filterManager struct {
	filters []extension.HttpFilter

	mu sync.RWMutex
}

func NewFilterManager() *filterManager {
	return &filterManager{filters: make([]extension.HttpFilter, 0, 16)}
}

func (fm *filterManager) GetFilters() []extension.HttpFilter {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	return fm.filters
}

// Load init or reload filter configs
func (fm *filterManager) Load(filters []*model.Filter) {
	tmp := make([]extension.HttpFilter, 0, len(filters))
	for _, f := range filters {
		apply, err := fm.Apply(f.Name, f.Config)
		if err != nil {
			logger.Errorf("apply [%s] init fail, %s", err)
		}
		tmp = append(tmp, apply)
	}
	// avoid filter inconsistency
	fm.mu.Lock()
	defer fm.mu.Unlock()

	fm.filters = tmp
}

// Apply return a new filter by name & conf
func (fm *filterManager) Apply(name string, conf map[string]interface{}) (extension.HttpFilter, error) {
	plugin, err := extension.GetHttpFilterPlugin(name)
	if err != nil {
		return nil, errors.New("filter not found")
	}

	filter, err := plugin.CreateFilter()

	if err != nil {
		return nil, errors.New("plugin create filter error")
	}

	factoryConf := filter.Config()
	if err := yaml.ParseConfig(factoryConf, conf); err != nil {
		return nil, errors.Wrap(err, "config error")
	}
	err = filter.Apply()
	if err != nil {
		return nil, errors.Wrap(err, "create fail")
	}
	return filter, nil
}