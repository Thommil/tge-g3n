// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package g3n

import (
	tge "github.com/thommil/tge"
)

// Name name of the plugin
const Name = "g3n"

type plugin struct {
	runtime tge.Runtime
}

var _pluginInstance = &plugin{}

func init() {
	tge.Register(_pluginInstance)
}

func (p *plugin) Init(runtime tge.Runtime) error {
	p.runtime = runtime
	return nil
}

func (p *plugin) GetName() string {
	return Name
}

func (p *plugin) Dispose() {
	p.runtime = nil
}

// Runtime gives access to current running TGE Runtime
func Runtime() tge.Runtime {
	return _pluginInstance.runtime
}
