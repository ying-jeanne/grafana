package ngalert

import (
	"context"

	"github.com/grafana/grafana/pkg/services/sqlstore"
)

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
