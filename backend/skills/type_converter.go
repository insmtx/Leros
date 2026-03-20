package skills

import (
	"github.com/insmtx/SingerOS/backend/types"
)

// ConvertToDBModel 将Skill接口实例转换为可用于数据库存储的types.Skill模型
func ConvertToDBModel(skill Skill) *types.Skill {
	info := skill.Info()

	return &types.Skill{
		Code:         info.ID,
		Name:         info.Name,
		Description:  info.Description,
		Author:       info.Author,
		Version:      info.Version,
		Category:     info.Category,
		SkillType:    string(info.SkillType),
		Icon:         info.Icon,
		InputSchema:  interfaceMap(info.InputSchema),
		OutputSchema: interfaceMap(info.OutputSchema),
		Permissions:  permissionsToInterfaceSlice(info.Permissions),
		Config:       map[string]interface{}{},
		Status:       "active",
		IsSystem:     false,
	}
}

// ConvertFromDBModel 将数据库中的types.Skill模型转换为Skill引用信息
func ConvertFromDBModel(model *types.Skill) *SkillInfo {
	return &SkillInfo{
		ID:           model.Code,
		Name:         model.Name,
		Description:  model.Description,
		Author:       model.Author,
		Version:      model.Version,
		Category:     model.Category,
		SkillType:    SkillType(model.SkillType),
		Icon:         model.Icon,
		InputSchema:  convertToInputSchema(model.InputSchema),
		OutputSchema: convertToOutputSchema(model.OutputSchema),
		Permissions:  convertToPermissions(model.Permissions),
	}
}

// interfaceMap 将interface{}类型的map转换通用map
func interfaceMap(schema interface{}) map[string]interface{} {
	if m, ok := schema.(map[string]interface{}); ok {
		return m
	}
	return make(map[string]interface{})
}

// permissionsToInterfaceSlice 将Permission切片转换为interface{}切片
func permissionsToInterfaceSlice(perms []Permission) []interface{} {
	result := make([]interface{}, len(perms))
	for i, perm := range perms {
		result[i] = perm
	}
	return result
}

// convertToInputSchema 转换输入 schema
func convertToInputSchema(interf map[string]interface{}) InputSchema {
	if interf == nil {
		return InputSchema{}
	}

	schema := InputSchema{}
	if v, ok := interf["type"]; ok {
		if s, ok := v.(string); ok {
			schema.Type = s
		}
	}

	if v, ok := interf["required"]; ok {
		switch val := v.(type) {
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok {
					schema.Required = append(schema.Required, s)
				}
			}
		case []string:
			schema.Required = val
		}
	}

	if v, ok := interf["properties"]; ok {
		properties := make(map[string]*Property)
		if propMap, ok := v.(map[string]interface{}); ok {
			for k, v := range propMap {
				properties[k] = convertInterfaceToProperty(v)
			}
			schema.Properties = properties
		}
	}

	return schema
}

// convertToOutputSchema 转换输出 schema
func convertToOutputSchema(interf map[string]interface{}) OutputSchema {
	if interf == nil {
		return OutputSchema{}
	}

	schema := OutputSchema{}
	if v, ok := interf["type"]; ok {
		if s, ok := v.(string); ok {
			schema.Type = s
		}
	}

	if v, ok := interf["required"]; ok {
		switch val := v.(type) {
		case []interface{}:
			for _, item := range val {
				if s, ok := item.(string); ok {
					schema.Required = append(schema.Required, s)
				}
			}
		case []string:
			schema.Required = val
		}
	}

	if v, ok := interf["properties"]; ok {
		properties := make(map[string]*Property)
		if propMap, ok := v.(map[string]interface{}); ok {
			for k, v := range propMap {
				properties[k] = convertInterfaceToProperty(v)
			}
			schema.Properties = properties
		}
	}

	return schema
}

// convertInterfaceToProperty 将interface{}转为Property
func convertInterfaceToProperty(interf interface{}) *Property {
	prop := &Property{}

	if m, ok := interf.(map[string]interface{}); ok {
		if v, ok := m["type"]; ok {
			if s, ok := v.(string); ok {
				prop.Type = s
			}
		}

		if v, ok := m["title"]; ok {
			if s, ok := v.(string); ok {
				prop.Title = s
			}
		}

		if v, ok := m["description"]; ok {
			if s, ok := v.(string); ok {
				prop.Description = s
			}
		}

		if v, ok := m["default"]; ok {
			prop.Default = v
		}

		if v, ok := m["items"]; ok {
			prop.Items = convertInterfaceToProperty(v)
		}

		if v, ok := m["enum"]; ok {
			if enumList, ok := v.([]interface{}); ok {
				enumStrs := make([]string, len(enumList))
				for i, enumItem := range enumList {
					if s, ok := enumItem.(string); ok {
						enumStrs[i] = s
					}
				}
				prop.Enum = enumStrs
			}
		}
	}

	return prop
}

// convertToPermissions 转换权限切片
func convertToPermissions(interfs []interface{}) []Permission {
	perms := make([]Permission, 0)

	for _, interf := range interfs {
		if permMap, ok := interf.(map[string]interface{}); ok {
			var perm Permission

			if resource, exists := permMap["resource"]; exists {
				if resourceStr, ok := resource.(string); ok {
					perm.Resource = resourceStr
				}
			}

			if action, exists := permMap["action"]; exists {
				if actionStr, ok := action.(string); ok {
					perm.Action = actionStr
				}
			}

			perms = append(perms, perm)
		}
	}

	return perms
}
