<h1 align="center">TGE-G3N - G3N plugin for TGE</h1>

 <p align="center">
    <a href="https://godoc.org/github.com/thommil/tge-g3n"><img src="https://godoc.org/github.com/thommil/tge-g3n?status.svg" alt="Godoc"></img></a>
    <a href="https://goreportcard.com/report/github.com/thommil/tge-g3n"><img src="https://goreportcard.com/badge/github.com/thommil/tge-g3n"  alt="Go Report Card"/></a>
</p>

Game Engine for TGE runtime - [TGE](https://github.com/thommil/tge)

Based on :
 * [G3N](https://github.com/g3n/engine) - Copyright (c) 2016 The G3N Authors. All rights reserved.

## Targets
 * Desktop
 * Browsers
 * Mobile

## Dependencies
 * [TGE core](https://github.com/thommil/tge)
 * [TGE-GL](https://github.com/thommil/tge-gl)

## Limitations
### Not implemented:
 * audio (in study)
 * gui (planned)
 * window (replaced by TGE)

## Implementation
See example at [G3N examples](https://github.com/Thommil/tge-examples/tree/master/plugins/tge-g3n)

Just import package and replace G3N former entry point **Application** by **App** lifecycle:

```golang
package main

import (
	tge "github.com/thommil/tge"

	camera "github.com/thommil/tge-g3n/camera"
	core "github.com/thommil/tge-g3n/core"
	gls "github.com/thommil/tge-g3n/gls"
	light "github.com/thommil/tge-g3n/light"
	math32 "github.com/thommil/tge-g3n/math32"
	renderer "github.com/thommil/tge-g3n/renderer"
)

type G3NApp struct {
	runtime    tge.Runtime
	gls        *gls.GLS
	scene      *core.Node
	camPersp   *camera.Perspective
	renderer   *renderer.Renderer
}

func (app *G3NApp) OnStart(runtime tge.Runtime) error {
	runtime.Subscribe(tge.ResizeEvent{}.Channel(), app.OnResize)
	
	var err error

	// OpenGL
	app.gls, err = gls.New()
	if err != nil {
		return err
	}

    // INIT
	cc := math32.NewColor("black")
	app.gls.ClearColor(cc.R, cc.G, cc.B, 1)
	app.gls.Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)

    // SCENE
	app.scene = core.NewNode()
	app.renderer = renderer.NewRenderer(app.gls)
	err = app.renderer.AddDefaultShaders()
	if err != nil {
		return fmt.Errorf("Error from AddDefaulShaders:%v", err)
	}
	app.renderer.SetScene(app.scene)

	// TORUS
	geom := geometry.NewTorus(1, .4, 12, 32, math32.Pi*2)
	mat := material.NewPhong(math32.NewColor("DarkBlue"))
	torusMesh := graphic.NewMesh(geom, mat)
	app.scene.Add(torusMesh)
    ambientLight := light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8)
	app.scene.Add(ambientLight)

    // LIGHT
	pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 100.0)
	pointLight.SetPosition(0, 0, 25)
	app.scene.Add(pointLight)

    // CAMERA
	app.camPersp = camera.NewPerspective(65, 1, 0.01, 1000)
	app.camPersp.SetPosition(10, 10, 10)
	app.camPersp.LookAt(&math32.Vector3{0, 0, 0})

	return nil
}

func (app *G3NApp) OnResize(event tge.Event) bool {
	app.camPersp.SetAspect(float32(event.(tge.ResizeEvent).Width) / float32(event.(tge.ResizeEvent).Height))
	app.gls.Viewport(0, 0, event.(tge.ResizeEvent).Width, event.(tge.ResizeEvent).Height)
	return false
}

func (app *G3NApp) OnRender(elapsedTime time.Duration, syncChan <-chan interface{}) {
	<-syncChan
	app.gls.Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
	app.renderer.Render(app.camPersp)
}

func (app *G3NApp) OnTick(elapsedTime time.Duration, syncChan chan<- interface{}) {
    ...
    syncChan <- true
}

...

```

