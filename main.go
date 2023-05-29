package main

import (
	"fmt"
	"math"
	"time"

	"github.com/g3n/engine/app"
	"github.com/g3n/engine/camera"
	"github.com/g3n/engine/core"
	"github.com/g3n/engine/geometry"
	"github.com/g3n/engine/gls"
	"github.com/g3n/engine/graphic"
	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/light"
	"github.com/g3n/engine/material"
	"github.com/g3n/engine/math32"
	"github.com/g3n/engine/renderer"
	"github.com/g3n/engine/util/helper"
	"github.com/g3n/engine/window"
)

func NewGlobe(radius float32, size int) *geometry.Geometry {

	t := geometry.NewGeometry()

	// Create buffers
	vertexCount := 12
	positions := math32.NewArrayF32(0, vertexCount*3)
	normals := math32.NewArrayF32(vertexCount*3, vertexCount*3)
	uvs := math32.NewArrayF32(vertexCount*2, vertexCount*2)
	indices := math32.NewArrayU32(0, vertexCount*3)

	// the coordinates of a icosahedron (20 sided regular polyhedron)
	// the orientation is one vertex "north" and one "south"
	// there are 12 vertices and 30 edges (and 20 sides)
	sq5inv_64 := 1.0 / math.Sqrt(5)
	sq5inv := float32(sq5inv_64)
	sq5inv2 := float32(2.0 * sq5inv_64)
	min1sq5div2 := float32((1.0 - sq5inv_64) / 2.0)
	sqsq5div2 := float32(math.Sqrt((1.0 + sq5inv_64) / 2.0))
	negmin1sq5div2 := float32((-1.0 - sq5inv_64) / 2.0)
	negsqsq5div2 := float32(math.Sqrt((1.0 - sq5inv_64) / 2.0))
	positions.AppendVector3(&math32.Vector3{1.0, 0.0, 0.0})                           // 0
	positions.AppendVector3(&math32.Vector3{sq5inv, sq5inv2, 0.0})                    // 1
	positions.AppendVector3(&math32.Vector3{sq5inv, min1sq5div2, sqsq5div2})          // 2
	positions.AppendVector3(&math32.Vector3{sq5inv, negmin1sq5div2, negsqsq5div2})    // 3
	positions.AppendVector3(&math32.Vector3{sq5inv, negmin1sq5div2, -negsqsq5div2})   // 4
	positions.AppendVector3(&math32.Vector3{sq5inv, min1sq5div2, -sqsq5div2})         // 5
	positions.AppendVector3(&math32.Vector3{-sq5inv, -negmin1sq5div2, -negsqsq5div2}) // 6
	positions.AppendVector3(&math32.Vector3{-sq5inv, -min1sq5div2, -sqsq5div2})       // 7
	positions.AppendVector3(&math32.Vector3{-sq5inv, -sq5inv2, 0.0})                  // 8
	positions.AppendVector3(&math32.Vector3{-sq5inv, -min1sq5div2, sqsq5div2})        // 9
	positions.AppendVector3(&math32.Vector3{-sq5inv, -negmin1sq5div2, negsqsq5div2})  // 10
	positions.AppendVector3(&math32.Vector3{-1.0, 0.0, 0.0})                          // 11

	for i := 0; i < 12; i++ {
		fmt.Println(i, ": ", positions[i*3:i*3+3])
	}

	// define the triangles for those 20 sides
	indices.Append(0, 1, 2)
	indices.Append(0, 2, 3)
	indices.Append(0, 3, 4)
	indices.Append(0, 4, 5)
	indices.Append(0, 5, 1) // end top ring
	indices.Append(10, 2, 1)
	indices.Append(10, 9, 2)
	indices.Append(9, 3, 2)
	indices.Append(9, 8, 3)
	indices.Append(8, 4, 3)
	indices.Append(8, 7, 4)
	indices.Append(7, 5, 4)
	indices.Append(7, 6, 5)
	indices.Append(6, 1, 5)
	indices.Append(6, 10, 1)
	indices.Append(6, 7, 11) // start bottom ring
	indices.Append(7, 8, 11)
	indices.Append(8, 9, 11)
	indices.Append(9, 10, 11)
	indices.Append(10, 6, 11)

	// normals, always pointing away from (0,0,0)
	for i := 0; i < 12; i++ {
		fmt.Println("idx ", i)
		var normal math32.Vector3
		normal.Set(positions[i*3], positions[i*3+1], positions[i*3+2]).Normalize()
		normals.SetVector3(i*3, &normal)
	}

	t.SetIndices(indices)
	t.AddVBO(gls.NewVBO(positions).AddAttrib(gls.VertexPosition))
	t.AddVBO(gls.NewVBO(normals).AddAttrib(gls.VertexNormal))
	t.AddVBO(gls.NewVBO(uvs).AddAttrib(gls.VertexTexcoord))

	return t
}

func main() {

	// Create application and scene
	a := app.App()
	scene := core.NewNode()

	// Set the scene to be managed by the gui manager
	gui.Manager().Set(scene)

	// Create perspective camera
	cam := camera.New(1)
	cam.SetPosition(0, 0, 3)
	scene.Add(cam)

	// Set up orbit control for the camera
	camera.NewOrbitControl(cam)

	// Set up callback to update viewport and camera aspect ratio when the window is resized
	onResize := func(evname string, ev interface{}) {
		// Get framebuffer size and update viewport accordingly
		width, height := a.GetSize()
		a.Gls().Viewport(0, 0, int32(width), int32(height))
		// Update the camera's aspect ratio
		cam.SetAspect(float32(width) / float32(height))
	}
	a.Subscribe(window.OnWindowSize, onResize)
	onResize("", nil)

	// Create a blue globe, add it to the scene
	geom := NewGlobe(1, 1)
	mat := material.NewStandard(math32.NewColor("DarkBlue"))
	mesh := graphic.NewMesh(geom, mat)
	scene.Add(mesh)

	// Add a button to change color, for no particular reason
	isBlue := true
	btn := gui.NewButton("Make Red")
	btn.SetPosition(100, 40)
	btn.SetSize(40, 40)
	btn.Subscribe(gui.OnClick, func(name string, ev interface{}) {
		if isBlue {
			mat.SetColor(math32.NewColor("DarkRed"))
			isBlue = false
			btn.Label.SetText("Make Blue")
		} else {
			mat.SetColor(math32.NewColor("DarkBlue"))
			isBlue = true
			btn.Label.SetText("Make Red")
		}
	})
	scene.Add(btn)

	// Create and add lights to the scene
	scene.Add(light.NewAmbient(&math32.Color{1.0, 1.0, 1.0}, 0.8))
	numbers := []float32{2.0, 0.0, -2.0}
	for i := 0; i < len(numbers); i++ {
		for j := 0; j < len(numbers); j++ {
			for k := 0; k < len(numbers); k++ {
				pointLight := light.NewPoint(&math32.Color{1, 1, 1}, 5.0)
				pointLight.SetPosition(numbers[i], numbers[j], numbers[k])
				scene.Add(pointLight)
			}
		}
	}

	// Create and add an axis helper to the scene
	scene.Add(helper.NewAxes(2.0))

	// Set background color to gray
	a.Gls().ClearColor(0.5, 0.5, 0.5, 1.0)

	// Run the application
	a.Run(func(renderer *renderer.Renderer, deltaTime time.Duration) {
		a.Gls().Clear(gls.DEPTH_BUFFER_BIT | gls.STENCIL_BUFFER_BIT | gls.COLOR_BUFFER_BIT)
		renderer.Render(scene, cam)
	})
}
