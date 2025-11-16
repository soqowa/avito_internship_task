package postgres

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)


func buildUUIDInQuery(prefix, suffix string, ids []uuid.UUID) (string, []any) {
	args := make([]any, 0, len(ids))
	placeholders := make([]string, 0, len(ids))
	for i, id := range ids {
		args = append(args, id)
		placeholders = append(placeholders, "$"+strconv.Itoa(i+1))
	}
	query := prefix + strings.Join(placeholders, ",") + suffix
	return query, args
}
