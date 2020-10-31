package coredb

// Register dao models for db
func (c *CoreDB) RegisterModels(modelsMap map[string]DbModel) error {
	if len(modelsMap) == 0 {
		return Error{
			Reason: errRegisterModelEmpty,
		}
	}
	c.models = modelsMap
	return nil
}

func (c *CoreDB) MakeSchema() error {
	tx, err := c.createTx()
	if err != nil {
		return err
	}
	return c.txhelper(tx, func() error {
		for tblName, tbl := range c.models {
			sqlStatements := tbl.CreateSQL()
			c.logger.Debugf("create table for: %s", tblName)
			for _, stmt := range sqlStatements {
				c.logger.Debugf("executing %s", stmt)
				if err := tx.Exec(stmt); err != nil {
					return err
				}
			}
		}
		return nil
	})
}


