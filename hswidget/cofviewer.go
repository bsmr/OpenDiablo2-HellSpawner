package hswidget

import (
	"fmt"
	"strconv"

	"github.com/ianling/giu"

	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2enum"
	"github.com/OpenDiablo2/OpenDiablo2/d2common/d2fileformats/d2cof"
)

const (
	indicatorSize = 64
)

// COFViewerState represents cof viewer's state
type COFViewerState struct {
	layerIndex     int32
	directionIndex int32
	frameIndex     int32
	layer          *d2cof.CofLayer
}

// Dispose clears viewer's layers
func (s *COFViewerState) Dispose() {
	s.layer = nil
}

// COFViewerWidget represents cof viewer's widget
type COFViewerWidget struct {
	id  string
	cof *d2cof.COF
}

// COFViewer creates a cof viewer widget
func COFViewer(id string, cof *d2cof.COF) *COFViewerWidget {
	result := &COFViewerWidget{
		id:  id,
		cof: cof,
	}

	return result
}

// Build builds a cof viewer
func (p *COFViewerWidget) Build() {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	s := giu.Context.GetState(stateID)

	if s == nil {
		giu.Context.SetState(stateID, &COFViewerState{
			layer: &p.cof.CofLayers[0],
		})

		return
	}

	state := s.(*COFViewerState)

	var l1, l2, l3, l4 string

	numDirs := p.cof.NumberOfDirections
	numFrames := p.cof.FramesPerDirection

	l1 = fmt.Sprintf("Directions: %v", numDirs)

	if numDirs > 1 {
		l2 = fmt.Sprintf("Frames (x%v): %v", numDirs, numFrames)
	} else {
		l2 = fmt.Sprintf("Frames: %v", numFrames)
	}

	// nolint:gomnd // constant
	fps := 25 * (float64(p.cof.Speed) / float64(256))
	if fps == 0 {
		fps = 25
	}

	l3 = fmt.Sprintf("FPS: %.1f", fps)
	// nolint:gomnd // miliseconds
	l4 = fmt.Sprintf("Duration: %.2fms", float64(numFrames)*(1/fps)*1000)

	layerStrings := make([]string, 0)
	for idx := range p.cof.CofLayers {
		layerStrings = append(layerStrings, strconv.Itoa(int(p.cof.CofLayers[idx].Type)))
	}

	layerList := giu.Combo("##"+p.id+"layer", layerStrings[state.layerIndex], layerStrings, &state.layerIndex).
		Size(indicatorSize).OnChange(p.onUpdate)

	directionStrings := make([]string, 0)
	for idx := range p.cof.Priority {
		directionStrings = append(directionStrings, fmt.Sprintf("%d", idx))
	}

	directionList := giu.Combo("##"+p.id+"dir", directionStrings[state.directionIndex], directionStrings, &state.directionIndex).
		Size(indicatorSize).OnChange(p.onUpdate)

	frameStrings := make([]string, 0)
	for idx := range p.cof.Priority[state.directionIndex] {
		frameStrings = append(frameStrings, fmt.Sprintf("%d", idx))
	}

	frameList := giu.Combo("##"+p.id+"frame", frameStrings[state.frameIndex], frameStrings, &state.frameIndex).
		Size(indicatorSize).OnChange(p.onUpdate)

	const vspace = 4 //nolint:unused // will be used

	giu.TabBar("COFViewerTabs").Layout(giu.Layout{
		giu.TabItem("Animation").Layout(giu.Layout{
			giu.Label(l1),
			giu.Label(l2),
			giu.Label(l3),
			giu.Label(l4),
		}),
		giu.TabItem("Layer").Layout(giu.Layout{
			giu.Layout{
				giu.Line(giu.Label("Selected Layer: "), layerList),
				giu.Separator(),
				p.makeLayerLayout(),
				giu.Button("Add layer...##"+p.id+"AddLayer").Size(200, 30),
				giu.Button("Delete current layer...##"+p.id+"DeleteLayer").Size(200, 30).OnClick(func() { p.deleteCurrentLayer(state.layerIndex) }),
			},
		}),
		giu.TabItem("Priority").Layout(giu.Layout{
			giu.Line(
				giu.Label("Direction: "), directionList,
				giu.Label("Frame: "), frameList,
			),
			giu.Separator(),
			p.makeDirectionLayout(),
		}),
	}).Build()
}

func (p *COFViewerWidget) deleteCurrentLayer(index int32) {
	p.cof.NumberOfLayers--

	newLayers := make([]d2cof.CofLayer, 0)
	for n, i := range p.cof.CofLayers {
		if int32(n) != index {
			newLayers = append(newLayers, i)
		}
	}

	p.cof.CofLayers = newLayers
}

func (p *COFViewerWidget) onUpdate() {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	state := giu.Context.GetState(stateID).(*COFViewerState)

	clone := p.cof.CofLayers[state.layerIndex]
	state.layer = &clone

	giu.Context.SetState(p.id, state)
}

