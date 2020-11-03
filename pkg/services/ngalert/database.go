package ngalert

import (
	"context"

	"github.com/grafana/grafana/pkg/services/sqlstore"
)

func (ng *AlertNG) registerBusHandlers() {
	ng.Bus.AddHandler(ng.SaveAlertDefinition)
	ng.Bus.AddHandler(ng.UpdateAlertDefinition)
	ng.Bus.AddHandler(ng.GetAlertDefinitions)
}

func getAlertDefinitionByID(alertDefinitionID int64, sess *sqlstore.DBSession) (*AlertDefinition, error) {
	alertDefinition := AlertDefinition{}
	has, err := sess.ID(alertDefinitionID).Get(&alertDefinition)
	if !has {
		return nil, ErrAlertDefinitionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &alertDefinition, nil
}

// deleteAlertDefinitionByID deletes an alert definition.
// It returns models.ErrAlertDefinitionNotFound if no alert definition is found for the provided ID.
func (ng *AlertNG) deleteAlertDefinitionByID(id int64) (int64, error) {
	var rowsAffected int64
	err := ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		res, err := sess.Exec("DELETE FROM alert_definition WHERE id = ?", id)
		if err != nil {
			return err
		}

		rowsAffected, err = res.RowsAffected()
		if err != nil {
			return err
		}
		return nil
	})

	return rowsAffected, err
}

// getAlertDefinitionByID gets an alert definition from the database by its ID.
// It returns models.ErrAlertDefinitionNotFound if no alert definition is found for the provided ID.
func (ng *AlertNG) getAlertDefinitionByID(id int64) (*AlertDefinition, error) {
	var alertDefinition *AlertDefinition
	if err := ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		var err error
		alertDefinition, err = getAlertDefinitionByID(id, sess)
		return err
	}); err != nil {
		return nil, err
	}

	return alertDefinition, nil
}

// SaveAlertDefinition is a handler for saving a new alert definition.
func (ng *AlertNG) SaveAlertDefinition(cmd *SaveAlertDefinitionCommand) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinition := &AlertDefinition{
			OrgId:     cmd.OrgID,
			Name:      cmd.Name,
			Condition: cmd.Condition.RefID,
			Data:      cmd.Condition.QueriesAndExpressions,
		}

		if err := ng.validateAlertDefinition(alertDefinition, cmd.SignedInUser, cmd.SkipCache); err != nil {
			return err
		}

		if err := alertDefinition.preSave(); err != nil {
			return err
		}

		if _, err := sess.Insert(alertDefinition); err != nil {
			return err
		}

		cmd.Result = alertDefinition
		return nil
	})
}

// UpdateAlertDefinition is a handler for updating an existing alert definition.
// It returns models.ErrAlertDefinitionNotFound if no alert definition is found for the provided ID.
func (ng *AlertNG) UpdateAlertDefinition(cmd *UpdateAlertDefinitionCommand) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinition := &AlertDefinition{
			Name:      cmd.Name,
			Condition: cmd.Condition.RefID,
			Data:      cmd.Condition.QueriesAndExpressions,
		}

		if err := ng.validateAlertDefinition(alertDefinition, cmd.SignedInUser, cmd.SkipCache); err != nil {
			return err
		}

		if err := alertDefinition.preSave(); err != nil {
			return err
		}

		affectedRows, err := sess.ID(cmd.ID).Update(alertDefinition)
		if err != nil {
			return err
		}

		cmd.Result = alertDefinition
		cmd.RowsAffected = affectedRows
		return nil
	})
}

// GetAlertDefinitions is a handler for retrieving alert definitions of specific organisation.
func (ng *AlertNG) GetAlertDefinitions(cmd *ListAlertDefinitionsCommand) error {
	return ng.SQLStore.WithTransactionalDbSession(context.Background(), func(sess *sqlstore.DBSession) error {
		alertDefinitions := make([]*AlertDefinition, 0)
		q := "SELECT * FROM alert_definition WHERE org_id = ?"
		if err := sess.SQL(q, cmd.OrgID).Find(&alertDefinitions); err != nil {
			return err
		}

		cmd.Result = alertDefinitions
		return nil
	})
}
