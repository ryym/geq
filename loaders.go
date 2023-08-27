package geq

import (
	"context"
	"database/sql"
)

type MultiScanLoader[Q any] struct {
	query    *Query[Q]
	scanners []RowsScanner
}

func (ms *MultiScanLoader[Q]) Load(ctx context.Context, db *sql.DB) (err error) {
	q := ms.query.Finalize()
	rows, err := db.QueryContext(ctx, q.Query, q.Args...)
	if err != nil {
		return err
	}
	for idx := 0; rows.Next(); idx++ {
		allPtrs := make([]any, 0)
		ptrGroups := make([][]any, len(ms.scanners))
		for i, s := range ms.scanners {
			ptrs := s.BeforeEachScan(idx, ms.query.selections)
			ptrGroups[i] = ptrs
			allPtrs = append(allPtrs, ptrs...)
		}
		err = rows.Scan(allPtrs...)
		if err != nil {
			return err
		}
		for i, s := range ms.scanners {
			s.AfterEachScan(ptrGroups[i])
		}
	}
	return nil
}
