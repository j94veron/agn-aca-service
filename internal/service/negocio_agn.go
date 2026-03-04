package service

import "strings"

// ResolveUsuarioPorUniNego devuelve el usuario lógico
// que debe usarse según la unidad de negocio.
func ResolveUsuarioPorUniNego(uniNego string) string {
	switch strings.ToUpper(uniNego) {
	case "AC":
		return "ACOPIO"
	case "CO":
		return "CORRETAJE"
	default:
		return "SISTEMA"
	}
}
