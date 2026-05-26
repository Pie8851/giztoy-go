package peer

import (
	"github.com/GizClaw/gizclaw-go/pkg/gizclaw/api/apitypes"
	"github.com/gofiber/fiber/v2"
)

type getGearConfig500JSONResponse apitypes.ErrorResponse

func (response getGearConfig500JSONResponse) VisitGetPeerConfigResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)
	return ctx.JSON(&response)
}

type putGearConfig500JSONResponse apitypes.ErrorResponse

func (response putGearConfig500JSONResponse) VisitPutPeerConfigResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)
	return ctx.JSON(&response)
}

type getGearInfo500JSONResponse apitypes.ErrorResponse

func (response getGearInfo500JSONResponse) VisitGetPeerInfoResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)
	return ctx.JSON(&response)
}

type refreshGear500JSONResponse apitypes.ErrorResponse

func (response refreshGear500JSONResponse) VisitRefreshPeerResponse(ctx *fiber.Ctx) error {
	ctx.Response().Header.Set("Content-Type", "application/json")
	ctx.Status(500)
	return ctx.JSON(&response)
}
