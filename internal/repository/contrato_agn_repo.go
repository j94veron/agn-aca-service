package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"agn-service/internal/oracle"
)

type ContratoRepo struct{ oc *oracle.Client }

func NewContratoRepo(oc *oracle.Client) *ContratoRepo { return &ContratoRepo{oc: oc} }

func (r *ContratoRepo) Exists(ctx context.Context, schema string, contInterno int64) (bool, error) {
	q := fmt.Sprintf(`SELECT 1 FROM %s.CONTRATO WHERE contInterno = :1`, schema)
	row := r.oc.DB.QueryRowContext(ctx, q, contInterno)
	var x int
	err := row.Scan(&x)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Lock para escritura concurrente
func (r *ContratoRepo) LockForUpdate(ctx context.Context, tx *sql.Tx, schema string, contInterno int64) error {
	q := fmt.Sprintf(`SELECT contInterno FROM %s.CONTRATO WHERE contInterno = :1 FOR UPDATE NOWAIT`, schema)
	_, err := tx.ExecContext(ctx, q, contInterno)
	return err
}

func (r *ContratoRepo) TonDisponiblesFijacion(ctx context.Context, schema string, contInterno int64, origenTipo string) (float64, error) {
	q := fmt.Sprintf(`
		SELECT NVL(TTMAXFIN,0) - NVL(ETTFIJADAS,0)
		FROM %s.CONTRATO WHERE contInterno = :1`, schema)
	row := r.oc.DB.QueryRowContext(ctx, q, contInterno)
	var disp float64
	if err := row.Scan(&disp); err != nil {
		return 0, err
	}
	if disp < 0 {
		disp = 0
	}
	return disp, nil
}

func (r *ContratoRepo) TonDisponiblesReduccion(ctx context.Context, schema string, contInterno int64) (float64, error) {
	q := fmt.Sprintf(`
		SELECT NVL(TTMAXFIN,0) - NVL(ETTFIJADAS,0)
		FROM %s.CONTRATO WHERE contInterno = :1`, schema)
	row := r.oc.DB.QueryRowContext(ctx, q, contInterno)
	var disp float64
	if err := row.Scan(&disp); err != nil {
		return 0, err
	}
	if disp < 0 {
		disp = 0
	}
	return disp, nil
}

func (r *ContratoRepo) UpdateFijadas(ctx context.Context, tx *sql.Tx, schema string, contInterno int64, origenTipo string, olTTMin float64, tNegocio string) error {
	q := fmt.Sprintf(`
		UPDATE %s.CONTRATO
		   SET ETTFIJADAS = CASE WHEN :2 = 'VEN' THEN NVL(ETTFIJADAS,0) + :3 ELSE NVL(ETTFIJADAS,0) END,
		       ITTFIJADAS = CASE WHEN :2 IN ('COM','CV') THEN NVL(ITTFIJADAS,0) + :3 ELSE NVL(ITTFIJADAS,0) END
		 WHERE contInterno = :1`, schema)
	_, err := tx.ExecContext(ctx, q, contInterno, tNegocio, olTTMin)
	return err
}

func (r *ContratoRepo) UpdateTTMaxFinByCtamre(ctx context.Context, tx *sql.Tx, schema string, contInterno int64, toneladas float64, ampRed int) error {
	sign := 1
	if ampRed == 2 {
		sign = -1
	}
	q := fmt.Sprintf(`UPDATE %s.CONTRATO SET TTMAXFIN = NVL(TTMAXFIN,0) + (:2 * :3) WHERE contInterno = :1`, schema)
	_, err := tx.ExecContext(ctx, q, contInterno, toneladas, sign)
	return err
}

func (r *ContratoRepo) ReadNegocio(ctx context.Context, schema string, contInterno int64) (string, error) {
	q := fmt.Sprintf(`SELECT NVL(TNEGOCIO,'') FROM %s.CONTRATO WHERE contInterno = :1`, schema)
	row := r.oc.DB.QueryRowContext(ctx, q, contInterno)
	var t string
	err := row.Scan(&t)
	if err != nil {
		return "", fmt.Errorf("read negocio: %w", err)
	}
	return t, nil
}
