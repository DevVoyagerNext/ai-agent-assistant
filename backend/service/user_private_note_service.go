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
	// 调用带分页的版本，默认第一页，100条（目录通常不分页或分页很大）
	return s.GetNoteOrChildrenWithScope(ctx, userID, noteID, 2, 1, 100)
}

// GetNoteOrChildrenWithScope 获取笔记内容或子文件夹列表 (支持分页和 scope)
func (s *UserPrivateNoteService) GetNoteOrChildrenWithScope(ctx context.Context, userID uint, noteID int, scope int, page, pageSize int) (interface{}, error) {
	if userID == 0 {
		return nil, errors.New("用户未登录")
	}
	if scope != 0 && scope != 1 && scope != 2 {
		return nil, errors.New("scope 参数错误")
	}

	// 1. 如果 id 为 0，默认查询根目录
	if noteID == 0 {
		notes, total, err := s.privateNoteDao.GetNotesByParentWithScope(ctx, userID, 0, scope, page, pageSize)
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
				IsPublic:  note.IsPublic,
				UpdatedAt: note.UpdatedAt,
				CreatedAt: note.CreatedAt,
			})
		}
		return dto.PrivateNoteResponse{
			Type:     "folder",
			Total:    total,
			Children: children,
		}, nil
	}

	// 2. 查询该笔记
	note, err := s.privateNoteDao.GetNoteByIDWithScope(ctx, userID, noteID, scope)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("笔记不存在")
		}
		return nil, err
	}

	// 3. 如果是文件夹，获取其子节点
	if note.Type == "folder" {
		notes, total, err := s.privateNoteDao.GetNotesByParentWithScope(ctx, userID, noteID, scope, page, pageSize)
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
				IsPublic:  n.IsPublic,
				UpdatedAt: n.UpdatedAt,
				CreatedAt: n.CreatedAt,
			})
		}
		return dto.PrivateNoteResponse{
			Type:     "folder",
			Total:    total,
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
			IsPublic:  note.IsPublic,
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

	// 1. 如果是文件，内容长度不能超过 1000（允许为空）
	if req.Type == "markdown" {
		content := strings.TrimSpace(req.Content)
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

	// 4. 检查同一文件夹下是否存在同名同类型的文件/文件夹
	exists, err := s.privateNoteDao.CheckNoteExists(ctx, userID, req.ParentID, req.Title, req.Type)
	if err != nil {
		return err
	}
	if exists {
		if req.Type == "folder" {
			return errors.New("该文件夹下已存在同名的文件夹")
		}
		return errors.New("该文件夹下已存在同名的文件")
	}

	// 5. 创建笔记
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

// UpdatePrivateNoteContent 修改笔记内容 (仅限 markdown)
func (s *UserPrivateNoteService) UpdatePrivateNoteContent(ctx context.Context, userID uint, noteID int, req dto.UpdatePrivateNoteContentReq) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}

	// 1. 获取笔记并校验
	note, err := s.privateNoteDao.GetNoteByID(ctx, userID, noteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("笔记不存在")
		}
		return err
	}

	// 2. 只有 markdown 文件可以修改内容
	if note.Type != "markdown" {
		return errors.New("只有文件类型可以修改内容")
	}

	// 3. 校验内容（允许为空）
	content := strings.TrimSpace(req.Content)
	if len([]rune(content)) > 1000 {
		return errors.New("内容不能超过 1000 字符")
	}

	// 4. XSS 过滤
	safeContent := utils.XSSFilter(content)

	return s.privateNoteDao.UpdateNote(ctx, userID, noteID, map[string]interface{}{
		"content": safeContent,
	})
}

// UpdatePrivateNoteTitle 修改文件或文件夹标题
func (s *UserPrivateNoteService) UpdatePrivateNoteTitle(ctx context.Context, userID uint, noteID int, req dto.UpdatePrivateNoteTitleReq) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return errors.New("标题不能为空")
	}
	if len([]rune(title)) > 255 {
		return errors.New("标题不能超过 255 字符")
	}

	// 1. 先获取当前笔记，得知它的 type 和 parentId
	note, err := s.privateNoteDao.GetNoteByID(ctx, userID, noteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("要修改的笔记或文件夹不存在")
		}
		return err
	}

	// 2. 如果标题没变，直接返回成功
	if note.Title == title {
		return nil
	}

	// 3. 检查同一文件夹下是否已存在同名同类型的文件/文件夹
	exists, err := s.privateNoteDao.CheckNoteExists(ctx, userID, note.ParentID, title, note.Type)
	if err != nil {
		return err
	}
	if exists {
		if note.Type == "folder" {
			return errors.New("该文件夹下已存在同名的文件夹")
		}
		return errors.New("该文件夹下已存在同名的文件")
	}

	// 4. 更新标题
	return s.privateNoteDao.UpdateNote(ctx, userID, noteID, map[string]interface{}{
		"title": title,
	})
}

// UpdatePrivateNotePublic 修改文件或文件夹公开状态
func (s *UserPrivateNoteService) UpdatePrivateNotePublic(ctx context.Context, userID uint, noteID int, req dto.UpdatePrivateNotePublicReq) error {
	if userID == 0 {
		return errors.New("用户未登录")
	}

	return s.privateNoteDao.UpdateNote(ctx, userID, noteID, map[string]interface{}{
		"is_public": req.IsPublic,
	})
}

func (s *UserPrivateNoteService) collectDescendantIDs(ctx context.Context, userID uint, parentID int, ids *[]int) error {
	// 递归时默认获取全部(scope=2)，且不分页(给个很大的数)
	notes, _, err := s.privateNoteDao.GetNotesByParentWithScope(ctx, userID, parentID, 2, 1, 10000)
	if err != nil {
		return err
	}
	for _, child := range notes {
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
