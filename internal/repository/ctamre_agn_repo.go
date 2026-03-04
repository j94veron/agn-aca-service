package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"agn-service/internal/oracle"
)

type CtamreRepo struct{ oc *oracle.Client }

func NewCtamreRepo(oc *oracle.Client) *CtamreRepo { return &CtamreRepo{oc: oc} }

func (r *CtamreRepo) ExistsByIDAGN(ctx context.Context, schema string, idagn string) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM %s.CTAMRE WHERE IDAGN = :1`, schema)
	row := r.oc.DB.QueryRowContext(ctx, q, idagn)
	var x int
	err := row.Scan(&x)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (r *CtamreRepo) NextOrden(ctx context.Context, tx *sql.Tx) (int64, error) {
	row := tx.QueryRowContext(ctx, `SELECT numcomer13.nextval FROM DUAL`)
	var n int64
	return n, row.Scan(&n)
}

func (r *CtamreRepo) Insert(ctx context.Context, tx *sql.Tx, schema string, cols map[string]interface{}) error {
	q := fmt.Sprintf(`
		INSERT INTO %s.CTAMRE (
			CONTINTERNO, ORDENINTER, ORDEN, FECHA, TONELADAS, CAMIONES,
			PIZARRA, PRECIO, PORCENTAJE, IMPADICIONAL, NUEVO, BAJA,
			AMPRED, OBSERVACION, IDAGN
		)
		VALUES (
			:continterno, :ordeninter, :orden, :fecha, :toneladas, :camiones,
			:pizarra, :precio, :porcentaje, :impadicional, 0, 0,
			:ampred, :observacion, :idagn
		)`, schema)
	_, err := tx.ExecContext(ctx, q, cols)
	return err
}
