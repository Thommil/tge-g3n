// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gls

import (
	"math"
	"unsafe"

	"github.com/thommil/tge-g3n/math32"
	gl "github.com/thommil/tge-gl"
)

// GLS encapsulates the state of an OpenGL context and contains
// methods to call OpenGL functions.
type GLS struct {
	stats               Stats             // statistics
	prog                *Program          // current active shader program
	programs            map[*Program]bool // shader programs cache
	checkErrors         bool              // check openGL API errors flag
	activeTexture       uint32            // cached last set active texture unit
	viewportX           int32             // cached last set viewport x
	viewportY           int32             // cached last set viewport y
	viewportWidth       int32             // cached last set viewport width
	viewportHeight      int32             // cached last set viewport height
	lineWidth           float32           // cached last set line width
	sideView            int               // cached last set triangle side view mode
	frontFace           uint32            // cached last set glFrontFace value
	depthFunc           uint32            // cached last set depth function
	depthMask           int               // cached last set depth mask
	capabilities        map[int]int       // cached capabilities (Enable/Disable)
	blendEquation       uint32            // cached last set blend equation value
	blendSrc            uint32            // cached last set blend src value
	blendDst            uint32            // cached last set blend equation destination value
	blendEquationRGB    uint32            // cached last set blend equation rgb value
	blendEquationAlpha  uint32            // cached last set blend equation alpha value
	blendSrcRGB         uint32            // cached last set blend src rgb
	blendSrcAlpha       uint32            // cached last set blend src alpha value
	blendDstRGB         uint32            // cached last set blend destination rgb value
	blendDstAlpha       uint32            // cached last set blend destination alpha value
	polygonModeFace     uint32            // cached last set polygon mode face
	polygonModeMode     uint32            // cached last set polygon mode mode
	polygonOffsetFactor float32           // cached last set polygon offset factor
	polygonOffsetUnits  float32           // cached last set polygon offset units
	// gobuf               []byte            // conversion buffer with GO memory
	// cbuf                []byte            // conversion buffer with C memory
}

// Stats contains counters of OpenGL resources being used as well
// the cumulative numbers of some OpenGL calls for performance evaluation.
type Stats struct {
	Shaders    int    // Current number of shader programs
	Vaos       int    // Number of Vertex Array Objects
	Buffers    int    // Number of Buffer Objects
	Textures   int    // Number of Textures
	Caphits    uint64 // Cumulative number of hits for Enable/Disable
	UnilocHits uint64 // Cumulative number of uniform location cache hits
	UnilocMiss uint64 // Cumulative number of uniform location cache misses
	Unisets    uint64 // Cumulative number of uniform sets
	Drawcalls  uint64 // Cumulative number of draw calls
}

// Polygon side view.
const (
	FrontSide = iota + 1
	BackSide
	DoubleSide
)

const (
	FloatSize = int32(unsafe.Sizeof(float32(0)))
)

const (
	capUndef    = 0
	capDisabled = 1
	capEnabled  = 2
	uintUndef   = math.MaxUint16
	intFalse    = 0
	intTrue     = 1
)

// New creates and returns a new instance of a GLS object,
// which encapsulates the state of an OpenGL context.
// This should be called only after an active OpenGL context
// is established, such as by creating a new window.
func New() (*GLS, error) {

	gs := new(GLS)
	gs.reset()
	gs.setDefaultState()
	gs.checkErrors = true
	return gs, nil
}

// SetCheckErrors enables/disables checking for errors after the
// call of any OpenGL function. It is enabled by default but
// could be disabled after an application is stable to improve the performance.
func (gs *GLS) SetCheckErrors(enable bool) {
	gs.checkErrors = enable
}

// CheckErrors returns if error checking is enabled or not.
func (gs *GLS) CheckErrors() bool {

	return gs.checkErrors
}

