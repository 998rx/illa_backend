// Copyright 2022 The ILLA Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package state

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/illa-family/builder-backend/internal/repository"
	"github.com/illa-family/builder-backend/pkg/app"
	"go.uber.org/zap"
)

type TreeStateService interface {
	CreateTreeState(versionId int, treestate TreeStateDto) (TreeStateDto, error)
	DeleteTreeState(treestateId int) error
	UpdateTreeState(versionId int, treestate TreeStateDto) (TreeStateDto, error)
	GetTreeStateByID(treestateID int) (TreeStateDto, error)
	GetAllTypeTreeStateByApp(app *app.AppDto, version int) ([]*TreeStateDto, error)
	GetTreeStateByApp(app *app.AppDto, statetype int, version int) ([]*TreeStateDto, error)
	ReleaseTreeStateByApp(app *app.AppDto) error
}

type TreeStateDto struct {
	ID                 int       `json:"id"`
	StateType          int       `json:"state_type"`
	ParentNodeRefID    int       `json:"parent_node_ref_id"`
	ChildrenNodeRefIDs []int     `json:"children_node_ref_ids"`
	AppRefID           int       `json:"app_ref_id"`
	Version            int       `json:"version"`
	Name               string    `json:"name"`
	Content            string    `json:"content"`
	CreatedAt          time.Time `json:"created_at"`
	CreatedBy          int       `json:"created_by"`
	UpdatedAt          time.Time `json:"updated_at"`
	UpdatedBy          int       `json:"updated_by"`
}

type TreeStateServiceImpl struct {
	logger              *zap.SugaredLogger
	treestateRepository repository.TreeStateRepository
}

func NewTreeStateServiceImpl(logger *zap.SugaredLogger, treestateRepository repository.TreeStateRepository) *TreeStateServiceImpl {
	return &TreeStateServiceImpl{
		logger:              logger,
		treestateRepository: treestateRepository,
	}
}

func (impl *TreeStateServiceImpl) NewTreeStateByComponentState(apprefid int, cnode *repository.ComponentNode) (*TreeStateDto, error) {
	var cnodeserilized []byte
	var err error
	if cnodeserilized, err = cnode.SerializationForDatabase(); err != nil {
		return nil, err
	}

	treestatedto := &TreeStateDto{
		StateType: repository.TREE_STATE_TYPE_COMPONENTS,
		AppRefID:  apprefid,
		Version:   repository.APP_EDIT_VERSION,
		Name:      cnode.DisplayName,
		Content:   string(cnodeserilized),
	}
	return treestatedto, nil
}

func (impl *TreeStateServiceImpl) CreateTreeState(treestate TreeStateDto) (TreeStateDto, error) {
	// TODO: validate the versionId
	validate := validator.New()
	if err := validate.Struct(treestate); err != nil {
		return TreeStateDto{}, err
	}
	treestate.CreatedAt = time.Now().UTC()
	treestate.UpdatedAt = time.Now().UTC()
	treestateIDsJSON, err := json.Marshal(treestate.ChildrenNodeRefIDs)
	if err != nil {
		return TreeStateDto{}, err
	}
	treeStateForStorage := repository.TreeState{
		StateType:          treestate.StateType,
		ParentNodeRefID:    treestate.ParentNodeRefID,
		ChildrenNodeRefIDs: string(treestateIDsJSON),
		AppRefID:           treestate.AppRefID,
		Version:            treestate.Version,
		Name:               treestate.Name,
		Content:            treestate.Content,
		CreatedAt:          treestate.CreatedAt,
		CreatedBy:          treestate.CreatedBy,
		UpdatedAt:          treestate.UpdatedAt,
		UpdatedBy:          treestate.UpdatedBy,
	}
	if err := impl.treestateRepository.Create(&treeStateForStorage); err != nil {
		return TreeStateDto{}, err
	}
	// fill created id
	fmt.Printf("[DUMP indb id] treeStateForStorage.ID %d\n", treeStateForStorage.ID)
	treestate.ID = treeStateForStorage.ID
	return treestate, nil
}

