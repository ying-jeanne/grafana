package ngalert

import (
	"context"

	"github.com/grafana/grafana/pkg/services/ngalert/eval"
	"github.com/grafana/grafana/pkg/services/sqlstore"

	"github.com/go-macaron/binding"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana/pkg/api"
	"github.com/grafana/grafana/pkg/api/routing"
	"github.com/grafana/grafana/pkg/middleware"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/grafana/grafana/pkg/tsdb"
	"github.com/grafana/grafana/pkg/util"
)

func (ng *AlertNG) registerAPIEndpoints() {
	ng.RouteRegister.Group("/api/alert-definitions", func(alertDefinitions routing.RouteRegister) {
		alertDefinitions.Get("", middleware.ReqSignedIn, api.Wrap(ng.listAlertDefinitions))
		alertDefinitions.Get("/eval/:alertDefinitionId", ng.validateOrgAlertDefinition, api.Wrap(ng.AlertDefinitionEval))
		alertDefinitions.Post("/eval", middleware.ReqSignedIn, binding.Bind(evalAlertConditionCommand{}), api.Wrap(ng.conditionEval))
		alertDefinitions.Get("/:alertDefinitionId", ng.validateOrgAlertDefinition, api.Wrap(ng.getAlertDefinitionEndpoint))
		alertDefinitions.Delete("/:alertDefinitionId", ng.validateOrgAlertDefinition, api.Wrap(ng.deleteAlertDefinitionEndpoint))
		alertDefinitions.Post("/", middleware.ReqSignedIn, binding.Bind(SaveAlertDefinitionCommand{}), api.Wrap(ng.createAlertDefinitionEndpoint))
		alertDefinitions.Put("/:alertDefinitionId", ng.validateOrgAlertDefinition, binding.Bind(updateAlertDefinitionCommand{}), api.Wrap(ng.updateAlertDefinitionEndpoint))
	})
}

// conditionEval handles POST /api/alert-definitions/eval.
func (ng *AlertNG) conditionEval(c *models.ReqContext, dto evalAlertConditionCommand) api.Response {
	alertCtx, cancelFn := context.WithTimeout(context.Background(), setting.AlertingEvaluationTimeout)
	defer cancelFn()

	alertExecCtx := eval.AlertExecCtx{Ctx: alertCtx, SignedInUser: c.SignedInUser}

	fromStr := c.Query("from")
	if fromStr == "" {
		fromStr = "now-3h"
	}

	toStr := c.Query("to")
	if toStr == "" {
		toStr = "now"
	}

	execResult, err := dto.Condition.Execute(alertExecCtx, fromStr, toStr)
	if err != nil {
		return api.Error(400, "Failed to execute conditions", err)
	}

	evalResults, err := eval.EvaluateExecutionResult(execResult)
	if err != nil {
		return api.Error(400, "Failed to evaluate results", err)
	}

	frame := evalResults.AsDataFrame()
	df := tsdb.NewDecodedDataFrames([]*data.Frame{&frame})
	instances, err := df.Encoded()
	if err != nil {
		return api.Error(400, "Failed to encode result dataframes", err)
	}

	return api.JSON(200, util.DynMap{
		"instances": instances,
	})
}

// §AlertDefinitionEval handles GET /api/alert-definitions/eval/:dashboardId/:panelId/:refId".
func (ng *AlertNG) AlertDefinitionEval(c *models.ReqContext) api.Response {
	alertDefinitionID := c.ParamsInt64(":alertDefinitionId")

	fromStr := c.Query("from")
	if fromStr == "" {
		fromStr = "now-3h"
	}

	toStr := c.Query("to")
	if toStr == "" {
		toStr = "now"
	}

	conditions, err := ng.LoadAlertCondition(alertDefinitionID, c.SignedInUser, c.SkipCache)
	if err != nil {
		return api.Error(400, "Failed to load conditions", err)
	}

	alertCtx, cancelFn := context.WithTimeout(context.Background(), setting.AlertingEvaluationTimeout)
	defer cancelFn()

	alertExecCtx := eval.AlertExecCtx{Ctx: alertCtx, SignedInUser: c.SignedInUser}

	execResult, err := conditions.Execute(alertExecCtx, fromStr, toStr)
	if err != nil {
		return api.Error(400, "Failed to execute conditions", err)
	}

	evalResults, err := eval.EvaluateExecutionResult(execResult)
	if err != nil {
		return api.Error(400, "Failed to evaluate results", err)
	}

	frame := evalResults.AsDataFrame()

	df := tsdb.NewDecodedDataFrames([]*data.Frame{&frame})
	instances, err := df.Encoded()
	if err != nil {
		return api.Error(400, "Failed to encode result dataframes", err)
	}

	return api.JSON(200, util.DynMap{
		"instances": instances,
	})
}

