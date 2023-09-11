package tests

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ryym/geq"
)

type testCase struct {
	name string
	data string
	run  func(*sql.Tx) error
}

func runTestCases(t *testing.T, db *sql.DB, cases []testCase) {
	for i, c := range cases {
		runTestCase(t, db, i, c)
	}
}

func runTestCase(t *testing.T, db *sql.DB, idx int, c testCase) {
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("failed to begin tx: %v", err)
	}

	defer func() {
		err = tx.Rollback()
		if err != nil {
			t.Fatalf("failed to rollback tx: %v", err)
		}
	}()

	if c.data != "" {
		_, err = tx.Exec(c.data)
		if err != nil {
			t.Fatalf("failed to execute data query: %v\nquery: %s", err, c.data)
		}
	}

	err = c.run(tx)
	if err != nil {
		t.Logf("FAILED: case[%d] %s", idx, c.name)
		t.Error(err)
	}
}

func assertEqual[V any](got, want V) error {
	diff := cmp.Diff(want, got)
	if diff != "" {
		return fmt.Errorf("values not equal:\n%s", diff)
	}
	return nil
}

func assertQuery(q geq.AnyQuery, wantSQL string, wantArgs ...any) error {
	return assertQueryWith(&geq.DialectGeneric{}, q, wantSQL, wantArgs...)
}

func assertQueryWith(d geq.Dialect, q geq.AnyQuery, wantSQL string, wantArgs ...any) error {
	qcfg := geq.NewQueryConfig(d)
	got, err := q.FinalizeWith(qcfg)
	if err != nil {
		return err
	}
	want := &geq.FinalQuery{Query: wantSQL, Args: wantArgs}
	return assertEqual(got, want)
}

func sjoin(ss ...string) string {
	return strings.Join(ss, " ")
}
