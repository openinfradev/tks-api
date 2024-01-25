package filter

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

type Search struct {
	Query    string
	Operator *Operator
	Fields   []string
}

func (s *Search) Scope(schema *schema.Schema) func(*gorm.DB) *gorm.DB {
	if len(s.Fields) == 0 {
		return nil
	}

	return func(tx *gorm.DB) *gorm.DB {
		searchQuery := tx.Session(&gorm.Session{NewDB: true})

		for _, field := range s.Fields {

			f, sch, joinName := getField(field, schema, nil)
			if f == nil {
				continue
			}
			dataType := getDataType(f)
			if dataType == DataTypeUnsupported {
				continue
			}

			if joinName != "" {
				if err := tx.Statement.Parse(tx.Statement.Model); err != nil {
					_ = tx.AddError(err)
					return tx
				}
				tx = join(tx, joinName, schema)
			}

			filter := &Filter{
				Field:    f.DBName,
				Operator: s.Operator,
				Args:     []string{s.Query},
				Or:       true,
			}

			table := tx.Statement.Quote(tableFromJoinName(sch.Table, joinName))

			computed := f.StructField.Tag.Get("computed")
			var fieldExpr string
			if computed != "" {
				fieldExpr = fmt.Sprintf("(%s)", strings.ReplaceAll(computed, clause.CurrentTable, table))
			} else {
				fieldExpr = table + "." + tx.Statement.Quote(f.DBName)
			}

			searchQuery = s.Operator.Function(searchQuery, filter, fieldExpr, dataType)
		}

		return tx.Where(searchQuery)
	}
}
