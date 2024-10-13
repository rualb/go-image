package i18n

import (
	"go-image/internal/config"
	"testing"
)

// Mock function to create AppConfig for testing
func createMockAppConfig(langs []string) *config.AppConfig {
	return &config.AppConfig{
		Lang: config.AppConfigLang{
			Langs: langs,
		},
	}
}

// Test the creation of AppLang and language loading
func TestNewAppLang(t *testing.T) {
	// Create a mock configuration for testing
	appConfig := createMockAppConfig([]string{"en", "es"})

	// Initialize the AppLang
	appLang := MustNewAppLang(appConfig)

	if appLang == nil {
		t.Fatal("Expected a valid AppLang instance, got nil")
	}

	// Test if the default language is set correctly
	userLang := appLang.UserLang("en")
	if userLang.LangCode() != "en" {
		t.Errorf("Expected default language to be 'en', got '%s'", userLang.LangCode())
	}

	// Test fallback to default when unsupported language is requested
	unsupportedLang := appLang.UserLang("fr")
	if unsupportedLang.LangCode() != "en" {
		t.Errorf("Expected fallback to 'en', got '%s'", unsupportedLang.LangCode())
	}
}

// Test Lang method for placeholder replacement
func TestLangWithPlaceholder(t *testing.T) {
	lang := &userLang{
		code: "en",
		data: map[string]string{
			"Hello, {0}": "Hello, {0}",
		},
	}

	result := lang.Lang("Hello, {0}", "John")
	expected := "Hello, John"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// Test Lang method when translation is not available in the language file
func TestLangMissingKey(t *testing.T) {
	lang := &userLang{
		code: "en",
		data: map[string]string{}, // No translations loaded
	}

	result := lang.Lang("Non-existent key")
	expected := "Non-existent key"
	if result != expected {
		t.Errorf("Expected fallback to original text, got '%s'", result)
	}
}

// Test HasLang to verify language availability
func TestHasUserLang(t *testing.T) {
	appLang := &appLang{
		langs: []string{"en", "es"},
	}

	if !appLang.HasLang("en") {
		t.Error("Expected 'en' to be available, but it was not found")
	}

	if appLang.HasLang("fr") {
		t.Error("Expected 'fr' to be unavailable, but it was found")
	}
}

// Test Lang method for translation
func TestLangTranslation(t *testing.T) {
	lang := &userLang{
		code: "es",
		data: map[string]string{
			"Sign in": "Iniciar sesión",
		},
	}

	result := lang.Lang("Sign in")
	expected := "Iniciar sesión"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// Test LangCode method to ensure it returns the correct language code
func TestLangCode(t *testing.T) {
	lang := &userLang{
		code: "en",
	}

	if lang.LangCode() != "en" {
		t.Errorf("Expected 'en', got '%s'", lang.LangCode())
	}
}
