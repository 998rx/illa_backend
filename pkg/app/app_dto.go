package app

import (
	"time"

	"github.com/google/uuid"
	"github.com/illacloud/builder-backend/internal/repository"
)

const APP_EDITED_BY_LEN = 4

type AppDto struct {
	ID              int                       `json:"appId"` // generated by database primary key serial
	UID             uuid.UUID                 `json:"uid"`
	TeamID          int                       `json:"teamID"`
	Name            string                    `json:"appName" validate:"required"`
	ReleaseVersion  int                       `json:"releaseVersion"`  // release version used for mark the app release version.
	MainlineVersion int                       `json:"mainlineVersion"` // mainline version keep the newest app version in database.
	Config          *repository.AppConfig     `json:"config"`
	CreatedBy       int                       `json:"-" `
	CreatedAt       time.Time                 `json:"-"`
	UpdatedBy       int                       `json:"updatedBy"`
	UpdatedAt       time.Time                 `json:"updatedAt"`
	AppActivity     AppActivity               `json:"appActivity"`
	EditedBy        []*repository.AppEditedBy `json:"editedBy"`
}

func NewAppDto() *AppDto {
	return &AppDto{}
}

func (a *AppDto) InitUID() {
	a.UID = uuid.New()
}

func (a *AppDto) InitConfig() {
	a.Config = repository.NewAppConfigByDefault()
}

func (a *AppDto) InitUpdatedAt() {
	a.UpdatedAt = time.Now().UTC()
}

func (a *AppDto) UpdateAppDTOConfig(appConfig *repository.AppConfig, userID int) {
	a.Config = appConfig
	a.UpdatedBy = userID
	a.InitUpdatedAt()
}

func (a *AppDto) ExportAppDtoConfig() *repository.AppConfig {
	return a.Config
}

func (a *AppDto) ExportEditedByUserIDs() []int {
	ids := make([]int, len(a.EditedBy))
	for _, appEditedBy := range a.EditedBy {
		ids = append(ids, appEditedBy.UserID)
	}
	return ids
}

func (a *AppDto) SetTeamID(teamID int) {
	a.TeamID = teamID
}

func (appd *AppDto) ConstructByMap(data interface{}) {

	udata, ok := data.(map[string]interface{})
	if !ok {
		return
	}
	for k, v := range udata {
		switch k {
		case "id":
			idf, _ := v.(float64)
			appd.ID = int(idf)
		case "name":
			appd.Name, _ = v.(string)
		}
	}
}

func (appd *AppDto) ConstructWithID(id int) {
	appd.ID = id
}

func (appd *AppDto) ConstructWithUpdateBy(updateBy int) {
	appd.UpdatedBy = updateBy
}

func (appd *AppDto) AddEditedBy(pendingEditedBy *repository.AppEditedBy) {
	// check if edited by target already in payload, then remove it
	for serial, appEditBy := range appd.EditedBy {
		if appEditBy.UserID == pendingEditedBy.UserID {
			appd.EditedBy = append(appd.EditedBy[:serial], appd.EditedBy[serial+1:]...)
		}
	}
	// insert edited by
	appd.EditedBy = append(appd.EditedBy, pendingEditedBy)
	if len(appd.EditedBy) > APP_EDITED_BY_LEN {
		appd.EditedBy = appd.EditedBy[:len(appd.EditedBy)-1]
	}
}
