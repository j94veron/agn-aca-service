package repository

import (
	"context"
	"database/sql"
	"fmt"

	"agn-service/internal/oracle"
)

type MovAuditRepo struct{ oc *oracle.Client }

func NewMovAuditRepo(oc *oracle.Client) *MovAuditRepo { return &MovAuditRepo{oc: oc} }

func (r *MovAuditRepo) NextOrden(ctx context.Context, tx *sql.Tx) (int64, error) {
	row := tx.QueryRowContext(ctx, `SELECT numcomer40.nextval FROM DUAL`)
	var n int64
	return n, row.Scan(&n)
}

func (r *MovAuditRepo) InsertAltaAgroneg(ctx context.Context, tx *sql.Tx, schema string, cols map[string]interface{}) error {
	q := fmt.Sprintf(`
		INSERT INTO %s.MOVAUDIT (
			ORDEN, OBSERVACION, EMPRESA, UNINEGO, ZONA,
			TIPORIG, TIPOCOMP, COMPINTERNO, TIPOAUDIT, COMPROBANTE,
			CONTINTERNO, USUARIO, FECHA, HORA, NUEVO, BAJA
		) VALUES (
			:orden, :observacion, :empresa, :uninego, :zona,
			:tiporig, :tipocomp, :compinterno, :tipoaudit, :comprobante,
			:continterno, :usuario, :fecha, :hora, 0, 0
		)`, schema)
	_, err := tx.ExecContext(ctx, q, cols)
	return err
}
