package handler

import (
	"net/http"

	"github.com/awwal/voxmeet/api-gateway/internal/auth"
	"github.com/awwal/voxmeet/api-gateway/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CurrentUser returns a handler that returns the authenticated user's profile.
func CurrentUser(queries db.Querier) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.UserIDFromContext(r.Context())
		if userID == "" {
			respondError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		uid, err := uuid.Parse(userID)
		if err != nil {
			respondError(w, http.StatusBadRequest, "invalid user id")
			return
		}

		var id pgtype.UUID
		id.Bytes = uid
		id.Valid = true

		user, err := queries.GetUserById(r.Context(), id)
		if err != nil {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":           userIDToString(user.ID),
			"username":     user.Username,
			"email":        user.Email,
			"display_name": user.DisplayName.String,
		})
	}
}
