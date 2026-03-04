package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type contratoReader interface {
	Exists(ctx context.Context, schema string, contInterno int64) (bool, error)
	TonDisponiblesFijacion(ctx context.Context, schema string, contInterno int64, origenTipo string) (float64, error)
	TonDisponiblesReduccion(ctx context.Context, schema string, contInterno int64) (float64, error)
}

type oliquiReader interface {
	ExistsByIDAGN(ctx context.Context, schema string, idagn string) (bool, error)
}
type ctamreReader interface {
	ExistsByIDAGN(ctx context.Context, schema string, idagn string) (bool, error)
}

type Validators struct {
	contrato contratoReader
	oliqui   oliquiReader
	ctamre   ctamreReader
}

func NewValidators(c contratoReader, o oliquiReader, t ctamreReader) *Validators {
	return &Validators{c, o, t}
}

var (
	ErrDuplicateIDAGN         = errors.New("ya existe un registro con el mismo IDAGN")
	ErrContratoNoExiste       = errors.New("el contrato (contInterno) no existe")
	ErrToneladasInsuficientes = errors.New("no hay toneladas disponibles")
)

func (v *Validators) CheckCommon(ctx context.Context, idagn string, contInterno int64, schema string) error {
	if strings.TrimSpace(idagn) == "" {
		return errors.New("IDAGN requerido")
	}
	exists, err := v.contrato.Exists(ctx, schema, contInterno)
	if err != nil {
		return err
	}
	if !exists {
		return ErrContratoNoExiste
	}
	return nil
}

func (v *Validators) CheckOliqui(ctx context.Context, idagn string, contInterno int64, origenTipo string, olTTMin float64, schema string) error {
	if err := v.CheckCommon(ctx, idagn, contInterno, schema); err != nil {
		return err
	}
	ex, err := v.oliqui.ExistsByIDAGN(ctx, idagn, schema)
	if err != nil {
		return err
	}
	if ex {
		return ErrDuplicateIDAGN
	}
	disp, err := v.contrato.TonDisponiblesFijacion(ctx, schema, contInterno, origenTipo)
	if err != nil {
		return err
	}
	if disp < olTTMin {
		return ErrToneladasInsuficientes
	}
	return nil
}

func (v *Validators) CheckCtamre(ctx context.Context, idagn string, contInterno int64, ampRed int, toneladas float64, schema string) error {
	if err := v.CheckCommon(ctx, idagn, contInterno, schema); err != nil {
		return err
	}
	ex, err := v.ctamre.ExistsByIDAGN(ctx, schema, idagn)
	if err != nil {
		return err
	}
	if ex {
		return ErrDuplicateIDAGN
	}
	if ampRed == 2 { // Reducción
		disp, err := v.contrato.TonDisponiblesReduccion(ctx, schema, contInterno)
		if err != nil {
			return err
		}
		if disp < toneladas {
			return ErrToneladasInsuficientes
		}
	}
	return nil
}

func ValidateUniNego(u string) error {
	switch strings.ToUpper(u) {
	case "AC", "CO":
		return nil
	default:
		return fmt.Errorf("uninego inválido: %s (valores permitidos: AC, CO)", u)
	}
}
