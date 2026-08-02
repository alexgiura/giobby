package repository

import (
	"fmt"
	"strings"
)

func buildPatchSet(columns map[string]string, fields map[string]any, startArg int) (string, []any, error) {
	var parts []string
	args := []any{}
	n := startArg
	for key, value := range fields {
		col, ok := columns[key]
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf(`%s=$%d`, col, n))
		args = append(args, value)
		n++
	}
	return strings.Join(parts, ", "), args, nil
}

func emptyAsNull(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
