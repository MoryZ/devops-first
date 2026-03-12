package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"devops-first/internal/database"
	"devops-first/internal/model"

	"gorm.io/gorm"
)

type GlobalVariableItem struct {
	ID          uint                `json:"id"`
	Key         string              `json:"key"`
	IsSecret    bool                `json:"is_secret"`
	Description string              `json:"description"`
	Fields      []GlobalVariableField `json:"fields"`
}

type GlobalVariableField struct {
	Name         string `json:"name"`
	Value        string `json:"value,omitempty"`
	IsSecret     bool   `json:"is_secret"`
	ValuePreview string `json:"value_preview"`
}

type storedField struct {
	Value    string `json:"value"`
	IsSecret bool   `json:"is_secret"`
}

type GlobalVariableService struct{}

func NewGlobalVariableService() *GlobalVariableService {
	return &GlobalVariableService{}
}

func (s *GlobalVariableService) List(userID uint) ([]GlobalVariableItem, error) {
	var rows []model.GlobalVariable
	if err := database.GetDB().Where("user_id = ?", userID).Order("`key` ASC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list global variables failed: %w", err)
	}

	items := make([]GlobalVariableItem, 0, len(rows))
	for _, row := range rows {
		fieldItems, hasSecret, err := parseFieldsForList(row.Value)
		if err != nil {
			return nil, err
		}
		items = append(items, GlobalVariableItem{
			ID:          row.ID,
			Key:         row.Key,
			IsSecret:    hasSecret,
			Description: row.Description,
			Fields:      fieldItems,
		})
	}
	return items, nil
}

func (s *GlobalVariableService) Upsert(userID uint, key string, fields []GlobalVariableField, description string) (*GlobalVariableItem, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, fmt.Errorf("key is required")
	}
	stored, hasSecret, err := normalizeFieldsForStore(fields)
	if err != nil {
		return nil, err
	}
	rawValue, _ := json.Marshal(stored)

	db := database.GetDB()
	var existing model.GlobalVariable
	err = db.Where("user_id = ? AND `key` = ?", userID, key).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("query global variable failed: %w", err)
	}

	if err == gorm.ErrRecordNotFound {
		entity := model.GlobalVariable{
			UserID:      userID,
			Key:         key,
			Value:       string(rawValue),
			IsSecret:    hasSecret,
			Description: description,
		}
		if err := db.Create(&entity).Error; err != nil {
			return nil, fmt.Errorf("create global variable failed: %w", err)
		}
		listFields, _, _ := parseFieldsForList(entity.Value)
		return &GlobalVariableItem{
			ID:          entity.ID,
			Key:         entity.Key,
			IsSecret:    entity.IsSecret,
			Description: entity.Description,
			Fields:      listFields,
		}, nil
	}

	existing.Value = string(rawValue)
	existing.IsSecret = hasSecret
	existing.Description = description
	if err := db.Save(&existing).Error; err != nil {
		return nil, fmt.Errorf("update global variable failed: %w", err)
	}
	listFields, _, _ := parseFieldsForList(existing.Value)

	return &GlobalVariableItem{
		ID:          existing.ID,
		Key:         existing.Key,
		IsSecret:    existing.IsSecret,
		Description: existing.Description,
		Fields:      listFields,
	}, nil
}

func (s *GlobalVariableService) Delete(userID uint, id uint) error {
	result := database.GetDB().Where("id = ? AND user_id = ?", id, userID).Delete(&model.GlobalVariable{})
	if result.Error != nil {
		return fmt.Errorf("delete global variable failed: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("global variable not found")
	}
	return nil
}

func (s *GlobalVariableService) ResolveValue(userID uint, key, field string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return "", nil
	}
	field = strings.TrimSpace(field)
	if field == "" {
		return "", nil
	}
	var row model.GlobalVariable
	if err := database.GetDB().Where("user_id = ? AND `key` = ?", userID, key).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("global variable %s not found", key)
		}
		return "", fmt.Errorf("resolve global variable failed: %w", err)
	}
	obj, err := parseObjectValue(row.Value)
	if err != nil {
		return "", err
	}
	v, ok := obj[field]
	if !ok {
		return "", fmt.Errorf("global variable %s.%s not found", key, field)
	}
	text := strings.TrimSpace(v.Value)
	if text == "<nil>" {
		return "", nil
	}
	return text, nil
}

func parseObjectValue(raw string) (map[string]storedField, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("value is required")
	}
	obj := map[string]storedField{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return nil, fmt.Errorf("value must be valid JSON object")
	}
	if len(obj) == 0 {
		return nil, fmt.Errorf("value JSON object cannot be empty")
	}
	return obj, nil
}

func parseFieldsForList(raw string) ([]GlobalVariableField, bool, error) {
	obj, err := parseObjectValue(raw)
	if err != nil {
		return nil, false, err
	}
	hasSecret := false
	fields := make([]GlobalVariableField, 0, len(obj))
	for k, item := range obj {
		name := strings.TrimSpace(k)
		if name == "" {
			continue
		}
		v := strings.TrimSpace(item.Value)
		preview := v
		if item.IsSecret {
			hasSecret = true
			preview = "******"
		}
		if !item.IsSecret && len(preview) > 24 {
			preview = preview[:24] + "..."
		}
		fields = append(fields, GlobalVariableField{
			Name:         name,
			Value:        "",
			IsSecret:     item.IsSecret,
			ValuePreview: preview,
		})
	}
	return fields, hasSecret, nil
}

func normalizeFieldsForStore(fields []GlobalVariableField) (map[string]storedField, bool, error) {
	if len(fields) == 0 {
		return nil, false, fmt.Errorf("fields are required")
	}
	out := make(map[string]storedField)
	hasSecret := false
	for _, f := range fields {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			return nil, false, fmt.Errorf("field name is required")
		}
		if _, exists := out[name]; exists {
			return nil, false, fmt.Errorf("duplicate field name: %s", name)
		}
		out[name] = storedField{
			Value:    strings.TrimSpace(f.Value),
			IsSecret: f.IsSecret,
		}
		if f.IsSecret {
			hasSecret = true
		}
	}
	if len(out) == 0 {
		return nil, false, fmt.Errorf("fields are required")
	}
	return out, hasSecret, nil
}
