package data

import (
	"context"
	"errors"
	"time"

	"LingChat/internal/data/ent/ent"
	"LingChat/internal/data/ent/ent/character"
)

var (
	ErrCharacterNotFound = errors.New("character not found")
	ErrCharacterExists   = errors.New("character already exists")
)

// CharacterRepo 定义角色的仓库接口
type CharacterRepo interface {
	// 创建角色
	Create(ctx context.Context, input *CharacterInput) (*ent.Character, error)
	// 根据character_id获取角色
	GetByCharacterID(ctx context.Context, characterID string) (*ent.Character, error)
	// 根据ID获取角色
	GetByID(ctx context.Context, id int64) (*ent.Character, error)
	// 列出所有角色
	List(ctx context.Context, offset, limit int) ([]*ent.Character, int, error)
	// 更新角色
	Update(ctx context.Context, id int64, input *CharacterInput) (*ent.Character, error)
	// 删除角色（软删除）
	Delete(ctx context.Context, id int64) error
	// 检查角色是否存在
	Exists(ctx context.Context, characterID string) (bool, error)
}

// CharacterInput 角色输入结构
type CharacterInput struct {
	CharacterID   string                 `json:"character_id"`
	Title         string                 `json:"title"`
	Info          string                 `json:"info"`
	ResourcePath  string                 `json:"resource_path"`
	DisplayParams map[string]interface{} `json:"display_params"`
	SystemPrompt  string                 `json:"system_prompt"`
}

// characterRepo 是实现 CharacterRepo 接口的仓库
type characterRepo struct {
	data *Data
}

// NewCharacterRepo 创建新的角色仓库
func NewCharacterRepo(data *Data) CharacterRepo {
	return &characterRepo{
		data: data,
	}
}

// Create 创建新角色
func (r *characterRepo) Create(ctx context.Context, input *CharacterInput) (*ent.Character, error) {
	// 检查character_id是否已存在
	exists, err := r.Exists(ctx, input.CharacterID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCharacterExists
	}

	create := r.data.db.Character.Create().
		SetCharacterID(input.CharacterID).
		SetTitle(input.Title).
		SetInfo(input.Info).
		SetResourcePath(input.ResourcePath).
		SetSystemPrompt(input.SystemPrompt)

	if input.DisplayParams != nil {
		create.SetDisplayParams(input.DisplayParams)
	}

	return create.Save(ctx)
}

// GetByCharacterID 根据character_id获取角色
func (r *characterRepo) GetByCharacterID(ctx context.Context, characterID string) (*ent.Character, error) {
	char, err := r.data.db.Character.Query().
		Where(character.CharacterID(characterID)).
		Where(character.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCharacterNotFound
		}
		return nil, err
	}
	return char, nil
}

// GetByID 根据ID获取角色
func (r *characterRepo) GetByID(ctx context.Context, id int64) (*ent.Character, error) {
	char, err := r.data.db.Character.Query().
		Where(character.ID(int(id))).
		Where(character.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCharacterNotFound
		}
		return nil, err
	}
	return char, nil
}

// List 列出所有角色
func (r *characterRepo) List(ctx context.Context, offset, limit int) ([]*ent.Character, int, error) {
	// 查询总数
	count, err := r.data.db.Character.Query().
		Where(character.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	chars, err := r.data.db.Character.Query().
		Where(character.DeletedAtIsNil()).
		Order(ent.Asc(character.FieldCreatedAt)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return chars, count, nil
}

// Update 更新角色
func (r *characterRepo) Update(ctx context.Context, id int64, input *CharacterInput) (*ent.Character, error) {
	// 检查角色是否存在
	char, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 如果要更新character_id，需要检查新的character_id是否已存在
	if input.CharacterID != "" && input.CharacterID != char.CharacterID {
		exists, err := r.Exists(ctx, input.CharacterID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCharacterExists
		}
	}

	update := r.data.db.Character.UpdateOneID(int(id))

	if input.CharacterID != "" {
		update.SetCharacterID(input.CharacterID)
	}
	if input.Title != "" {
		update.SetTitle(input.Title)
	}
	if input.Info != "" {
		update.SetInfo(input.Info)
	}
	if input.ResourcePath != "" {
		update.SetResourcePath(input.ResourcePath)
	}
	if input.DisplayParams != nil {
		update.SetDisplayParams(input.DisplayParams)
	}
	if input.SystemPrompt != "" {
		update.SetSystemPrompt(input.SystemPrompt)
	}

	return update.Save(ctx)
}

// Delete 删除角色（软删除）
func (r *characterRepo) Delete(ctx context.Context, id int64) error {
	// 检查角色是否存在
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return r.data.db.Character.UpdateOneID(int(id)).
		SetDeletedAt(time.Now()).
		Exec(ctx)
}

// Exists 检查角色是否存在
func (r *characterRepo) Exists(ctx context.Context, characterID string) (bool, error) {
	count, err := r.data.db.Character.Query().
		Where(character.CharacterID(characterID)).
		Where(character.DeletedAtIsNil()).
		Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