// reset resets the internal state kept of the OpenGL
func (gs *GLS) reset() {

	gs.lineWidth = 0.0
	gs.sideView = uintUndef
	gs.frontFace = 0
	gs.depthFunc = 0
	gs.depthMask = uintUndef
	gs.capabilities = make(map[int]int)
	gs.programs = make(map[*Program]bool)
	gs.prog = nil

	gs.activeTexture = uintUndef
	gs.blendEquation = uintUndef
	gs.blendSrc = uintUndef
	gs.blendDst = uintUndef
	gs.blendEquationRGB = 0
	gs.blendEquationAlpha = 0
	gs.blendSrcRGB = uintUndef
	gs.blendSrcAlpha = uintUndef
	gs.blendDstRGB = uintUndef
	gs.blendDstAlpha = uintUndef
	gs.polygonModeFace = 0
	gs.polygonModeMode = 0
	gs.polygonOffsetFactor = -1
	gs.polygonOffsetUnits = -1
}

// setDefaultState is used internally to set the initial state of OpenGL
// for this context.
func (gs *GLS) setDefaultState() {

	gs.Enable(DEPTH_TEST)
	gs.DepthFunc(LEQUAL)
	gs.FrontFace(CCW)
	gs.CullFace(BACK)
	gs.Enable(CULL_FACE)
	gs.Enable(BLEND)
	gs.BlendEquation(FUNC_ADD)
	gs.BlendFunc(SRC_ALPHA, ONE_MINUS_SRC_ALPHA)
	gs.Enable(VERTEX_PROGRAM_POINT_SIZE)
	gs.Enable(PROGRAM_POINT_SIZE)
	gs.Enable(MULTISAMPLE)
	gs.Enable(POLYGON_OFFSET_FILL)
	gs.Enable(POLYGON_OFFSET_LINE)
	gs.Enable(POLYGON_OFFSET_POINT)
}

// Stats copy the current values of the internal statistics structure
// to the specified pointer.
func (gs *GLS) Stats(s *Stats) {

	*s = gs.stats
	s.Shaders = len(gs.programs)
}

// ActiveTexture selects which texture unit subsequent texture state calls
// will affect. The number of texture units an implementation supports is
// implementation dependent, but must be at least 48 in GL 3.3.
func (gs *GLS) ActiveTexture(texture uint32) {

	if gs.activeTexture == texture {
		return
	}
	gl.ActiveTexture(gl.Enum(texture))
	gs.activeTexture = texture
}

// AttachShader attaches the specified shader object to the specified program object.
func (gs *GLS) AttachShader(program, shader uint32) {
	gl.AttachShader(gl.Program(program), gl.Shader(shader))
}

// BindBuffer binds a buffer object to the specified buffer binding point.
func (gs *GLS) BindBuffer(target int, vbo uint32) {
	gl.BindBuffer(gl.Enum(target), gl.Buffer(vbo))
}

// BindTexture lets you create or use a named texture.
func (gs *GLS) BindTexture(target int, tex uint32) {
	gl.BindTexture(gl.Enum(target), gl.Texture(tex))
}

// BindVertexArray binds the vertex array object.
func (gs *GLS) BindVertexArray(vao uint32) {
	gl.BindVertexArray(gl.VertexArray(vao))
}

// BlendEquation sets the blend equations for all draw buffers.
func (gs *GLS) BlendEquation(mode uint32) {

	if gs.blendEquation == mode {
		return
	}
	gl.BlendEquation(gl.Enum(mode))
	gs.blendEquation = mode
}

// BlendEquationSeparate sets the blend equations for all draw buffers
// allowing different equations for the RGB and alpha components.
func (gs *GLS) BlendEquationSeparate(modeRGB uint32, modeAlpha uint32) {
	if gs.blendEquationRGB == modeRGB && gs.blendEquationAlpha == modeAlpha {
		return
	}
	gl.BlendEquationSeparate(gl.Enum(modeRGB), gl.Enum(modeAlpha))
	gs.blendEquationRGB = modeRGB
	gs.blendEquationAlpha = modeAlpha
}

