package obs

import (
	"fmt"

	"github.com/fzeitner/beecs_masterthesis/comp"
	"github.com/fzeitner/beecs_masterthesis/comp_etox"
	"github.com/mlange-42/ark/ecs"
)

// PatchNectar is a row observer for the nectar availability of all patches in L (liters).
type PatchNectar struct {
	patchMapper *ecs.Map1[comp.Resource]

	data    []float64
	patches []ecs.Entity
	header  []string
}

func (o *PatchNectar) Initialize(w *ecs.World) {
	o.patchMapper = o.patchMapper.New(w)

	patchFilter := *ecs.NewFilter1[comp.Resource](w)
	query := patchFilter.Query()
	for query.Next() {
		e := query.Entity()
		o.patches = append(o.patches, e)
	}

	for i := range o.patches {
		o.header = append(o.header, fmt.Sprintf("Nectar_%d", i))
	}

	o.data = make([]float64, len(o.patches))
}
func (o *PatchNectar) Update(w *ecs.World) {}
func (o *PatchNectar) Header() []string {
	return o.header
}
func (o *PatchNectar) Values(w *ecs.World) []float64 {
	for i, e := range o.patches {
		res := o.patchMapper.Get(e)
		o.data[i] = res.Nectar * 0.000_001
	}
	return o.data
}

// PatchPPPNectar is a row observer for the PPPNectar availability of all patches in .
type PatchPPPNectar struct {
	patchMapper      *ecs.Map1[comp.Resource]
	patchMapper_etox *ecs.Map1[comp_etox.Resource_etox]

	data    []float64
	patches []ecs.Entity
	header  []string
}

func (o *PatchPPPNectar) Initialize(w *ecs.World) {
	o.patchMapper = o.patchMapper.New(w)
	o.patchMapper_etox = o.patchMapper_etox.New(w)

	patchFilter := *ecs.NewFilter1[comp.Resource](w)
	query := patchFilter.Query()
	for query.Next() {
		e := query.Entity()
		o.patches = append(o.patches, e)
	}

	for i := range o.patches {
		o.header = append(o.header, fmt.Sprintf("PPPConcentrationNectar_%d", i))
	}

	o.data = make([]float64, len(o.patches))
}
func (o *PatchPPPNectar) Update(w *ecs.World) {}
func (o *PatchPPPNectar) Header() []string {
	return o.header
}
func (o *PatchPPPNectar) Values(w *ecs.World) []float64 {
	for i, e := range o.patches {
		res := o.patchMapper_etox.Get(e)
		o.data[i] = res.PPPconcentrationNectar
	}
	return o.data
}

// PatchPPPPollen is a row observer for the PPPPollen availability of all patches in .
type PatchPPPPollen struct {
	patchMapper      *ecs.Map1[comp.Resource]
	patchMapper_etox *ecs.Map1[comp_etox.Resource_etox]

	data    []float64
	patches []ecs.Entity
	header  []string
}

func (o *PatchPPPPollen) Initialize(w *ecs.World) {
	o.patchMapper = o.patchMapper.New(w)
	o.patchMapper_etox = o.patchMapper_etox.New(w)

	patchFilter := *ecs.NewFilter1[comp.Resource](w)
	query := patchFilter.Query()
	for query.Next() {
		e := query.Entity()
		o.patches = append(o.patches, e)
	}

	for i := range o.patches {
		o.header = append(o.header, fmt.Sprintf("PPPConcentrationPollen_%d", i))
	}

	o.data = make([]float64, len(o.patches))
}
func (o *PatchPPPPollen) Update(w *ecs.World) {}
func (o *PatchPPPPollen) Header() []string {
	return o.header
}
func (o *PatchPPPPollen) Values(w *ecs.World) []float64 {
	for i, e := range o.patches {
		res := o.patchMapper_etox.Get(e)
		o.data[i] = res.PPPconcentrationPollen
	}
	return o.data
}

// PatchPPPContact is a row observer for the contact dose in all patches in .
type PatchPPPcontact struct {
	patchMapper      *ecs.Map1[comp.Resource]
	patchMapper_etox *ecs.Map1[comp_etox.Resource_etox]

	data    []float64
	patches []ecs.Entity
	header  []string
}

func (o *PatchPPPcontact) Initialize(w *ecs.World) {
	o.patchMapper = o.patchMapper.New(w)
	o.patchMapper_etox = o.patchMapper_etox.New(w)

	patchFilter := *ecs.NewFilter1[comp.Resource](w)
	query := patchFilter.Query()
	for query.Next() {
		e := query.Entity()
		o.patches = append(o.patches, e)
	}

	for i := range o.patches {
		o.header = append(o.header, fmt.Sprintf("PPPcontactPatch_%d", i))
	}

	o.data = make([]float64, len(o.patches))
}
func (o *PatchPPPcontact) Update(w *ecs.World) {}
func (o *PatchPPPcontact) Header() []string {
	return o.header
}
func (o *PatchPPPcontact) Values(w *ecs.World) []float64 {
	for i, e := range o.patches {
		res := o.patchMapper_etox.Get(e)
		o.data[i] = res.PPPcontactDose
	}
	return o.data
}

// PatchPollen is a row observer for the pollen availability of all patches, in g (grams).
type PatchPollen struct {
	patchMapper *ecs.Map1[comp.Resource]
	data        []float64
	patches     []ecs.Entity
	header      []string
}

func (o *PatchPollen) Initialize(w *ecs.World) {
	o.patchMapper = o.patchMapper.New(w)

	patchFilter := ecs.NewFilter1[comp.Resource](w)
	query := patchFilter.Query()
	for query.Next() {
		e := query.Entity()
		o.patches = append(o.patches, e)
	}

	for i := range o.patches {
		o.header = append(o.header, fmt.Sprintf("Pollen_%d", i))
	}

	o.data = make([]float64, len(o.patches))
}
func (o *PatchPollen) Update(w *ecs.World) {}
func (o *PatchPollen) Header() []string {
	return o.header
}
func (o *PatchPollen) Values(w *ecs.World) []float64 {
	for i, e := range o.patches {
		res := o.patchMapper.Get(e)
		o.data[i] = res.Pollen
	}
	return o.data
}
