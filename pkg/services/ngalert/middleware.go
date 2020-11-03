package ngalert

import (
	"github.com/grafana/grafana/pkg/models"
)

func (ng *AlertNG) validateOrgAlertDefinition(c *models.ReqContext) {
	id := c.ParamsInt64(":alertDefinitionId")
	alertDefinition, err := ng.getAlertDefinitionByID(id)
	if err != nil {
		// TODO: Distinguish between errors
		c.JsonApiErr(404, "Alert definition not found", nil)
		return
	}

	if c.OrgId != alertDefinition.OrgId {
		c.JsonApiErr(403, "You are not allowed to edit/view alert definition", nil)
		return
	}
}