// BlendFunc defines the operation of blending for
// all draw buffers when blending is enabled.
func (gs *GLS) BlendFunc(sfactor, dfactor uint32) {

	if gs.blendSrc == sfactor && gs.blendDst == dfactor {
		return
	}
	gl.BlendFunc(gl.Enum(sfactor), gl.Enum(dfactor))
	gs.blendSrc = sfactor
	gs.blendDst = dfactor
}

// BlendFuncSeparate defines the operation of blending for all draw buffers when blending
// is enabled, allowing different operations for the RGB and alpha components.
func (gs *GLS) BlendFuncSeparate(srcRGB uint32, dstRGB uint32, srcAlpha uint32, dstAlpha uint32) {

	if gs.blendSrcRGB == srcRGB && gs.blendDstRGB == dstRGB &&
		gs.blendSrcAlpha == srcAlpha && gs.blendDstAlpha == dstAlpha {
		return
	}
	gl.BlendFuncSeparate(gl.Enum(srcRGB), gl.Enum(dstRGB), gl.Enum(srcAlpha), gl.Enum(dstAlpha))
	gs.blendSrcRGB = srcRGB
	gs.blendDstRGB = dstRGB
	gs.blendSrcAlpha = srcAlpha
	gs.blendDstAlpha = dstAlpha
}

// BufferData creates a new data store for the buffer object currently
// bound to target, deleting any pre-existing data store.
func (gs *GLS) BufferData(target uint32, size int, data interface{}, usage uint32) {
	switch data.(type) {
	case math32.ArrayU32:
		gl.BufferData(gl.Enum(target), gl.PointerToBytes(&(data.(math32.ArrayU32)[0]), size), gl.Enum(usage))
	case math32.ArrayF32:
		gl.BufferData(gl.Enum(target), gl.PointerToBytes(&(data.(math32.ArrayF32)[0]), size), gl.Enum(usage))
	default:
		gl.BufferData(gl.Enum(target), gl.PointerToBytes(data, size), gl.Enum(usage))
	}
}

// ClearColor specifies the red, green, blue, and alpha values
// used by glClear to clear the color buffers.
func (gs *GLS) ClearColor(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
}

// Clear sets the bitplane area of the window to values previously
// selected by ClearColor, ClearDepth, and ClearStencil.
func (gs *GLS) Clear(mask uint) {
	gl.Clear(gl.Enum(mask))
}

// CompileShader compiles the source code strings that
// have been stored in the specified shader object.
func (gs *GLS) CompileShader(shader uint32) {
	gl.CompileShader(gl.Shader(shader))
}

// CreateProgram creates an empty program object and returns
// a non-zero value by which it can be referenced.
func (gs *GLS) CreateProgram() uint32 {
	return uint32(gl.CreateProgram())
}

// CreateShader creates an empty shader object and returns
// a non-zero value by which it can be referenced.
func (gs *GLS) CreateShader(stype uint32) uint32 {
	return uint32(gl.CreateShader(gl.Enum(stype)))
}

// DeleteBuffers deletes n​buffer objects named
// by the elements of the provided array.
func (gs *GLS) DeleteBuffers(bufs ...uint32) {
	for _, buf := range bufs {
		gl.DeleteBuffer(gl.Buffer(buf))
		gs.stats.Buffers--
	}
}

// DeleteShader frees the memory and invalidates the name
// associated with the specified shader object.
func (gs *GLS) DeleteShader(shader uint32) {
	gl.DeleteShader(gl.Shader(shader))
}

// DeleteProgram frees the memory and invalidates the name
// associated with the specified program object.
func (gs *GLS) DeleteProgram(program uint32) {
	gl.DeleteProgram(gl.Program(program))
}

// DeleteTextures deletes n​textures named
// by the elements of the provided array.
func (gs *GLS) DeleteTextures(texs ...uint32) {
	for _, tex := range texs {
		gl.DeleteTexture(gl.Texture(tex))
		gs.stats.Textures--
	}
}

