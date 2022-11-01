package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/opensourceways/xihe-inference-evaluate/app"
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
)

func AddRouterForInferenceController(
	rg *gin.RouterGroup,
	manager inference.Inference,
) {
	ctl := InferenceController{
		s: app.NewInferenceService(manager),
	}

	rg.POST("/v1/inference/project", ctl.Create)
	rg.PUT("/v1/inference/project", ctl.ExtendExpiry)
}

type InferenceController struct {
	baseController

	s app.InferenceService
}

// @Summary Create
// @Description create inference
// @Tags  Inference
// @Accept json
// @Success 201
// @Failure 400 bad_request_body    can't parse request body
// @Failure 401 bad_request_param   some parameter of body is invalid
// @Failure 500 system_error        system error
// @Router /v1/inference/project [post]
func (ctl *InferenceController) Create(ctx *gin.Context) {
	req := InferenceCreateRequest{}
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

	if err = ctl.s.Create(&cmd); err != nil {
		ctl.sendRespWithInternalError(ctx, newResponseError(err))

		return
	}

	ctx.JSON(http.StatusCreated, newResponseData("successfully"))
}

// @Summary ExtendExpiry
// @Description extend expiry for inference
// @Tags  Inference
// @Accept json
// @Success 202
// @Failure 400 bad_request_body    can't parse request body
// @Failure 401 bad_request_param   some parameter of body is invalid
// @Failure 500 system_error        system error
// @Router /v1/inference/project [put]
func (ctl *InferenceController) ExtendExpiry(ctx *gin.Context) {
	req := InferenceUpdateRequest{}
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
