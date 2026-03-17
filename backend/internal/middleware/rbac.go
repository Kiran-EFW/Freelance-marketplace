package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ContextKeyUserRole is the Fiber locals key for the authenticated user's role.
const ContextKeyUserRole = "user_role"

// RequireRole returns middleware that checks the authenticated user's role
// against the provided list of allowed roles. The role is expected to be set
// in c.Locals(ContextKeyUserRole) by the JWT auth middleware.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		role, ok := c.Locals(ContextKeyUserRole).(string)
		if !ok || role == "" {
			log.Warn().Msg("rbac: no user role found in context")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied: role not set",
			})
		}

		if _, permitted := allowed[role]; !permitted {
			log.Warn().
				Str("role", role).
				Strs("required", roles).
				Msg("rbac: insufficient role")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied: insufficient permissions",
			})
		}

		return c.Next()
	}
}

// RequireAdmin returns middleware that only allows admin and super_admin roles.
func RequireAdmin() fiber.Handler {
	return RequireRole("admin", "super_admin")
}

// RequireProvider returns middleware that only allows the provider role
// (and admin/super_admin).
func RequireProvider() fiber.Handler {
	return RequireRole("provider", "admin", "super_admin")
}

// RequireOwner returns middleware that checks the authenticated user owns the
// resource identified by the given URL parameter. For example,
// RequireOwner("userID") ensures that the :userID route parameter matches the
// authenticated user's ID. Admin and super_admin roles bypass this check.
func RequireOwner(paramName string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Admin/super_admin can access any resource.
		role, _ := c.Locals(ContextKeyUserRole).(string)
		if role == "admin" || role == "super_admin" {
			return c.Next()
		}

		userID := GetUserID(c)
		if userID == uuid.Nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "authentication required",
			})
		}

		paramValue := c.Params(paramName)
		if paramValue == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "missing resource identifier",
			})
		}

		resourceOwnerID, err := uuid.Parse(paramValue)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid resource identifier",
			})
		}

		if userID != resourceOwnerID {
			log.Warn().
				Str("user_id", userID.String()).
				Str("resource_owner", resourceOwnerID.String()).
				Msg("rbac: user does not own resource")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "access denied: you do not own this resource",
			})
		}

		return c.Next()
	}
}
