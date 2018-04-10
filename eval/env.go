package eval

import "fmt"

// Environment represents an evaluation context for SQL statements.
type Environment struct {
	tables map[string]*Table // key is table name
}

func (env *Environment) CreateTable(table *Table) error {
	// Check for duplicate tables.
	for name := range env.tables {
		if name == table.Name {
			return fmt.Errorf("relation %q already exists", name)
		}
	}
	// Check for duplicate columns.
	seen := make(map[string]bool)
	for _, col := range table.Columns {
		if seen[col.Name] {
			return fmt.Errorf("column %q specified more than once", col.Name)
		}
		seen[col.Name] = true
	}
	// OK: add the table.
	if env.tables == nil {
		env.tables = make(map[string]*Table)
	}
	env.tables[table.Name] = table
	return nil
}

// Retrieves a table by name.
func (env *Environment) lookupTable(name string) *Table {
	if tab, ok := env.tables[name]; ok {
		return tab
	}
	panic(fmt.Errorf("relation %q does not exist", name))
}
