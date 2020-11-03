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
	Data      []AlertQuery
}

var (
	// ErrAlertDefinitionNotFound is an error for an unknown alert definition.
	ErrAlertDefinitionNotFound = fmt.Errorf("could not find alert definition")
)

// Condition is the structure used by storing/updating alert definition commmands
type Condition struct {
	RefID string `json:"refId"`

	QueriesAndExpressions []AlertQuery `json:"queriesAndExpressions"`
}

// SaveAlertDefinitionCommand contains parameters for saving a new alert definition.
type SaveAlertDefinitionCommand struct {
	Name      string    `json:"name"`
	Condition Condition `json:"condition"`
}

// IsValid validates a SaveAlertDefinitionCommand.
// Always returns true.
func (cmd *SaveAlertDefinitionCommand) IsValid() bool {
	return true
}

// UpdateAlertDefinitionCommand contains parameters for updating an existing alert definition.
type UpdateAlertDefinitionCommand struct {
	Name      string    `json:"name"`
	Condition Condition `json:"condition"`
}

// IsValid validates an UpdateAlertDefinitionCommand.
// Always returns true.
func (cmd *UpdateAlertDefinitionCommand) IsValid() bool {
	return true
}

type EvalAlertConditionCommand struct {
	Condition eval.Condition `json:"condition"`
	Now       time.Time      `json:"now"`
}
