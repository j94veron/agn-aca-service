package service

import "strings"

func ResolveSchemaByUniNego(uniNego string) string {
	switch strings.ToUpper(uniNego) {
	case "AC":
		return "ACOPIO"
	case "CO":
		return "CORRETAJE"
	default:
		return "AGRONEGOCIOS"
	}
}
