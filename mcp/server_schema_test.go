package mcp

import "testing"

func TestListToolsHaveCompleteSchemas(t *testing.T) {
	var s Server
	tools := s.ListTools()
	for _, ti := range tools {
		if ti.InputSchema == nil {
			t.Fatalf("tool %s has nil InputSchema", ti.Name)
		}
		if typ, ok := ti.InputSchema["type"]; !ok || typ != "object" {
			t.Errorf("tool %s schema missing or wrong type: got %v", ti.Name, typ)
		}
		if props, ok := ti.InputSchema["properties"]; !ok {
			t.Errorf("tool %s schema missing properties", ti.Name)
		} else {
			if pm, ok := props.(map[string]interface{}); ok {
				if len(pm) == 0 {
					if _, ok := ti.InputSchema["additionalProperties"]; !ok {
						t.Errorf("tool %s has empty properties and missing additionalProperties", ti.Name)
					}
				}
			}
		}
		if _, ok := ti.InputSchema["required"]; !ok {
			t.Errorf("tool %s schema missing required", ti.Name)
		}
	}
}