// DeleteVertexArrays deletes n​vertex array objects named
// by the elements of the provided array.
func (gs *GLS) DeleteVertexArrays(vaos ...uint32) {
	for _, vao := range vaos {
		gl.DeleteVertexArray(gl.VertexArray(vao))
		gs.stats.Vaos--
	}
}

// DepthFunc specifies the function used to compare each incoming pixel
// depth value with the depth value present in the depth buffer.
func (gs *GLS) DepthFunc(mode uint32) {

	if gs.depthFunc == mode {
		return
	}
	gl.DepthFunc(gl.Enum(mode))
	gs.depthFunc = mode
}

// DepthMask enables or disables writing into the depth buffer.
func (gs *GLS) DepthMask(flag bool) {

	if gs.depthMask == intTrue && flag {
		return
	}
	if gs.depthMask == intFalse && !flag {
		return
	}
	gl.DepthMask(flag)
	if flag {
		gs.depthMask = intTrue
	} else {
		gs.depthMask = intFalse
	}
}

// DrawArrays renders primitives from array data.
func (gs *GLS) DrawArrays(mode uint32, first int32, count int32) {
	gl.DrawArrays(gl.Enum(mode), int(first), int(count))
	gs.stats.Drawcalls++
}

// DrawElements renders primitives from array data.
func (gs *GLS) DrawElements(mode uint32, count int32, itype uint32, start uint32) {
	gl.DrawElements(gl.Enum(mode), int(count), gl.Enum(itype), int(start))
	gs.stats.Drawcalls++
}

// Enable enables the specified capability.
func (gs *GLS) Enable(cap int) {

	if gs.capabilities[cap] == capEnabled {
		gs.stats.Caphits++
		return
	}
	gl.Enable(gl.Enum(cap))
	gs.capabilities[cap] = capEnabled
}

// Disable disables the specified capability.
func (gs *GLS) Disable(cap int) {

	if gs.capabilities[cap] == capDisabled {
		gs.stats.Caphits++
		return
	}
	gl.Disable(gl.Enum(cap))
	gs.capabilities[cap] = capDisabled
}

// EnableVertexAttribArray enables a generic vertex attribute array.
func (gs *GLS) EnableVertexAttribArray(index uint32) {
	gl.EnableVertexAttribArray(gl.Attrib(int32(index)))
}

// CullFace specifies whether front- or back-facing facets can be culled.
func (gs *GLS) CullFace(mode uint32) {
	gl.CullFace(gl.Enum(mode))
}

// FrontFace defines front- and back-facing polygons.
func (gs *GLS) FrontFace(mode uint32) {

	if gs.frontFace == mode {
		return
	}
	gl.FrontFace(gl.Enum(mode))
	gs.frontFace = mode
}

// GenBuffer generates a​buffer object name.
func (gs *GLS) GenBuffer() uint32 {
	buf := gl.CreateBuffer()
	gs.stats.Buffers++
	return uint32(buf)
}

// GenerateMipmap generates mipmaps for the specified texture target.
func (gs *GLS) GenerateMipmap(target uint32) {
	gl.GenerateMipmap(gl.Enum(target))
}

// GenTexture generates a texture object name.
func (gs *GLS) GenTexture() uint32 {
	tex := gl.CreateTexture()
	gs.stats.Textures++
	return uint32(tex)
}

// GenVertexArray generates a vertex array object name.
func (gs *GLS) GenVertexArray() uint32 {
	vao := gl.CreateVertexArray()
	gs.stats.Vaos++
	return uint32(vao)
}

// GetAttribLocation returns the location of the specified attribute variable.
func (gs *GLS) GetAttribLocation(program uint32, name string) int32 {
	loc := gl.GetAttribLocation(gl.Program(program), name)
	return int32(loc)
}

// GetProgramiv returns the specified parameter from the specified program object.
func (gs *GLS) GetProgramiv(program, pname uint32, params *int32) {
	*params = int32(gl.GetProgrami(gl.Program(program), gl.Enum(pname)))
}

