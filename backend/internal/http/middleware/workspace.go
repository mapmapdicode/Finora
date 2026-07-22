package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"wealthos-backend/internal/domain"
	"wealthos-backend/internal/storage"
)

func WorkspaceMembershipMiddleware(store storage.Store, optional bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := currentUserID(c)
		if userID == "" {
			if optional {
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "missing user context"})
			c.Abort()
			return
		}

		requested := strings.TrimSpace(c.GetHeader("x-workspace-id"))
		if requested == "" {
			requested = strings.TrimSpace(c.Query("workspaceId"))
		}

		workspaceID := domain.ID(requested)
		role := domain.Role("")

		if workspaceID == "" {
			if optional {
				c.Next()
				return
			}
			workspaces := store.ListWorkspaces(domain.ID(userID))
			if len(workspaces) == 0 {
				c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "user has no workspace"})
				c.Abort()
				return
			}
			workspaceID = workspaces[0].ID
		}

		role, ok := store.GetWorkspaceMemberRole(domain.ID(userID), workspaceID)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN", "message": "workspace is not accessible"})
			c.Abort()
			return
		}
		c.Set("workspace_role", string(role))

		c.Set("workspace_id", string(workspaceID))
		c.Next()
	}
}

func currentUserID(c *gin.Context) string {
	if val, ok := c.Get("user_id"); ok {
		if uid, ok2 := val.(string); ok2 {
			return strings.TrimSpace(uid)
		}
	}
	return ""
}
