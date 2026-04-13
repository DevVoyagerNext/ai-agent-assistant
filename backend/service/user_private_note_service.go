package service

import (
	"backend/dao"
	"backend/dto"
	"backend/model"
	"backend/pkg/utils"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type UserPrivateNoteService struct {
	privateNoteDao dao.UserPrivateNoteDao
}

// GetNoteOrChildren 获取笔记内容或子文件夹列表
func (s *UserPrivateNoteService) GetNoteOrChildren(ctx context.Context, userID uint, noteID int) (interface{}, error) {
	if userID == 0 {
		return nil, errors.New("用户未登录")
	}

	// 1. 如果 id 为 0，默认查询根目录
	if noteID == 0 {
		notes, err := s.privateNoteDao.GetNotesByParent(ctx, userID, 0)
		if err != nil {
			return nil, err
		}
		var children []dto.PrivateNoteItemRes
		for _, note := range notes {
			children = append(children, dto.PrivateNoteItemRes{
				ID:        note.ID,
				ParentID:  note.ParentID,
				Type:      note.Type,
				Title:     note.Title,
				UpdatedAt: note.UpdatedAt,
				CreatedAt: note.CreatedAt,
			})
		}
		return dto.PrivateNoteResponse{
			Type:     "folder",
			Children: children,
		}, nil
	}

	// 2. 查询该笔记
	note, err := s.privateNoteDao.GetNoteByID(ctx, userID, noteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("笔记不存在")
		}
		return nil, err
	}

	// 3. 如果是文件夹，获取其子节点
	if note.Type == "folder" {
		notes, err := s.privateNoteDao.GetNotesByParent(ctx, userID, noteID)
		if err != nil {
			return nil, err
		}
		var children []dto.PrivateNoteItemRes
		for _, n := range notes {
			children = append(children, dto.PrivateNoteItemRes{
				ID:        n.ID,
				ParentID:  n.ParentID,
				Type:      n.Type,
				Title:     n.Title,
				UpdatedAt: n.UpdatedAt,
				CreatedAt: n.CreatedAt,
			})
		}
		return dto.PrivateNoteResponse{
			Type:     "folder",
			Children: children,
		}, nil
	}

	// 4. 如果是文件，返回文件内容
	return dto.PrivateNoteResponse{
		Type: "markdown",
		Content: &dto.PrivateNoteDetailRes{
			ID:        note.ID,
			ParentID:  note.ParentID,
			Type:      note.Type,
			Title:     note.Title,
			Content:   note.Content,
			UpdatedAt: note.UpdatedAt,
			CreatedAt: note.CreatedAt,
		},
	}, nil
}

func (s *UserPrivateNoteService) GetNoteOrChildrenWithScope(ctx context.Context, userID uint, noteID int, scope int) (interface{}, error) {
	if userID == 0 {
		return nil, errors.New("用户未登录")
	}
	if scope != 0 && scope != 1 && scope != 2 {
		return nil, errors.New("scope 参数错误")
	}

	if noteID == 0 {
		notes, err := s.privateNoteDao.GetNotesByParentWithScope(ctx, userID, 0, scope)
		if err != nil {
			return nil, err
		}
		var children []dto.PrivateNoteItemRes
		for _, note := range notes {
			children = append(children, dto.PrivateNoteItemRes{
				ID:        note.ID,
				ParentID:  note.ParentID,
				Type:      note.Type,
				Title:     note.Title,
				UpdatedAt: note.UpdatedAt,
				CreatedAt: note.CreatedAt,
			})
		}
		return dto.PrivateNoteResponse{
			Type:     "folder",
			Children: children,
		}, nil
	}

	note, err := s.privateNoteDao.GetNoteByIDWithScope(ctx, userID, noteID, scope)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("笔记不存在")
		}
		return nil, err
	}

	if note.Type == "folder" {
		notes, err := s.privateNoteDao.GetNotesByParentWithScope(ctx, userID, noteID, scope)
		if err != nil {
			return nil, err
		}
		var children []dto.PrivateNoteItemRes
		for _, n := range notes {
			children = append(children, dto.PrivateNoteItemRes{
				ID:        n.ID,
				ParentID:  n.ParentID,
				Type:      n.Type,
				Title:     n.Title,
				UpdatedAt: n.UpdatedAt,
				CreatedAt: n.CreatedAt,
			})
		}
		return dto.PrivateNoteResponse{
			Type:     "folder",
			Children: children,
		}, nil
	}

	return dto.PrivateNoteResponse{
		Type: "markdown",
		Content: &dto.PrivateNoteDetailRes{
			ID:        note.ID,
			ParentID:  note.ParentID,
			Type:      note.Type,
			Title:     note.Title,
			Content:   note.Content,
			UpdatedAt: note.UpdatedAt,
			CreatedAt: note.CreatedAt,
		},
	}, nil
}

// CreatePrivateNote 创建私人笔记或文件夹
func (s *UserPrivateNoteService) CreatePrivateNote(ctx context.Context, userID uint, req dto.CreatePrivateNoteReq) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}

	// 1. 如果是文件，内容不能为空且长度不能超过 1000
	if req.Type == "markdown" {
		content := strings.TrimSpace(req.Content)
		if content == "" {
			return errors.New("文件内容不能为空")
		}
		if len([]rune(content)) > 1000 {
			return errors.New("文件内容不能超过 1000 个字符")
		}
		// 2. 防止 XSS 攻击
		req.Content = utils.XSSFilter(content)
	}

	// 3. 校验父文件夹是否存在且类型为 folder
	if req.ParentID != 0 {
		parent, err := s.privateNoteDao.GetNoteByID(ctx, userID, req.ParentID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("父文件夹不存在")
			}
			return err
		}
		if parent.Type != "folder" {
			return errors.New("无法在非文件夹节点下创建")
		}
	}

	// 4. 创建笔记
	note := &model.UserPrivateNote{
		UserID:   int(userID),
		ParentID: req.ParentID,
		Type:     req.Type,
		Title:    req.Title,
		Content:  req.Content,
		IsPublic: req.IsPublic,
	}

	return s.privateNoteDao.CreateNote(ctx, note)
}

func (s *UserPrivateNoteService) collectDescendantIDs(ctx context.Context, userID uint, parentID int, ids *[]int) error {
	children, err := s.privateNoteDao.GetNotesByParent(ctx, userID, parentID)
	if err != nil {
		return err
	}
	for _, child := range children {
		*ids = append(*ids, int(child.ID))
		if child.Type == "folder" {
			if err := s.collectDescendantIDs(ctx, userID, int(child.ID), ids); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeletePrivateNote 删除私人笔记或文件夹（文件夹递归删除所有子节点）
func (s *UserPrivateNoteService) DeletePrivateNote(ctx context.Context, userID uint, noteID int) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}
	if noteID <= 0 {
		return errors.New("笔记ID格式错误")
	}

	note, err := s.privateNoteDao.GetNoteByID(ctx, userID, noteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("笔记不存在")
		}
		return err
	}

	var ids []int
	ids = append(ids, noteID)
	if note.Type == "folder" {
		if err := s.collectDescendantIDs(ctx, userID, noteID, &ids); err != nil {
			return err
		}
	}

	return s.privateNoteDao.DeleteNotesByIDs(ctx, userID, ids)
}
