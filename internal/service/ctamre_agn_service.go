package service

import (
	"context"
	"database/sql"

	"agn-service/internal/domain"
	"agn-service/internal/oracle"
)

type ctamreMut interface {
	NextOrden(ctx context.Context, tx *sql.Tx) (int64, error)
	Insert(ctx context.Context, tx *sql.Tx, schema string, cols map[string]interface{}) error
}

type CtamreService struct {
	oc           *oracle.Client
	val          *Validators
	contratoRepo contratoMut
	ctamreRepo   ctamreMut
	auditRepo    movauditMut
}

func NewCtamreService(oc *oracle.Client, val *Validators, c contratoMut, t ctamreMut, a movauditMut) *CtamreService {
	return &CtamreService{oc: oc, val: val, contratoRepo: c, ctamreRepo: t, auditRepo: a}
}

func (s *CtamreService) Create(ctx context.Context, in domain.CtamreRequest) error {
	if err := ValidateUniNego(in.UniNego); err != nil {
		return err
	}

	schema := ResolveSchemaByUniNego(in.UniNego)

	if err := s.val.CheckCtamre(ctx, in.IDAGN, in.ContInterno, in.AmpRed, in.Toneladas, schema); err != nil {
		return err
	}

	tx, err := s.oc.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Lock contrato
	if err = s.contratoRepo.LockForUpdate(ctx, tx, schema, in.ContInterno); err != nil {
		return err
	}

	orden, err := s.ctamreRepo.NextOrden(ctx, tx)
	if err != nil {
		return err
	}

	cols := map[string]interface{}{
		"continterno":  in.ContInterno,
		"ordeninter":   0,
		"orden":        orden,
		"fecha":        in.Fecha,
		"toneladas":    in.Toneladas,
		"camiones":     in.Camiones,
		"pizarra":      in.Pizarra,
		"precio":       in.Precio,
		"porcentaje":   in.Porcentaje,
		"impadicional": in.ImpAdicional,
		"ampred":       in.AmpRed,
		"observacion":  truncate(in.Observacion, 200),
		"idagn":        in.IDAGN,
	}
	if err = s.ctamreRepo.Insert(ctx, tx, schema, cols); err != nil {
		return err
	}

	// Actualiza TTMAXFIN (+/- toneladas)
	if err = s.contratoRepo.UpdateTTMaxFinByCtamre(ctx, tx, schema, in.ContInterno, in.Toneladas, in.AmpRed); err != nil {
		return err
	}

	// MOVAUDIT (AMA o REA)

	usuarioNegocio := ResolveUsuarioPorUniNego(in.UniNego)

	auditOrden, err := s.auditRepo.NextOrden(ctx, tx)
	if err != nil {
		return err
	}
	audit := map[string]interface{}{
		"orden":       auditOrden,
		"observacion": "Alta Agronegocios",
		"empresa":     1,
		"uninego":     in.UniNego,
		"zona":        nil,
		"tiporig":     "CON",
		"tipocomp":    "CONT",
		"compinterno": orden,
		"tipoaudit":   mapAuditType(in.AmpRed),
		"comprobante": nil,
		"continterno": in.ContInterno,
		"usuario":     usuarioNegocio,
		"fecha":       sysdateNumber(),
		"hora":        sysTimeNumber(),
	}
	if err = s.auditRepo.InsertAltaAgroneg(ctx, tx, schema, audit); err != nil {
		return err
	}

	return tx.Commit()
}

func mapAuditType(ampRed int) string {
	if ampRed == 1 {
		return "AMA"
	} // Ampliación
	return "REA" // Reducción
}
