package dimensions

import (
	"fmt"
)

// Step is a Dimension.
type Step struct {
	Number int
}

// Table implements the Dimension interface.
func (d Step) Table() []string {
	return []string{fmt.Sprintf("step%d", d.Number)}
}

// Column implements the Dimension interface.
func (d Step) Column(_ string) string {
	return ""
}

// Expression implements the Dimension interface.
func (d Step) Expression() string {
	return ""
}

// Args implements the Dimension interface.
func (d Step) Args() []any {
	return nil
}

// ScanType implements the Metric interface.
func (d Step) ScanType() any {
	return nil
}
