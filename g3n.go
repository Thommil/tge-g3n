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
}

var _pluginInstance = &plugin{}

func (p *plugin) Init(runtime tge.Runtime) error {
	runtime.Use(gl.GetPlugin())
	return nil
}

func (p *plugin) GetName() string {
	return Name
}

func (p *plugin) Dispose() {

}

// GetPlugin returns plugin handler
func GetPlugin() tge.Plugin {
	return _pluginInstance
}
