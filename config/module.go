package config

// Module is the configuration shared by the application coordinator and the
// native renderer for one named panel component.
type Module struct {
	Name         string
	Kind         string
	Value        string
	Interval     int
	Icon         string
	TilingIcon   string
	FloatingIcon string
	WiredIcon    string
	IconColor    uint32
	HasIconColor bool
	FillColor    uint32
	HasFillColor bool
	Action       string
	Step         int
	Width        int
}
