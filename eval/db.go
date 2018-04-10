package eval

import "fmt"

// Column contains column metadata.
type Column struct {
	Name     string
	Type     DataType
	Default  Value
	Nullable bool
	AutoInc  bool
	NextVal  int
}

// Row is an array of values.
type Row []Value

// Table contains table metadata plus its rows.
type Table struct {
	Name    string
	Columns []*Column
	Data    []Row
}

// Returns the index of the named column, or -1 if the column does not exist.
func (t *Table) colIndex(name string) int {
	for i, c := range t.Columns {
		if c.Name == name {
			return i
		}
	}
	return -1
}

func (tab *Table) insert(names []string, values []Value) {
	// Verify that the number of column names matches the number of values. No
	// names is equivalent to specifying every column in table order.
	numTargets := len(names)
	if numTargets == 0 {
		numTargets = len(tab.Columns)
	}
	if numTargets != len(values) {
		panic(fmt.Errorf("INSERT has %d expressions but %d target columns", len(values), numTargets))
	}

	// Create the row in proper column order.
	row := make([]Value, len(tab.Columns))
	if len(names) != 0 {
		// Verify that every name refers to a valid column.
		for _, name := range names {
			if tab.colIndex(name) < 0 {
				panic(fmt.Errorf("column %q of relation %q does not exist", name, tab.Name))
			}
		}
	next:
		for i, col := range tab.Columns {
			for j, name := range names {
				if col.Name == name {
					row[i] = values[j]
					continue next
				}
			}
			// If the column was not specified in the insert statement, the
			// auto-increment setting and default value become relevant.
			if col.Type == Integer && col.AutoInc {
				row[i] = IntegerValue(col.NextVal)
				col.NextVal++
			} else if col.Default != nil {
				row[i] = col.Default
			}
		}
	} else {
		// No names were given, so the values must be in table order.
		copy(row, values)
	}

	// Enforce not-null constraints and coerce to the target data type.
	for i, value := range row {
		col := tab.Columns[i]
		if value != nil {
			row[i] = coerce(value, col.Type)
		} else if !col.Nullable {
			panic(fmt.Errorf("null value in column %q violates not-null constraint", col.Name))
		}

	}

	// Append the row to the table.
	tab.Data = append(tab.Data, row)
}