// GetProgramInfoLog returns the information log for the specified program object.
func (gs *GLS) GetProgramInfoLog(program uint32) string {
	return gl.GetProgramInfoLog(gl.Program(program))
}

// GetShaderInfoLog returns the information log for the specified shader object.
func (gs *GLS) GetShaderInfoLog(shader uint32) string {
	return gl.GetShaderInfoLog(gl.Shader(shader))
}

// GetString returns a string describing the specified aspect of the current GL connection.
func (gs *GLS) GetString(name uint32) string {
	return gl.GetString(gl.Enum(name))
}

// GetUniformLocation returns the location of a uniform variable for the specified program.
func (gs *GLS) GetUniformLocation(program uint32, name string) int32 {
	loc := gl.GetUniformLocation(gl.Program(program), name)
	return int32(loc)
}

// GetViewport returns the current viewport information.
func (gs *GLS) GetViewport() (x, y, width, height int32) {
	return gs.viewportX, gs.viewportY, gs.viewportWidth, gs.viewportHeight
}

// LineWidth specifies the rasterized width of both aliased and antialiased lines.
func (gs *GLS) LineWidth(width float32) {
	if gs.lineWidth == width {
		return
	}
	gl.LineWidth(width)
	gs.lineWidth = width
}

// LinkProgram links the specified program object.
func (gs *GLS) LinkProgram(program uint32) {
	gl.LinkProgram(gl.Program(program))
}

// GetShaderiv returns the specified parameter from the specified shader object.
func (gs *GLS) GetShaderiv(shader, pname uint32, params *int32) {
	*params = int32(gl.GetShaderi(gl.Shader(shader), gl.Enum(pname)))
}

// Scissor defines the scissor box rectangle in window coordinates.
func (gs *GLS) Scissor(x, y int32, width, height uint32) {
	gl.Scissor(x, y, int32(width), int32(height))
}

// ShaderSource sets the source code for the specified shader object.
func (gs *GLS) ShaderSource(shader uint32, src string) {
	gl.ShaderSource(gl.Shader(shader), src)
}

// TexImage2D specifies a two-dimensional texture image.
func (gs *GLS) TexImage2D(target uint32, level int32, iformat int32, width int32, height int32, border int32, format uint32, itype uint32, data interface{}) {
	gl.TexImage2D(gl.Enum(target), int(level), int(width), int(height), gl.Enum(format), gl.Enum(itype), data.([]byte))
}

// TexParameteri sets the specified texture parameter on the specified texture.
func (gs *GLS) TexParameteri(target uint32, pname uint32, param int32) {
	gl.TexParameteri(gl.Enum(target), gl.Enum(target), int(param))
}

// PolygonMode controls the interpretation of polygons for rasterization.
func (gs *GLS) PolygonMode(face, mode uint32) {

	if gs.polygonModeFace == face && gs.polygonModeMode == mode {
		return
	}
	gl.PolygonMode(gl.Enum(face), gl.Enum(mode))
	gs.polygonModeFace = face
	gs.polygonModeMode = mode
}

// PolygonOffset sets the scale and units used to calculate depth values.
func (gs *GLS) PolygonOffset(factor float32, units float32) {

	if gs.polygonOffsetFactor == factor && gs.polygonOffsetUnits == units {
		return
	}
	gl.PolygonOffset(factor, units)
	gs.polygonOffsetFactor = factor
	gs.polygonOffsetUnits = units
}

// Uniform1i sets the value of an int uniform variable for the current program object.
func (gs *GLS) Uniform1i(location int32, v0 int32) {
	gl.Uniform1i(gl.Uniform(location), int(v0))
	gs.stats.Unisets++
}

// Uniform1f sets the value of a float uniform variable for the current program object.
func (gs *GLS) Uniform1f(location int32, v0 float32) {
	gl.Uniform1f(gl.Uniform(location), v0)
	gs.stats.Unisets++
}

// Uniform2f sets the value of a vec2 uniform variable for the current program object.
func (gs *GLS) Uniform2f(location int32, v0, v1 float32) {
	gl.Uniform2f(gl.Uniform(location), v0, v1)
	gs.stats.Unisets++
}

