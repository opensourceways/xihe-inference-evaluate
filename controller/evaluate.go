package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/opensourceways/xihe-inference-evaluate/app"
	"github.com/opensourceways/xihe-inference-evaluate/domain/evaluate"
)

func AddRouterForEvaluateController(
	rg *gin.RouterGroup,
	manager evaluate.Evaluate,
) {
	ctl := EvaluateController{
		s: app.NewEvaluateService(manager),
	}

	rg.POST("/v1/evaluate/project/:type", ctl.Create)
	rg.PUT("/v1/evaluate/project", ctl.ExtendExpiry)
}

type EvaluateController struct {
	baseController

	s app.EvaluateService
}

// @Summary Create
// @Description create evaluate
// @Tags  Evaluate
// @Accept json
// @Success 201
// @Failure 400 bad_request_body    can't parse request body
// @Failure 401 bad_request_param   some parameter of body is invalid
// @Failure 500 system_error        system error
// @Router /v1/evaluate/project/{type} [post]
func (ctl *EvaluateController) Create(ctx *gin.Context) {
	switch ctx.Param("type") {
	case "custom":
		ctl.createCustom(ctx)
	case "standard":
		ctl.createStandard(ctx)
	default:
		ctx.JSON(http.StatusBadRequest, newResponseCodeMsg(
			errorBadRequestParam, "unknown type",
		))
	}
}

func (ctl *EvaluateController) createCustom(ctx *gin.Context) {
	req := CustomEvaluateCreateRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, respBadRequestBody)

		return
	}

	cmd, err := req.toCmd()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newResponseCodeError(
			errorBadRequestParam, err,
		))

		return
	}

	if err = ctl.s.CreateCustom(&cmd); err != nil {
		ctl.sendRespWithInternalError(ctx, newResponseError(err))

		return
	}

	ctx.JSON(http.StatusCreated, newResponseData("successfully"))
}

func (ctl *EvaluateController) createStandard(ctx *gin.Context) {
	req := StandardEvaluateCreateRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, respBadRequestBody)

		return
	}

	cmd, err := req.toCmd()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newResponseCodeError(
			errorBadRequestParam, err,
		))

		return
	}

	if err = ctl.s.CreateStandard(&cmd); err != nil {
		ctl.sendRespWithInternalError(ctx, newResponseError(err))

		return
	}

	ctx.JSON(http.StatusCreated, newResponseData("successfully"))
}

// @Summary ExtendExpiry
// @Description extend expiry for evaluate
// @Tags  Evaluate
// @Accept json
// @Success 202
// @Failure 400 bad_request_body    can't parse request body
// @Failure 401 bad_request_param   some parameter of body is invalid
// @Failure 500 system_error        system error
// @Router /v1/evaluate/project [put]
func (ctl *EvaluateController) ExtendExpiry(ctx *gin.Context) {
	req := EvaluateUpdateRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, respBadRequestBody)

		return
	}

	cmd, err := req.toCmd()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, newResponseCodeError(
			errorBadRequestParam, err,
		))

		return
	}

	if err = ctl.s.ExtendExpiry(&cmd); err != nil {
		ctl.sendRespWithInternalError(ctx, newResponseError(err))

		return
	}

	ctx.JSON(http.StatusCreated, newResponseData("successfully"))
}
