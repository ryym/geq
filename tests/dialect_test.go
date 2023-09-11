package tests

import (
	"testing"

	"github.com/ryym/geq"
	"github.com/ryym/geq/tests/b"
)

func TestQueryVariations(t *testing.T) {
	q := b.Select(b.Concat("a", "b", "c"))
	err := assertQueryWith(&geq.DialectGeneric{}, q, "SELECT ? || ? || ?", "a", "b", "c")
	if err != nil {
		t.Error(err)
	}
	err = assertQueryWith(&geq.DialectPostgres{}, q, "SELECT $1 || $2 || $3", "a", "b", "c")
	if err != nil {
		t.Error(err)
	}
	err = assertQueryWith(&geq.DialectMySQL{}, q, "SELECT CONCAT(?, ?, ?)", "a", "b", "c")
	if err != nil {
		t.Error(err)
	}
}
