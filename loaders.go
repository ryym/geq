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

type SliceLoader[Q, R any] struct {
	query  *Query[Q]
	mapper RowMapper[R]
}

func (l *SliceLoader[Q, R]) Load(ctx context.Context, db *sql.DB) ([]R, error) {
	var dest []R
	scanner := &SliceScanner[R]{mapper: l.mapper, dest: &dest}
	err := loadBySingleScanner(ctx, db, scanner, l.query)
	if err != nil {
		return nil, err
	}
	return dest, nil
}

type MapLoader[Q, R any, K comparable] struct {
	query  *Query[Q]
	mapper RowMapper[R]
	key    *Column[K]
}

func (l *MapLoader[Q, R, K]) Load(ctx context.Context, db *sql.DB) (map[K]R, error) {
	var dest map[K]R
	scanner := &MapScanner[R, K]{mapper: l.mapper, key: l.key, dest: &dest}
	err := loadBySingleScanner(ctx, db, scanner, l.query)
	if err != nil {
		return nil, err
	}
	return dest, nil
}

func loadBySingleScanner[Q any](ctx context.Context, db *sql.DB, s RowsScanner, q *Query[Q]) (err error) {
	fq := q.Finalize()
	rows, err := db.QueryContext(ctx, fq.Query, fq.Args...)
	if err != nil {
		return err
	}
	for i := 0; rows.Next(); i++ {
		ptrs := s.BeforeEachScan(i, q.selections)
		err = rows.Scan(ptrs...)
		if err != nil {
			return err
		}
		s.AfterEachScan(ptrs)
	}
	return nil
}
