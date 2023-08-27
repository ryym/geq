package tests

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ryym/geq"
)

func TestBuiltQueries(t *testing.T) {
	b := NewQueryBuilder()
	var fq *geq.FinalQuery

	fq = b.Users.Query().Finalize()
	assertFinalQuery(t, fq, "SELECT users.id,users.name FROM users")
}

func assertFinalQuery(t *testing.T, got *geq.FinalQuery, wantQuery string, wantArgs ...any) {
	want := &geq.FinalQuery{Query: wantQuery, Args: wantArgs}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("wrong final query:%s", diff)
	}
}
