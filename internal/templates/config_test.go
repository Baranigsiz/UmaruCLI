package templates

import (
	"testing"
)

func TestGetAvailableTemplates(t *testing.T) {
	templates, err := GetAvailableTemplates()
	if err != nil {
		t.Fatalf("GetAvailableTemplates() failed: %v", err)
	}

	if len(templates) == 0 {
		t.Fatalf("Expected at least 1 template, got 0")
	}

	for _, tmpl := range templates {
		if tmpl.ID == "" {
			t.Errorf("Template has empty ID: %+v", tmpl)
		}
		if tmpl.Name == "" {
			t.Errorf("Template %s has empty Name", tmpl.ID)
		}
		if tmpl.Description == "" {
			t.Errorf("Template %s has empty Description", tmpl.ID)
		}
	}
}

func TestFindTemplateByID(t *testing.T) {
	// 1. Existing template
	tmpl, err := FindTemplateByID("go-fiber")
	if err != nil {
		t.Fatalf("Expected to find 'go-fiber', got error: %v", err)
	}
	if tmpl.ID != "go-fiber" {
		t.Errorf("Expected ID 'go-fiber', got '%s'", tmpl.ID)
	}

	// 2. Non-existent template
	_, err = FindTemplateByID("non-existent-template-xyz")
	if err == nil {
		t.Errorf("Expected error for non-existent template, got nil")
	}
}

