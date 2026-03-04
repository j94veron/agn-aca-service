package service

import (
	"context"
	"database/sql"
	_ "fmt"
	"time"

	"agn-service/internal/domain"
	"agn-service/internal/oracle"
)

type contratoMut interface {
	LockForUpdate(ctx context.Context, tx *sql.Tx, schema string, contInterno int64) error
	UpdateFijadas(ctx context.Context, tx *sql.Tx, schema string, contInterno int64, origenTipo string, olTTMin float64, tNegocio string) error
	ReadNegocio(ctx context.Context, schema string, contInterno int64) (string, error)
	UpdateTTMaxFinByCtamre(ctx context.Context, tx *sql.Tx, schema string, contInterno int64, toneladas float64, ampRed int) error
}
type oliquiMut interface {
	Insert(ctx context.Context, tx *sql.Tx, schema string, cols map[string]interface{}) (int64, error)
}
type movauditMut interface {
	NextOrden(ctx context.Context, tx *sql.Tx) (int64, error)
	InsertAltaAgroneg(ctx context.Context, tx *sql.Tx, schema string, cols map[string]interface{}) error
}

type oliquiQueryRepo interface {
	CountTooltip(ctx context.Context, f domain.OliquiFilter) (int, error)
	GetTooltipPage(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiTooltipDTO, error)

	CountGrid(ctx context.Context, f domain.OliquiFilter) (int, error)
	GetGridPage(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiGridDTO, error)
}

type OliquiService struct {
	oc           *oracle.Client
	val          *Validators
	contratoRepo contratoMut
	oliquiRepo   oliquiMut
	queryRepo    oliquiQueryRepo
	auditRepo    movauditMut
}

func NewOliquiService(oc *oracle.Client, val *Validators, c contratoMut, o oliquiMut, a movauditMut, q oliquiQueryRepo) *OliquiService {
	return &OliquiService{oc: oc, val: val, contratoRepo: c, oliquiRepo: o, auditRepo: a, queryRepo: q}
}

func (s *OliquiService) Create(ctx context.Context, in domain.OliqUiRequest) error {
	if err := ValidateUniNego(in.UniNego); err != nil {
		return err
	}

	schema := ResolveSchemaByUniNego(in.UniNego)

	if err := s.val.CheckOliqui(ctx, in.IDAGN, in.ContInterno, in.OrigenTipo, in.OlTTMin, schema); err != nil {
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

	// Lock del contrato
	if err = s.contratoRepo.LockForUpdate(ctx, tx, schema, in.ContInterno); err != nil {
		return err
	}

	// Leer negocio para la regla de actualización de fijadas
	tNegocio, err := s.contratoRepo.ReadNegocio(ctx, schema, in.ContInterno)
	if err != nil {
		return err
	}

	// Secuencias
	ordenInter, ordenLiq, auditOrden := time.Now().UnixNano(), time.Now().UnixNano()+1, int64(0)
	if auditOrden, err = s.auditRepo.NextOrden(ctx, tx); err != nil {
		return err
	}

	// Insert OLIQUI (mapear columnas clave; el resto se completa luego si hace falta)
	cols := map[string]interface{}{
		"empresa":      1,
		"continterno":  in.ContInterno,
		"contrato":     nil,        // TODO: leer campo desde Contrato
		"negocio":      nil,        // TODO: idem
		"cuenta":       nil,        // TODO
		"uninego":      in.UniNego, // TODO
		"zona":         nil,        // TODO
		"ordeninter":   ordenInter,
		"ordenliquid":  ordenLiq,
		"tiporden":     mapOrigenTipoToTipOrden(in.OrigenTipo),
		"feccreacion":  in.FecCreacion,
		"fecdesde":     in.FecDesde,
		"diasvto":      in.DiasVto,
		"fechasta":     in.FechaHasta,
		"fechapizarra": in.FechaPizarra,
		"horafijar":    in.HoraFijar,
		"ol_ttmin":     in.OlTTMin,
		"ol_ttmax":     in.OlTTMax,
		"ol_tipprec":   nil, // from Contrato TIPOPRECIO si aplica
		"ol_tippiza":   nil, // from Contrato TIPOPIZARRA
		"ol_pizarra":   nil, // from Contrato PIZARRA
		"ol_precio":    in.Precio,
		"pordenliq":    in.POrdenLiq,
		"moneda":       normalizeMoneda(in.MonFijada, in.Moneda),
		"status":       defaultStatus(in.OrigenTipo),
		"observa":      truncate(in.Observaciones, 500),
		"idagn":        in.IDAGN,
	}
	if _, err = s.oliquiRepo.Insert(ctx, tx, schema, cols); err != nil {
		return err
	}

	// Actualizar fijadas según negocio
	if err = s.contratoRepo.UpdateFijadas(ctx, tx, schema, in.ContInterno, in.OrigenTipo, in.OlTTMin, tNegocio); err != nil {
		return err
	}

	usuarioNegocio := ResolveUsuarioPorUniNego(in.UniNego)

	// MOVAUDIT
	audit := map[string]interface{}{
		"orden":       auditOrden,
		"observacion": "Alta Agronegocios",
		"empresa":     1,
		"uninego":     in.UniNego, // TODO leer de Contrato
		"zona":        nil,        // TODO
		"tiporig":     mapTipOrig(in.OrigenTipo),
		"tipocomp":    mapTipComp(in.OrigenTipo),
		"compinterno": ordenLiq,
		"tipoaudit":   "ALT",
		"comprobante": nil, // contrato
		"continterno": in.ContInterno,
		"usuario":     usuarioNegocio, // o el user autenticado
		"fecha":       sysdateNumber(),
		"hora":        sysTimeNumber(),
	}
	if err = s.auditRepo.InsertAltaAgroneg(ctx, tx, schema, audit); err != nil {
		return err
	}

	return tx.Commit()
}

func mapOrigenTipoToTipOrden(s string) int {
	if s == "OP" {
		return 5
	}
	return 2
}
func mapTipOrig(s string) string {
	if s == "OL" {
		return "LIQ"
	}
	return "OFI"
}
func mapTipComp(s string) string {
	if s == "OL" {
		return "LIQ"
	}
	return "OFI"
}
func defaultStatus(origen string) string {
	if origen == "OL" {
		return "PI"
	}
	return "PA"
}
func normalizeMoneda(monFijada, moneda string) string {
	// Reglas mínimas: si '1' => DO, sino PE
	if monFijada == "1" || moneda == "1" {
		return "DO"
	}
	return "PE"
}
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
func sysdateNumber() int64 {
	// YYYYMMDD (ajustá a tu Ndate_C si necesitás)
	t := time.Now()
	return int64(t.Year()*10000 + int(t.Month())*100 + t.Day())
}
func sysTimeNumber() int64 {
	// hhmmss
	t := time.Now()
	return int64(t.Hour()*10000 + t.Minute()*100 + t.Second())
}

func (s *OliquiService) GetTooltip(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiTooltipDTO, int, error) {

	total, err := s.queryRepo.CountTooltip(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	items, err := s.queryRepo.GetTooltipPage(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (s *OliquiService) GetGrid(ctx context.Context, f domain.OliquiFilter) ([]domain.OliquiGridDTO, int, error) {

	total, err := s.queryRepo.CountGrid(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	items, err := s.queryRepo.GetGridPage(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