// getAlertDefinitionEndpoint handles GET /api/alert-definitions/:alertDefinitionId.
func (ng *AlertNG) getAlertDefinitionEndpoint(c *models.ReqContext) api.Response {
	id := c.ParamsInt64(":alertDefinitionId")

	alertDefinition, err := ng.getAlertDefinitionByID(id)
	if err != nil {
		return api.Error(500, "Failed to get alert definition", err)
	}

	return api.JSON(200, alertDefinition)
}

// deleteAlertDefinitionEndpoint handles DELETE /api/alert-definitions/:alertDefinitionId.
func (ng *AlertNG) deleteAlertDefinitionEndpoint(c *models.ReqContext) api.Response {
	alertDefinitionID := c.ParamsInt64(":alertDefinitionId")

	rowsAffected, err := ng.deleteAlertDefinitionByID(alertDefinitionID)
	if err != nil {
		return api.Error(500, "Failed to delete alert definition", err)
	}

	return api.JSON(200, util.DynMap{"affectedRows": rowsAffected})
}

// updateAlertDefinitionEndpoint handles PUT /api/alert-definitions/:alertDefinitionId.
func (ng *AlertNG) updateAlertDefinitionEndpoint(c *models.ReqContext, cmd updateAlertDefinitionCommand) api.Response {
	id := c.ParamsInt64(":alertDefinitionId")

	var affectedRows int64
	err := ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinition := &AlertDefinition{
			Name:      cmd.Name,
			Condition: cmd.Condition.RefID,
			Data:      cmd.Condition.QueriesAndExpressions,
		}

		if err := ng.validateAlertDefinition(alertDefinition, c.SignedInUser, c.SkipCache); err != nil {
			return err
		}

		if err := alertDefinition.preSave(); err != nil {
			return err
		}

		var err error
		affectedRows, err = sess.ID(id).Update(alertDefinition)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return api.Error(500, "Failed to update alert definition", err)
	}

	return api.JSON(200, util.DynMap{"affectedRows": affectedRows, "id": id})
}

// createAlertDefinitionEndpoint handles POST /api/alert-definitions.
func (ng *AlertNG) createAlertDefinitionEndpoint(c *models.ReqContext, cmd SaveAlertDefinitionCommand) api.Response {
	var alertDefinition *AlertDefinition
	err := ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinition := &AlertDefinition{
			OrgId:     c.SignedInUser.OrgId,
			Name:      cmd.Name,
			Condition: cmd.Condition.RefID,
			Data:      cmd.Condition.QueriesAndExpressions,
		}

		if err := ng.validateAlertDefinition(alertDefinition, c.SignedInUser, c.SkipCache); err != nil {
			return err
		}

		if err := alertDefinition.preSave(); err != nil {
			return err
		}

		if _, err := sess.Insert(alertDefinition); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return api.Error(500, "Failed to create alert definition", err)
	}

	return api.JSON(200, util.DynMap{"id": alertDefinition.Id})
}

// listAlertDefinitions handles GET /api/alert-definitions.
func (ng *AlertNG) listAlertDefinitions(c *models.ReqContext) api.Response {
	alertDefinitions := make([]*AlertDefinition, 0)
	err := ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		q := "SELECT * FROM alert_definition WHERE org_id = ?"
		if err := sess.SQL(q, c.SignedInUser.OrgId).Find(&alertDefinitions); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return api.Error(500, "Failed to list alert definitions", err)
	}

	return api.JSON(200, util.DynMap{"results": alertDefinitions})
}