func (impl *TreeStateServiceImpl) DeleteTreeState(treestateID int) error {
	if err := impl.treestateRepository.Delete(treestateID); err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateServiceImpl) UpdateTreeState(treestate TreeStateDto) (TreeStateDto, error) {
	validate := validator.New()
	if err := validate.Struct(treestate); err != nil {
		return TreeStateDto{}, err
	}
	treestate.UpdatedAt = time.Now().UTC()
	treestateIDsJSON, err := json.Marshal(treestate.ChildrenNodeRefIDs)
	if err != nil {
		return TreeStateDto{}, err
	}
	if err := impl.treestateRepository.Update(&repository.TreeState{
		ID:                 treestate.ID,
		StateType:          treestate.StateType,
		ParentNodeRefID:    treestate.ParentNodeRefID,
		ChildrenNodeRefIDs: string(treestateIDsJSON),
		AppRefID:           treestate.AppRefID,
		Version:            treestate.Version,
		Name:               treestate.Name,
		Content:            treestate.Content,
		UpdatedAt:          treestate.UpdatedAt,
		UpdatedBy:          treestate.UpdatedBy,
	}); err != nil {
		return TreeStateDto{}, err
	}
	return treestate, nil
}

func (impl *TreeStateServiceImpl) GetTreeStateByID(treestateID int) (TreeStateDto, error) {
	res, err := impl.treestateRepository.RetrieveByID(treestateID)
	if err != nil {
		return TreeStateDto{}, err
	}
	ids, err := res.ExportChildrenNodeRefIDs()
	if err != nil {
		return TreeStateDto{}, err
	}
	resDto := TreeStateDto{
		ID:                 res.ID,
		StateType:          res.StateType,
		ParentNodeRefID:    res.ParentNodeRefID,
		ChildrenNodeRefIDs: ids,
		AppRefID:           res.AppRefID,
		Version:            res.Version,
		Name:               res.Name,
		Content:            res.Content,
		CreatedAt:          res.CreatedAt,
		CreatedBy:          res.CreatedBy,
		UpdatedAt:          res.UpdatedAt,
		UpdatedBy:          res.UpdatedBy,
	}
	return resDto, nil
}

func (impl *TreeStateServiceImpl) GetAllTypeTreeStateByApp(app *app.AppDto, version int) ([]*TreeStateDto, error) {
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(app.ID, version)
	if err != nil {
		return nil, err
	}
	treestatesdto := make([]*TreeStateDto, len(treestates))
	for _, treestate := range treestates {
		ids, err := treestate.ExportChildrenNodeRefIDs()
		if err != nil {
			return nil, err
		}
		treestatesdto = append(treestatesdto, &TreeStateDto{
			ID:                 treestate.ID,
			StateType:          treestate.StateType,
			ParentNodeRefID:    treestate.ParentNodeRefID,
			ChildrenNodeRefIDs: ids,
			AppRefID:           treestate.AppRefID,
			Version:            treestate.Version,
			Name:               treestate.Name,
			Content:            treestate.Content,
			CreatedAt:          treestate.CreatedAt,
			CreatedBy:          treestate.CreatedBy,
			UpdatedAt:          treestate.UpdatedAt,
			UpdatedBy:          treestate.UpdatedBy,
		})
	}
	return treestatesdto, nil
}

func (impl *TreeStateServiceImpl) GetTreeStateByApp(app *app.AppDto, statetype int, version int) ([]*TreeStateDto, error) {
	treestates, err := impl.treestateRepository.RetrieveTreeStatesByApp(app.ID, statetype, version)
	if err != nil {
		return nil, err
	}
	treestatesdto := make([]*TreeStateDto, len(treestates))
	for _, treestate := range treestates {
		ids, err := treestate.ExportChildrenNodeRefIDs()
		if err != nil {
			return nil, err
		}
		treestatesdto = append(treestatesdto, &TreeStateDto{
			ID:                 treestate.ID,
			StateType:          treestate.StateType,
			ParentNodeRefID:    treestate.ParentNodeRefID,
			ChildrenNodeRefIDs: ids,
			AppRefID:           treestate.AppRefID,
			Version:            treestate.Version,
			Name:               treestate.Name,
			Content:            treestate.Content,
			CreatedAt:          treestate.CreatedAt,
			CreatedBy:          treestate.CreatedBy,
			UpdatedAt:          treestate.UpdatedAt,
			UpdatedBy:          treestate.UpdatedBy,
		})
	}
	return treestatesdto, nil
}

