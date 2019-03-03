// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package renderer

import (
	"sort"

	"github.com/thommil/tge-g3n/camera"
	"github.com/thommil/tge-g3n/core"
	"github.com/thommil/tge-g3n/gls"
	"github.com/thommil/tge-g3n/graphic"
	"github.com/thommil/tge-g3n/light"
	"github.com/thommil/tge-g3n/math32"
)

// Renderer renders a 3D scene and/or a 2D GUI on the current window.
type Renderer struct {
	gs           *gls.GLS
	shaman       Shaman                     // Internal shader manager
	stats        Stats                      // Renderer statistics
	prevStats    Stats                      // Renderer statistics for previous frame
	scene        core.INode                 // Node containing 3D scene to render
	ambLights    []*light.Ambient           // Array of ambient lights for last scene
	dirLights    []*light.Directional       // Array of directional lights for last scene
	pointLights  []*light.Point             // Array of point
	spotLights   []*light.Spot              // Array of spot lights for the scene
	others       []core.INode               // Other nodes (audio, players, etc)
	rgraphics    []*graphic.Graphic         // Array of rendered graphics
	cgraphics    []*graphic.Graphic         // Array of rendered graphics
	grmatsOpaque []*graphic.GraphicMaterial // Array of rendered opaque graphic materials for scene
	grmatsTransp []*graphic.GraphicMaterial // Array of rendered transparent graphic materials for scene
	rinfo        core.RenderInfo            // Preallocated Render info
	specs        ShaderSpecs                // Preallocated Shader specs
	sortObjects  bool                       // Flag indicating whether objects should be sorted before rendering
	rendered     bool                       // Flag indicating if anything was rendered
	frameBuffers int                        // Number of frame buffers
	frameCount   int                        // Current number of frame buffers to write
}

// Stats describes how many object types were rendered.
// It is cleared at the start of each render.
type Stats struct {
	Graphics int // Number of graphic objects rendered
	Lights   int // Number of lights rendered
	Panels   int // Number of Gui panels rendered
	Others   int // Number of other objects rendered
}

// NewRenderer creates and returns a pointer to a new Renderer.
func NewRenderer(gs *gls.GLS) *Renderer {

	r := new(Renderer)
	r.gs = gs
	r.shaman.Init(gs)

	r.ambLights = make([]*light.Ambient, 0)
	r.dirLights = make([]*light.Directional, 0)
	r.pointLights = make([]*light.Point, 0)
	r.spotLights = make([]*light.Spot, 0)
	r.others = make([]core.INode, 0)
	r.rgraphics = make([]*graphic.Graphic, 0)
	r.cgraphics = make([]*graphic.Graphic, 0)
	r.grmatsOpaque = make([]*graphic.GraphicMaterial, 0)
	r.grmatsTransp = make([]*graphic.GraphicMaterial, 0)
	r.frameBuffers = 2
	r.sortObjects = true
	return r
}

// AddDefaultShaders adds to this renderer's shader manager all default
// include chunks, shaders and programs statically registered.
func (r *Renderer) AddDefaultShaders() error {

	return r.shaman.AddDefaultShaders()
}

// AddChunk adds a shader chunk with the specified name and source code.
func (r *Renderer) AddChunk(name, source string) {

	r.shaman.AddChunk(name, source)
}

// AddShader adds a shader program with the specified name and source code.
func (r *Renderer) AddShader(name, source string) {

	r.shaman.AddShader(name, source)
}

// AddProgram adds the program with the specified name,
// with associated vertex and fragment shaders (previously registered).
func (r *Renderer) AddProgram(name, vertex, frag string, others ...string) {

	r.shaman.AddProgram(name, vertex, frag, others...)
}

// SetScene sets the 3D scene to be rendered.
// If set to nil, no 3D scene will be rendered.
func (r *Renderer) SetScene(scene core.INode) {

	r.scene = scene
}

// Stats returns a copy of the statistics for the last frame.
// Should be called after the frame was rendered.
func (r *Renderer) Stats() Stats {

	return r.stats
}

// SetObjectSorting sets whether objects will be sorted before rendering.
func (r *Renderer) SetObjectSorting(sort bool) {

	r.sortObjects = sort
}

// ObjectSorting returns whether objects will be sorted before rendering.
func (r *Renderer) ObjectSorting() bool {

	return r.sortObjects
}