// Uniform3f sets the value of a vec3 uniform variable for the current program object.
func (gs *GLS) Uniform3f(location int32, v0, v1, v2 float32) {
	gl.Uniform3f(gl.Uniform(location), v0, v1, v2)
	gs.stats.Unisets++
}

// Uniform4f sets the value of a vec4 uniform variable for the current program object.
func (gs *GLS) Uniform4f(location int32, v0, v1, v2, v3 float32) {
	gl.Uniform4f(gl.Uniform(location), v0, v1, v2, v3)
	gs.stats.Unisets++
}

// UniformMatrix3fv sets the value of one or many 3x3 float matrices for the current program object.
func (gs *GLS) UniformMatrix3fv(location int32, count int32, transpose bool, pm *float32) {
	gl.UniformMatrix3fvP(gl.Uniform(location), count, transpose, pm)
	gs.stats.Unisets++
}

// UniformMatrix4fv sets the value of one or many 4x4 float matrices for the current program object.
func (gs *GLS) UniformMatrix4fv(location int32, count int32, transpose bool, pm *float32) {
	gl.UniformMatrix4fvP(gl.Uniform(location), count, transpose, pm)
	gs.stats.Unisets++
}

// Uniform1fv sets the value of one or many float uniform variables for the current program object.
func (gs *GLS) Uniform1fv(location int32, count int32, v []float32) {
	gl.Uniform1fv(gl.Uniform(location), v)
	gs.stats.Unisets++
}

// Uniform2fv sets the value of one or many vec2 uniform variables for the current program object.
func (gs *GLS) Uniform2fv(location int32, count int32, v *float32) {
	gl.Uniform2fvP(gl.Uniform(location), count, v)
	gs.stats.Unisets++
}

func (gs *GLS) Uniform2fvUP(location int32, count int32, v unsafe.Pointer) {
	gl.Uniform2fvUP(gl.Uniform(location), count, v)
	gs.stats.Unisets++
}

// Uniform3fv sets the value of one or many vec3 uniform variables for the current program object.
func (gs *GLS) Uniform3fv(location int32, count int32, v *float32) {
	gl.Uniform3fvP(gl.Uniform(location), count, v)
	gs.stats.Unisets++
}

func (gs *GLS) Uniform3fvUP(location int32, count int32, v unsafe.Pointer) {
	gl.Uniform3fvUP(gl.Uniform(location), count, v)
	gs.stats.Unisets++
}

// Uniform4fv sets the value of one or many vec4 uniform variables for the current program object.
func (gs *GLS) Uniform4fv(location int32, count int32, v []float32) {
	gl.Uniform4fv(gl.Uniform(location), v)
	gs.stats.Unisets++
}

func (gs *GLS) Uniform4fvUP(location int32, count int32, v unsafe.Pointer) {
	gl.Uniform4fvUP(gl.Uniform(location), count, v)
	gs.stats.Unisets++
}

// VertexAttribPointer defines an array of generic vertex attribute data.
func (gs *GLS) VertexAttribPointer(index uint32, size int32, xtype uint32, normalized bool, stride int32, offset uint32) {
	gl.VertexAttribPointer(gl.Attrib(int32(index)), int(size), gl.Enum(xtype), normalized, int(stride), int(offset))
}

// Viewport sets the viewport.
func (gs *GLS) Viewport(x, y, width, height int32) {
	gl.Viewport(int(x), int(y), int(width), int(height))
	gs.viewportX = x
	gs.viewportY = y
	gs.viewportWidth = width
	gs.viewportHeight = height
}

// UseProgram sets the specified program as the current program.
func (gs *GLS) UseProgram(prog *Program) {
	if prog.handle == 0 {
		panic("Invalid program")
	}
	gl.UseProgram(gl.Program(prog.handle))
	gs.prog = prog

	// Inserts program in cache if not already there.
	if !gs.programs[prog] {
		gs.programs[prog] = true
	}
}
