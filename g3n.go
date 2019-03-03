// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package g3n

import (
	tge "github.com/thommil/tge"
	gl "github.com/thommil/tge-gl"
)

// Name name of the plugin
const Name = "g3n"

type plugin struct {
	runtime tge.Runtime
}

var _pluginInstance = &plugin{}

// GetInstance returns plugin handler
func GetInstance() tge.Plugin {
	return _pluginInstance
}

func (p *plugin) Init(runtime tge.Runtime) error {
	runtime.Use(gl.GetInstance())
	p.runtime = runtime
	return nil
}

func (p *plugin) GetName() string {
	return Name
}

func (p *plugin) Dispose() {
	p.runtime = nil
}

// LoadAsset gets assets from runtime instance
func LoadAsset(path string) ([]byte, error) {
	return _pluginInstance.runtime.GetAsset(path)
}