// @todo: should this method be in a transaction?
func (impl *TreeStateServiceImpl) ReleaseTreeStateByApp(app *app.AppDto) error {
	// get edit version K-V state from database
	treestates, err := impl.treestateRepository.RetrieveAllTypeTreeStatesByApp(app.ID, repository.APP_EDIT_VERSION)
	if err != nil {
		return err
	}
	// set version as minaline version
	for serial, _ := range treestates {
		treestates[serial].Version = app.MainlineVersion
	}
	// and put them to the database as duplicate
	for _, treestate := range treestates {
		if err := impl.treestateRepository.Create(treestate); err != nil {
			return err
		}
	}
	return nil
}

func (impl *TreeStateServiceImpl) MoveTreeStateNode(apprefid int, nowNode *TreeStateDto) error {
	// prepare data
	oldParentTreeState := &repository.TreeState{}
	newParentTreeState := &repository.TreeState{}
	nowTreeState := &repository.TreeState{}
	var err error
	// get nowTreeState by name
	if nowTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(apprefid, nowNode.StateType, nowNode.Name); err != nil {
		return err
	}
	// get oldParentTreeState by id
	if oldParentTreeState, err = impl.treestateRepository.RetrieveByID(nowTreeState.ParentNodeRefID); err != nil {
		return err
	}
	// get newParentTreeState by name
	switch nowNode.StateType {
	case repository.TREE_STATE_TYPE_COMPONENTS:
		componentNode := &repository.ComponentNode{}
		var err error
		if componentNode, err = repository.NewComponentNodeFromJSON([]byte(nowTreeState.Content)); err != nil {
			return err
		}
		if newParentTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(apprefid, nowNode.StateType, componentNode.ParentNode); err != nil {
			return err
		}
	case repository.TREE_STATE_TYPE_DEPENDENCIES:
		// @todo: finish this method
		return err

	case repository.TREE_STATE_TYPE_EXECUTION:
		return err
	}
	// fill into database
	// update nowTreeState
	if err := impl.treestateRepository.Update(nowTreeState); err != nil {
		return err
	}

	// add now TreeState id into new parent TreeState.ChildrenNodeRefIDs
	newParentTreeState.AppendChildrenNodeRefIDs(nowTreeState.ID)

	// update newParentTreeState
	if err := impl.treestateRepository.Update(newParentTreeState); err != nil {
		return err
	}

	// remove now TreeState id from old parent TreeState.ChildrenNodeRefIDs
	oldParentTreeState.RemoveChildrenNodeRefIDs(nowTreeState.ID)

	// update oldParentTreeState
	if err := impl.treestateRepository.Update(oldParentTreeState); err != nil {
		return err
	}
	return nil
}

func (impl *TreeStateServiceImpl) DeleteTreeStateNodeRecursive(apprefid int, nowNode *TreeStateDto) error {
	// get nowTreeState by displayName from database
	var err error
	nowTreeState := &repository.TreeState{}
	if nowTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(apprefid, nowNode.StateType, nowNode.Name); err != nil {
		return err
	}
	// get all sub nodes recursive
	targetNodes := []*repository.TreeState{}
	if targetNodes, err = impl.retrieveChildrenNodes(nowTreeState); err != nil {
		return err
	}
	// do not forget delete now node
	targetNodes = append(targetNodes, nowTreeState)
	// delete them all
	// @todo: replace this Delete to a batch method
	for _, node := range targetNodes {
		if err = impl.treestateRepository.Delete(node.ID); err != nil {
			return err
		}
	}
	return nil
}