// Render renders the previously set Scene and Gui using the specified camera.
// Returns an indication if anything was rendered and an error.
func (r *Renderer) Render(icam camera.ICamera) (bool, error) {

	r.rendered = false
	r.stats = Stats{}

	// Renders the 3D scene
	if r.scene != nil {
		err := r.renderScene(r.scene, icam)
		if err != nil {
			return r.rendered, err
		}
	}

	r.prevStats = r.stats
	return r.rendered, nil
}

// renderScene renders the 3D scene using the specified camera.
func (r *Renderer) renderScene(iscene core.INode, icam camera.ICamera) error {

	// Updates world matrices of all scene nodes
	iscene.UpdateMatrixWorld()
	scene := iscene.GetNode()

	// Builds RenderInfo calls RenderSetup for all visible nodes
	icam.ViewMatrix(&r.rinfo.ViewMatrix)
	icam.ProjMatrix(&r.rinfo.ProjMatrix)

	// Clear scene arrays
	r.ambLights = r.ambLights[0:0]
	r.dirLights = r.dirLights[0:0]
	r.pointLights = r.pointLights[0:0]
	r.spotLights = r.spotLights[0:0]
	r.others = r.others[0:0]
	r.rgraphics = r.rgraphics[0:0]
	r.cgraphics = r.cgraphics[0:0]
	r.grmatsOpaque = r.grmatsOpaque[0:0]
	r.grmatsTransp = r.grmatsTransp[0:0]

	// Prepare for frustum culling
	var proj math32.Matrix4
	proj.MultiplyMatrices(&r.rinfo.ProjMatrix, &r.rinfo.ViewMatrix)
	frustum := math32.NewFrustumFromMatrix(&proj)

	// Internal function to classify a node and its children
	var classifyNode func(inode core.INode)
	classifyNode = func(inode core.INode) {

		// If node not visible, ignore
		node := inode.GetNode()
		if !node.Visible() {
			return
		}

		// Checks if node is a Graphic
		igr, ok := inode.(graphic.IGraphic)
		if ok {
			if igr.Renderable() {

				gr := igr.GetGraphic()

				// Frustum culling
				if igr.Cullable() {
					mw := gr.MatrixWorld()
					geom := igr.GetGeometry()
					bb := geom.BoundingBox()
					bb.ApplyMatrix4(&mw)
					if frustum.IntersectsBox(&bb) {
						// Append graphic to list of graphics to be rendered
						r.rgraphics = append(r.rgraphics, gr)
					} else {
						// Append graphic to list of culled graphics
						r.cgraphics = append(r.cgraphics, gr)
					}
				} else {
					// Append graphic to list of graphics to be rendered
					r.rgraphics = append(r.rgraphics, gr)
				}
			}
			// Node is not a Graphic
		} else {
			// Checks if node is a Light
			il, ok := inode.(light.ILight)
			if ok {
				switch l := il.(type) {
				case *light.Ambient:
					r.ambLights = append(r.ambLights, l)
				case *light.Directional:
					r.dirLights = append(r.dirLights, l)
				case *light.Point:
					r.pointLights = append(r.pointLights, l)
				case *light.Spot:
					r.spotLights = append(r.spotLights, l)
				default:
					panic("Invalid light type")
				}
				// Other nodes
			} else {
				r.others = append(r.others, inode)
			}
		}

		// Classify node children
		for _, ichild := range node.Children() {
			classifyNode(ichild)
		}
	}

	// Classify all scene nodes
	classifyNode(scene)

	//log.Debug("Rendered/Culled: %v/%v", len(r.grmats), len(r.cgrmats))

	// Sets lights count in shader specs
	r.specs.AmbientLightsMax = len(r.ambLights)
	r.specs.DirLightsMax = len(r.dirLights)
	r.specs.PointLightsMax = len(r.pointLights)
	r.specs.SpotLightsMax = len(r.spotLights)

	// Pre-calculate MV and MVP matrices and compile lists of opaque and transparent graphic materials
	for _, gr := range r.rgraphics {
		// Calculate MV and MVP matrices for all graphics to be rendered
		gr.CalculateMatrices(r.gs, &r.rinfo)

		// Append all graphic materials of this graphic to list of graphic materials to be rendered
		materials := gr.Materials()
		for i := 0; i < len(materials); i++ {
			if materials[i].IMaterial().GetMaterial().Transparent() {
				r.grmatsTransp = append(r.grmatsTransp, &materials[i])
			} else {
				r.grmatsOpaque = append(r.grmatsOpaque, &materials[i])
			}
		}
	}

	// TODO: If both GraphicMaterials belong to same Graphic we might want to keep their relative order...

	// Z-sort graphic materials (opaque front-to-back and transparent back-to-front)
	if r.sortObjects {
		// Internal function to render a list of graphic materials
		var zSortGraphicMaterials func(grmats []*graphic.GraphicMaterial, backToFront bool)
		zSortGraphicMaterials = func(grmats []*graphic.GraphicMaterial, backToFront bool) {
			sort.Slice(grmats, func(i, j int) bool {
				gr1 := grmats[i].IGraphic().GetGraphic()
				gr2 := grmats[j].IGraphic().GetGraphic()

				// Check for user-supplied render order
				rO1 := gr1.RenderOrder()
				rO2 := gr2.RenderOrder()
				if rO1 != rO2 {
					return rO1 < rO2
				}

				mvm1 := gr1.ModelViewMatrix()
				mvm2 := gr2.ModelViewMatrix()
				g1pos := gr1.Position()
				g2pos := gr2.Position()
				g1pos.ApplyMatrix4(mvm1)
				g2pos.ApplyMatrix4(mvm2)

				if backToFront {
					return g1pos.Z < g2pos.Z
				}

				return g1pos.Z > g2pos.Z
			})
		}

		zSortGraphicMaterials(r.grmatsOpaque, false) // Sort opaque graphics front to back
		zSortGraphicMaterials(r.grmatsTransp, true)  // Sort transparent graphics back to front
	}

	// Render other nodes (audio players, etc)
	for i := 0; i < len(r.others); i++ {
		inode := r.others[i]
		if !inode.GetNode().Visible() {
			continue
		}
		r.others[i].Render(r.gs)
		r.stats.Others++
	}

	// If there is graphic material to render or there was in the previous frame
	// it is necessary to clear the screen.
	if len(r.grmatsOpaque) > 0 || len(r.grmatsTransp) > 0 || r.prevStats.Graphics > 0 {
		// Clears the area inside the current scissor
		r.gs.Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		r.rendered = true
	}

	err := error(nil)

	// Internal function to render a list of graphic materials
	var renderGraphicMaterials func(grmats []*graphic.GraphicMaterial)
	renderGraphicMaterials = func(grmats []*graphic.GraphicMaterial) {
		// For each *GraphicMaterial
		for _, grmat := range grmats {
			mat := grmat.IMaterial().GetMaterial()
			geom := grmat.IGraphic().GetGeometry()
			gr := grmat.IGraphic().GetGraphic()

			// Add defines from material and geometry
			r.specs.Defines = *gls.NewShaderDefines()
			r.specs.Defines.Add(&mat.ShaderDefines)
			r.specs.Defines.Add(&geom.ShaderDefines)
			r.specs.Defines.Add(&gr.ShaderDefines)

			// Sets the shader specs for this material and sets shader program
			r.specs.Name = mat.Shader()
			r.specs.ShaderUnique = mat.ShaderUnique()
			r.specs.UseLights = mat.UseLights()
			r.specs.MatTexturesMax = mat.TextureCount()

			// Set active program and apply shader specs
			_, err = r.shaman.SetProgram(&r.specs)
			if err != nil {
				return
			}

			// Setup lights (transfer lights' uniforms)
			for idx, l := range r.ambLights {
				l.RenderSetup(r.gs, &r.rinfo, idx)
				r.stats.Lights++
			}
			for idx, l := range r.dirLights {
				l.RenderSetup(r.gs, &r.rinfo, idx)
				r.stats.Lights++
			}
			for idx, l := range r.pointLights {
				l.RenderSetup(r.gs, &r.rinfo, idx)
				r.stats.Lights++
			}
			for idx, l := range r.spotLights {
				l.RenderSetup(r.gs, &r.rinfo, idx)
				r.stats.Lights++
			}

			// Render this graphic material
			grmat.Render(r.gs, &r.rinfo)
			r.stats.Graphics++
		}
	}

	renderGraphicMaterials(r.grmatsOpaque) // Render opaque objects (front to back)
	if err != nil {
		return err
	}
	renderGraphicMaterials(r.grmatsTransp) // Render transparent objects (back to front)

	return err
}
