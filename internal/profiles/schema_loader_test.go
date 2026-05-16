package profiles

import (
	"context"
	"os"
	"testing"

	schematest "aws-profile-manager/internal/schema/test"
	"aws-profile-manager/internal/task"
	"aws-profile-manager/internal/test"
)

func TestNewSchemaReader(t *testing.T) {
	test.SetupTestEnvironment(t)

	reader := NewSchemaReader(newTestConfig(t))
	if reader == nil {
		t.Fatal("NewSchemaReader should not return nil")
	}
}

func TestSchemaReader_Read_RequiresAtLeastOneSection(t *testing.T) {
	test.SetupTestEnvironment(t)

	reader := NewSchemaReader(newTestConfig(t))
	_, err := reader.Read(context.Background(), SchemaReadOptions{}, task.NoOpReporter{})
	if err == nil {
		t.Fatal("expected error when no sections are enabled")
	}
}

func TestSchemaReader_Read_ManagedOnly(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)
	writer := newConfigWriter(config)
	schemaData := schematest.NewManagedSsoSingle()

	_, _, _, _, err := writer.writeConfig(context.Background(), schemaData, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("failed to write managed test config: %v", err)
	}

	reader := NewSchemaReader(config)
	result, err := reader.Read(context.Background(), SchemaReadOptions{
		IncludeManaged: true,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil schema result")
	}
	if result.Managed == nil {
		t.Fatal("expected managed section to be present")
	}
	if len(result.Managed.Organizations) == 0 {
		t.Fatal("expected managed organizations to be present")
	}
}

func TestSchemaReader_Read_UnmanagedAboveAndBelow(t *testing.T) {
	test.SetupTestEnvironment(t)

	config := newTestConfig(t)
	content := `[profile personal-above]
region = us-east-1
output = json

# START
[profile work-dev]
region = us-west-2
# END

[profile personal-below]
region = eu-west-1
output = yaml
`
	if err := os.WriteFile(config.ConfigPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	reader := NewSchemaReader(config)
	result, err := reader.Read(context.Background(), SchemaReadOptions{
		IncludeUnmanagedAbove: true,
		IncludeUnmanagedBelow: true,
	}, task.NoOpReporter{})
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil schema result")
	}
	if result.Unmanaged == nil {
		t.Fatal("expected unmanaged section to be present")
	}
	if result.Unmanaged.Above == nil {
		t.Fatal("expected unmanaged above section to be present")
	}
	if result.Unmanaged.Below == nil {
		t.Fatal("expected unmanaged below section to be present")
	}
}