func (p *COFViewerWidget) makeLayerLayout() giu.Layout {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	state := giu.Context.GetState(stateID).(*COFViewerState)

	if state.layer == nil {
		p.onUpdate()
	}

	layerName := getLayerName(state.layer.Type)

	strType := fmt.Sprintf("Type: %s (%s)", state.layer.Type, layerName)
	strShadow := fmt.Sprintf("Shadow: %t", state.layer.Shadow > 0)
	strSelectable := fmt.Sprintf("Selectable: %t", state.layer.Selectable)
	strTransparent := fmt.Sprintf("Transparent: %t", state.layer.Transparent)

	effect := p.getDrawEffect(state.layer.DrawEffect)

	strEffect := fmt.Sprintf("Draw Effect: %s", effect)

	weapon := p.getWeaponClass(state.layer.WeaponClass)

	strWeaponClass := fmt.Sprintf("Weapon Class: (%s) %s", state.layer.WeaponClass, weapon)

	return giu.Layout{
		giu.Label(strType),
		giu.Label(strShadow),
		giu.Label(strSelectable),
		giu.Label(strTransparent),
		giu.Label(strEffect),
		giu.Label(strWeaponClass),
	}
}

func (p *COFViewerWidget) getDrawEffect(eff d2enum.DrawEffect) string {
	var effect string

	switch eff {
	case d2enum.DrawEffectPctTransparency25:
		effect = "25% alpha"
	case d2enum.DrawEffectPctTransparency50:
		effect = "50% alpha"
	case d2enum.DrawEffectPctTransparency75:
		effect = "75% alpha"
	case d2enum.DrawEffectModulate:
		effect = "Modulate"
	case d2enum.DrawEffectBurn:
		effect = "Burn"
	case d2enum.DrawEffectNormal:
		effect = "Normal"
	case d2enum.DrawEffectMod2XTrans:
		effect = "Mod2XTrans"
	case d2enum.DrawEffectMod2X:
		effect = "Mod2X"
	case d2enum.DrawEffectNone:
		// nolint:goconst // that's not a constant
		effect = "None"
	}

	return effect
}

// nolint:gocyclo // can't reduce
func (p *COFViewerWidget) getWeaponClass(cls d2enum.WeaponClass) string {
	var weapon string

	switch cls {
	case d2enum.WeaponClassNone:
		// nolint:goconst // that's not a constant
		weapon = "None"
	case d2enum.WeaponClassHandToHand:
		weapon = "Hand To Hand"
	case d2enum.WeaponClassBow:
		weapon = "Bow"
	case d2enum.WeaponClassOneHandSwing:
		weapon = "One Hand Swing"
	case d2enum.WeaponClassOneHandThrust:
		weapon = "One Hand Thrust"
	case d2enum.WeaponClassStaff:
		weapon = "Staff"
	case d2enum.WeaponClassTwoHandSwing:
		weapon = "Two Hand Swing"
	case d2enum.WeaponClassTwoHandThrust:
		weapon = "Two Hand Thrust"
	case d2enum.WeaponClassCrossbow:
		weapon = "Crossbow"
	case d2enum.WeaponClassLeftJabRightSwing:
		weapon = "Left Jab Right Swing"
	case d2enum.WeaponClassLeftJabRightThrust:
		weapon = "Left Jab Right Thrust"
	case d2enum.WeaponClassLeftSwingRightSwing:
		weapon = "Left Swing Right Swing"
	case d2enum.WeaponClassLeftSwingRightThrust:
		weapon = "Left Swing Right Thrust"
	case d2enum.WeaponClassOneHandToHand:
		weapon = "One Hand To Hand"
	case d2enum.WeaponClassTwoHandToHand:
		weapon = "Two Hand To Hand"
	}

	return weapon
}

// nolint:gocyclo // can't reduce
func getLayerName(i interface{}) string {
	var t d2enum.CompositeType

	switch j := i.(type) {
	case int:
		t = d2enum.CompositeType(j)
	case d2enum.CompositeType:
		t = j
	}

	var layerName string

	switch t {
	case d2enum.CompositeTypeHead:
		layerName = "Head"
	case d2enum.CompositeTypeTorso:
		layerName = "Torso"
	case d2enum.CompositeTypeLegs:
		layerName = "Legs"
	case d2enum.CompositeTypeRightArm:
		layerName = "Right Arm"
	case d2enum.CompositeTypeLeftArm:
		layerName = "Left Arm"
	case d2enum.CompositeTypeRightHand:
		layerName = "Right Hand"
	case d2enum.CompositeTypeLeftHand:
		layerName = "Left Hand"
	case d2enum.CompositeTypeShield:
		layerName = "Shield"
	case d2enum.CompositeTypeSpecial1:
		layerName = "Special 1"
	case d2enum.CompositeTypeSpecial2:
		layerName = "Special 2"
	case d2enum.CompositeTypeSpecial3:
		layerName = "Special 3"
	case d2enum.CompositeTypeSpecial4:
		layerName = "Special 4"
	case d2enum.CompositeTypeSpecial5:
		layerName = "Special 5"
	case d2enum.CompositeTypeSpecial6:
		layerName = "Special 6"
	case d2enum.CompositeTypeSpecial7:
		layerName = "Special 7"
	case d2enum.CompositeTypeSpecial8:
		layerName = "Special 8"
	}

	return layerName
}

func (p *COFViewerWidget) makeDirectionLayout() giu.Layout {
	stateID := fmt.Sprintf("COFViewerWidget_%s", p.id)
	state := giu.Context.GetState(stateID).(*COFViewerState)

	frames := p.cof.Priority[state.directionIndex]
	layers := frames[int(state.frameIndex)%len(frames)]

	return giu.Layout{
		giu.Label("Render Order (first to last):"),
		giu.Custom(func() {
			for idx := range layers {
				giu.Label(fmt.Sprintf("\t%d: %s", idx, getLayerName(layers[idx]))).Build()
			}
		}),
	}
}
