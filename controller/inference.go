package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/xihe-server/app"
	"github.com/opensourceways/xihe-server/domain"
	"github.com/opensourceways/xihe-server/domain/message"
	"github.com/opensourceways/xihe-server/domain/platform"
	"github.com/opensourceways/xihe-server/domain/repository"
	userapp "github.com/opensourceways/xihe-server/user/app"
	"github.com/opensourceways/xihe-server/utils"
)

func AddRouterForInferenceController(
	rg *gin.RouterGroup,
	p platform.RepoFile,
	repo repository.Inference,
	project repository.Project,
	sender message.Sender,
	whitelist userapp.WhiteListService,
) {
	ctl := InferenceController{
		s: app.NewInferenceService(
			p, repo, sender, apiConfig.MinSurvivalTimeOfInference,
		),
		project:   project,
		whitelist: whitelist,
	}

	ctl.inferenceDir, _ = domain.NewDirectory(apiConfig.InferenceDir)
	ctl.inferenceBootFile, _ = domain.NewFilePath(apiConfig.InferenceBootFile)

	rg.GET("/v1/inference/project/:owner/:pid", ctl.Create)
}

type InferenceController struct {
	baseController

	s app.InferenceService

	project repository.Project

	inferenceDir      domain.Directory
	inferenceBootFile domain.FilePath
	whitelist         userapp.WhiteListService
}

// @Summary		Create
// @Description	create inference
// @Tags			Inference
// @Param			owner	path	string	true	"project owner"
// @Param			pid		path	string	true	"project id"
// @Accept			json
// @Success		201	{object}			app.InferenceDTO
// @Failure		400	bad_request_body	can't	parse		request	body
// @Failure		401	bad_request_param	some	parameter	of		body	is	invalid
// @Failure		500	system_error		system	error
// @Router			/v1/inference/project/{owner}/{pid} [get]
func (ctl *InferenceController) Create(ctx *gin.Context) {
	pl, csrftoken, _, ok := ctl.checkTokenForWebsocket(ctx, false)
	if !ok {
		return
	}

	// setup websocket
	upgrader := websocket.Upgrader{
		Subprotocols: []string{csrftoken},
		CheckOrigin: func(r *http.Request) bool {
			return r.Header.Get(headerSecWebsocket) == csrftoken
		},
	}

	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		//TODO delete
		log.Errorf("update ws failed, err:%s", err.Error())

		ctl.sendRespWithInternalError(ctx, newResponseError(err))

		return
	}

	defer ws.Close()

	// start
	owner, err := domain.NewAccount(ctx.Param("owner"))
	if err != nil {
		if wsErr := ws.WriteJSON(newResponseCodeError(errorBadRequestParam, err)); wsErr != nil {
			log.Errorf("inference failed: web socket write err:%s", wsErr.Error())
		}

		log.Errorf("inference failed: new account, err:%s", err.Error())

		return
	}

	projectId := ctx.Param("pid")
	v, err := ctl.project.GetSummary(owner, projectId)
	if err != nil {
		if wsErr := ws.WriteJSON(newResponseError(err)); wsErr != nil {
			log.Errorf("inference get | web socket write err:%s", wsErr.Error())
		}

		log.Errorf("inference failed: get summary, err:%s", err.Error())

		return
	}

	viewOther := pl.isNotMe(owner)

	if v.IsPrivate() {
		wsErr := ws.WriteJSON(
			newResponseCodeMsg(
				errorNotAllowed,
				"project is not found",
			),
		)
		if wsErr != nil {
			log.Errorf("inference get | web socket write err:%s", wsErr.Error())
		}

		log.Debug("inference failed: project is private")

		return
	}

	var level string
	if level, err = ctl.getResourceLevel(owner, projectId); err != nil {
		if wsErr := ws.WriteJSON(newResponseError(err)); wsErr != nil {
			log.Errorf("inference get | web socket write err:%s", wsErr.Error())
		}

		log.Errorf("inference failed: get resource, err:%s", err.Error())

		return
	}

	u := platform.UserInfo{}
	if viewOther {
		u.User = owner
	} else {
		u = pl.PlatformUserInfo()
	}

	cmd := app.InferenceCreateCmd{
		ProjectId:     v.Id,
		ProjectName:   v.Name,
		ProjectOwner:  owner,
		ResourceLevel: level,
		InferenceDir:  ctl.inferenceDir,
		BootFile:      ctl.inferenceBootFile,
	}

	dto, lastCommit, err := ctl.s.Create(pl.Account, &u, &cmd)
	if err != nil {
		if wsErr := ws.WriteJSON(newResponseError(err)); wsErr != nil {
			log.Errorf("inference get | web socket write err:%s", wsErr.Error())
		}

		log.Errorf("inference failed: create, err:%s", err.Error())

		return
	}

	utils.DoLog("", pl.Account, "create gradio",
		fmt.Sprintf("projectid: %s", v.Id), "success")

	if dto.Error != "" || dto.AccessURL != "" {
		if wsErr := ws.WriteJSON(newResponseData(dto)); wsErr != nil {
			log.Errorf("inference get | web socket write err:%s", wsErr.Error())
		}

		return
	}

	time.Sleep(10 * time.Second)

	info := app.InferenceIndex{
		Id:         dto.InstanceId,
		LastCommit: lastCommit,
	}
	info.Project.Id = projectId
	info.Project.Owner = owner

	for i := 0; i < apiConfig.InferenceTimeout; i++ {
		dto, err = ctl.s.Get(&info)
		if err != nil {
			if wsErr := ws.WriteJSON(newResponseError(err)); wsErr != nil {
				log.Errorf("inference create | web socket write err:%s", wsErr.Error())
			}

			log.Errorf("inference failed: get status, err:%s", err.Error())

			return
		}

		log.Debugf("info dto:%v", dto)

		if dto.Error != "" || dto.AccessURL != "" {
			if wsErr := ws.WriteJSON(newResponseData(dto)); wsErr != nil {
				log.Errorf("inference create | web socket write err:%s", wsErr.Error())
			}

			log.Debug("inference done")

			return
		}

		time.Sleep(time.Second)
	}

	log.Error("inference timeout")

	if wsErr := ws.WriteJSON(newResponseCodeMsg(errorSystemError, "timeout")); wsErr != nil {
		logrus.Errorf("inference | web socket write error: %s", wsErr.Error())
	}
}

func (ctl *InferenceController) getResourceLevel(owner domain.Account, pid string) (level string, err error) {
	resources, err := ctl.project.FindUserProjects(
		[]repository.UserResourceListOption{
			{
				Owner: owner,
				Ids: []string{
					pid,
				},
			},
		},
	)

	if err != nil || len(resources) < 1 {
		return
	}

	if resources[0].Level != nil {
		level = resources[0].Level.ResourceLevel()
	}

	return
}
