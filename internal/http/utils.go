package http

import "time"

func parseDate(v string) *time.Time {

	if v == "" {
		return nil
	}

	layout := "2006-01-02"

	t, err := time.Parse(layout, v)
	if err != nil {
		return nil
	}

	return &t
}
