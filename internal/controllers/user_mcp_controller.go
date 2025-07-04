package controllers

import (
	"github.com/AkashKesav/API2SDK/internal/models"
	"github.com/AkashKesav/API2SDK/internal/services"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserMCPController handles the requests for user-specific MCP instances.
type UserMCPController struct {
	mcpInstanceService services.MCPInstanceService
	integrationService services.IntegrationService
}

// NewUserMCPController creates a new UserMCPController.
func NewUserMCPController(mcpInstanceService services.MCPInstanceService, integrationService services.IntegrationService) *UserMCPController {
	return &UserMCPController{
		mcpInstanceService: mcpInstanceService,
		integrationService: integrationService,
	}
}

// CreateMCPInstance creates a new MCPInstance for the authenticated user.
func (c *UserMCPController) CreateMCPInstance(ctx fiber.Ctx) error {
	var req struct {
		IntegrationID string `json:"integrationId"`
	}

	if err := ctx.Bind().Body(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, ok := ctx.Locals("user_id").(string)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	mcpInstance := &models.MCPInstance{
		UserID:        userID,
		IntegrationID: req.IntegrationID,
	}

	createdInstance, err := c.mcpInstanceService.CreateMCPInstance(ctx.Context(), mcpInstance)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create MCP instance"})
	}

	return ctx.Status(fiber.StatusCreated).JSON(createdInstance)
}

// ListMCPInstances lists all MCPInstances for the authenticated user.
func (c *UserMCPController) ListMCPInstances(ctx fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(string)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	instances, err := c.mcpInstanceService.ListMCPInstances(ctx.Context(), userID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list MCP instances"})
	}

	return ctx.JSON(instances)
}

// DeleteMCPInstance deletes a specific MCPInstance for the authenticated user.
func (c *UserMCPController) DeleteMCPInstance(ctx fiber.Ctx) error {
	instanceID := ctx.Params("instanceID")
	objID, err := primitive.ObjectIDFromHex(instanceID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid instance ID"})
	}

	userID, ok := ctx.Locals("user_id").(string)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	instance, err := c.mcpInstanceService.GetMCPInstance(ctx.Context(), objID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "instance not found"})
	}

	if instance.UserID != userID {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	if err := c.mcpInstanceService.DeleteMCPInstance(ctx.Context(), objID); err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete MCP instance"})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

// ListResources lists all Resources for a specific MCPInstance.
func (c *UserMCPController) ListResources(ctx fiber.Ctx) error {
	instanceID := ctx.Params("instanceID")
	objID, err := primitive.ObjectIDFromHex(instanceID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid instance ID"})
	}

	userID, ok := ctx.Locals("user_id").(string)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	instance, err := c.mcpInstanceService.GetMCPInstance(ctx.Context(), objID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "instance not found"})
	}

	if instance.UserID != userID {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "forbidden"})
	}

	resources, err := c.mcpInstanceService.GetResources(ctx.Context(), objID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list resources"})
	}

	return ctx.JSON(resources)
}
