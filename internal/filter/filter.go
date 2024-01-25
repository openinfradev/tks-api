package filter

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Filter struct {
	Field    string
	Operator *Operator
	Args     []string
	Or       bool
}

func (f *Filter) Scope(settings *Settings, sch *schema.Schema) (func(*gorm.DB) *gorm.DB, func(*gorm.DB) *gorm.DB) {
	field, s, joinName := getField(f.Field, sch, &settings.Blacklist)
	if field == nil {
		return nil, nil
	}

	dataType := getDataType(field)

	joinScope := func(tx *gorm.DB) *gorm.DB {
		if dataType == DataTypeUnsupported {
			return tx
		}
		if joinName != "" {
			if err := tx.Statement.Parse(tx.Statement.Model); err != nil {
				tx.AddError(err)
				return tx
			}
			tx = join(tx, joinName, sch)
		}

		return tx
	}

	computed := field.StructField.Tag.Get("computed")

	conditionScope := func(tx *gorm.DB) *gorm.DB {
		if dataType == DataTypeUnsupported {
			return tx
		}

		table := tx.Statement.Quote(tableFromJoinName(s.Table, joinName))
		var fieldExpr string
		if computed != "" {
			fieldExpr = fmt.Sprintf("(%s)", strings.ReplaceAll(computed, clause.CurrentTable, table))
		} else {
			fieldExpr = table + "." + tx.Statement.Quote(field.DBName)
		}

		return f.Operator.Function(tx, f, fieldExpr, dataType)
	}

	return joinScope, conditionScope
}

// Where applies a condition to given transaction, automatically taking the "Or"
// filter value into account.
func (f *Filter) Where(tx *gorm.DB, query string, args ...interface{}) *gorm.DB {
	if f.Or {
		return tx.Or(query, args...)
	}
	return tx.Where(query, args...)
}
