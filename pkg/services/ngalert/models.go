package ngalert

import (
	"fmt"
	"time"

	"github.com/grafana/grafana/pkg/services/ngalert/eval"
)

// AlertDefinition is the model for alert definitions in Alerting NG.
type AlertDefinition struct {
	Id        int64
	OrgId     int64
	Name      string
	Condition string
	Data      []alertQuery
}

var (
	// ErrAlertDefinitionNotFound is an error for an unknown alert definition.
	ErrAlertDefinitionNotFound = fmt.Errorf("could not find alert definition")
)

// condition is the structure used by storing/updating alert definition commmands
type condition struct {
	RefID string `json:"refId"`

	QueriesAndExpressions []alertQuery `json:"queriesAndExpressions"`
}

// SaveAlertDefinitionCommand contains parameters for saving a new alert definition.
type SaveAlertDefinitionCommand struct {
	Name      string    `json:"name"`
	Condition condition `json:"condition"`
}

// IsValid validates a SaveAlertDefinitionCommand.
// Always returns true.
func (cmd *SaveAlertDefinitionCommand) IsValid() bool {
	return true
}

// updateAlertDefinitionCommand contains parameters for updating an existing alert definition.
type updateAlertDefinitionCommand struct {
	Name      string    `json:"name"`
	Condition condition `json:"condition"`
}

// IsValid validates an updateAlertDefinitionCommand.
// Always returns true.
func (cmd *updateAlertDefinitionCommand) IsValid() bool {
	return true
}

type evalAlertConditionCommand struct {
	Condition eval.Condition `json:"condition"`
	Now       time.Time      `json:"now"`
}
