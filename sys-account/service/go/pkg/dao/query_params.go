package dao

import "github.com/getcouragenow/sys/sys-account/service/go/pkg/utilities"

// QueryParams can be any condition
type QueryParams struct {
	Params map[string]interface{}
}

func (qp *QueryParams) ColumnsAndValues() ([]string, []interface{}) {
	var columns []string
	var values []interface{}
	for k, v := range qp.Params {
		columns = append(columns, utilities.ToSnakeCase(k))
		values = append(values, v)
	}
	return columns, values
}