func (impl *TreeStateServiceImpl) retrieveChildrenNodes(treeState *repository.TreeState) ([]*repository.TreeState, error) {
	childrenNodes := []*repository.TreeState{}
	// @todo: replace this RetrieveByID to a batch method
	ids, err := treeState.ExportChildrenNodeRefIDs()
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		var node *repository.TreeState
		var err error
		if node, err = impl.treestateRepository.RetrieveByID(id); err != nil {
			return nil, err
		}
		var subChildrenNodes []*repository.TreeState
		if subChildrenNodes, err = impl.retrieveChildrenNodes(node); err != nil {
			return nil, err
		}
		childrenNodes = append(childrenNodes, subChildrenNodes...)
	}
	return childrenNodes, nil
}

func (impl *TreeStateServiceImpl) CreateComponentTree(apprefid int, parentNodeID int, componentNodeTree *repository.ComponentNode) error {
	fmt.Printf("[DUMP] parentNodeID: %d\n", parentNodeID)
	fmt.Printf("[DUMP] componentNodeTree: %v\n", componentNodeTree)
	fmt.Printf("[DUMP] componentNodeTree.ChildrenNode: %v\n", componentNodeTree.ChildrenNode)
	// convert ComponentNode to TreeState
	nowNode := &TreeStateDto{}
	var err error
	if nowNode, err = impl.NewTreeStateByComponentState(apprefid, componentNodeTree); err != nil {
		return err
	}
	fmt.Printf("\n[D-0] %v\n", componentNodeTree.ChildrenNode)
	// get parentNode
	parentTreeState := &repository.TreeState{}
	isSummitNode := true
	if parentNodeID != 0 { // parentNode is in database
		isSummitNode = false
		if parentTreeState, err = impl.treestateRepository.RetrieveByID(parentNodeID); err != nil {
			return err
		}
	} else if componentNodeTree.ParentNode != "" { // or parentNode is exist
		isSummitNode = false
		if parentTreeState, err = impl.treestateRepository.RetrieveEditVersionByAppAndName(nowNode.AppRefID, nowNode.StateType, componentNodeTree.ParentNode); err != nil {
			return err
		}
	}

	// no parentNode, nowNode is tree summit
	if isSummitNode {
		nowNode.ParentNodeRefID = repository.TREE_STATE_SUMMIT_ID
	} else {
		// fill nowNode.ParentNodeRefID
		nowNode.ParentNodeRefID = parentTreeState.ID
	}
	// insert nowNode and get id
	fmt.Printf("\n[D-1] %v\n", componentNodeTree.ChildrenNode)
	var treeStateDtoInDB TreeStateDto
	if treeStateDtoInDB, err = impl.CreateTreeState(*nowNode); err != nil {
		return err
	}
	nowNode.ID = treeStateDtoInDB.ID

	// fill nowNode id into parentNode.ChildrenNodeRefIDs
	if !isSummitNode {
		fmt.Printf("[fill nowNode id into parentNode.ChildrenNodeRefIDs]\n")
		parentTreeState.AppendChildrenNodeRefIDs(nowNode.ID)
		fmt.Printf("[DUMP] nowNode.ID: %d\n", nowNode.ID)
		fmt.Printf("[DUMP] parentTreeState.ChildrenNodeRefIDs: %v\n", parentTreeState.ChildrenNodeRefIDs)
		// save parentNode
		if err = impl.treestateRepository.Update(parentTreeState); err != nil {
			return err
		}
	}
	// create nowNode.ChildrenNode
	fmt.Printf("\n[D2] %v\n", componentNodeTree.ChildrenNode)
	for _, childrenComponentNode := range componentNodeTree.ChildrenNode {
		if err := impl.CreateComponentTree(apprefid, nowNode.ID, childrenComponentNode); err != nil {
			return err
		}
	}
	return nil
}
