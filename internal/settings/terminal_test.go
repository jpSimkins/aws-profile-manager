package settings

import "testing"

func TestGetDefaultTerminal(t *testing.T) {
	s := GetDefaultTerminal()

	// ExecutablePath should be empty by default — lets the OS default be used.
	if s.ExecutablePath != "" {
		t.Errorf("default ExecutablePath should be empty, got %q", s.ExecutablePath)
	}
}

func TestTerminalValidate_EmptyPath(t *testing.T) {
	s := TerminalSettings{ExecutablePath: ""}
	if err := s.Validate(); err != nil {
		t.Errorf("empty ExecutablePath should be valid, got error: %v", err)
	}
}

func TestTerminalValidate_NonEmptyPath(t *testing.T) {
	s := TerminalSettings{ExecutablePath: "/usr/bin/gnome-terminal"}
	if err := s.Validate(); err != nil {
		t.Errorf("non-empty ExecutablePath should be valid, got error: %v", err)
	}
}

func TestTerminalGetSchema_HasExecutablePathField(t *testing.T) {
	s := GetDefaultTerminal()
	schema := s.GetSchema()

	if _, ok := schema.Fields["executable_path"]; !ok {
		t.Error("schema should contain 'executable_path' field")
	}
}

func TestTerminalGetSchema_ExecutablePathIsNotRequired(t *testing.T) {
	s := GetDefaultTerminal()
	schema := s.GetSchema()

	field := schema.Fields["executable_path"]
	if field.Required {
		t.Error("executable_path should not be required")
	}
}

func TestTerminalGetSchema_ExecutablePathIsStringType(t *testing.T) {
	s := GetDefaultTerminal()
	schema := s.GetSchema()

	field := schema.Fields["executable_path"]
	if field.Type != "file" {
		t.Errorf("executable_path type should be 'file', got %q", field.Type)
	}
}

func TestTerminalGetSchema_DefaultIsEmptyString(t *testing.T) {
	s := GetDefaultTerminal()
	schema := s.GetSchema()

	field := schema.Fields["executable_path"]
	if field.Default != "" {
		t.Errorf("executable_path default should be empty string, got %v", field.Default)
	}
}

func TestTerminalGetSchema_HasVersion(t *testing.T) {
	s := GetDefaultTerminal()
	schema := s.GetSchema()

	if schema.Version == "" {
		t.Error("schema should have a non-empty version")
	}
}
