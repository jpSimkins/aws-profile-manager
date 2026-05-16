package terminal

import "testing"

// --- LaunchConfig.envVars ---

func TestEnvVars_ContainsRequiredKeys(t *testing.T) {
	cfg := LaunchConfig{
		ProfileName: "commercial-acme-prod-Admin",
		Region:      "us-east-1",
	}
	env := cfg.envVars()

	if env["AWS_PROFILE"] != "commercial-acme-prod-Admin" {
		t.Errorf("AWS_PROFILE = %q, want commercial-acme-prod-Admin", env["AWS_PROFILE"])
	}
	if env["AWS_DEFAULT_REGION"] != "us-east-1" {
		t.Errorf("AWS_DEFAULT_REGION = %q, want us-east-1", env["AWS_DEFAULT_REGION"])
	}
	if env["AWS_SDK_LOAD_CONFIG"] != "1" {
		t.Errorf("AWS_SDK_LOAD_CONFIG = %q, want 1", env["AWS_SDK_LOAD_CONFIG"])
	}
}

func TestEnvVars_AlwaysHasThreeBaseKeys(t *testing.T) {
	// Only AWS_SDK_LOAD_CONFIG is unconditional; the others are omitted when empty.
	cfg := LaunchConfig{ProfileName: "p", Region: "r"}
	env := cfg.envVars()
	for _, key := range []string{"AWS_PROFILE", "AWS_DEFAULT_REGION", "AWS_SDK_LOAD_CONFIG"} {
		if _, ok := env[key]; !ok {
			t.Errorf("missing expected key %q in envVars()", key)
		}
	}
}

func TestEnvVars_MergesExtraEnv(t *testing.T) {
	cfg := LaunchConfig{
		ProfileName: "prod",
		Region:      "us-east-1",
		ExtraEnv:    map[string]string{"MY_CUSTOM_VAR": "hello"},
	}
	env := cfg.envVars()

	if env["MY_CUSTOM_VAR"] != "hello" {
		t.Errorf("MY_CUSTOM_VAR = %q, want hello", env["MY_CUSTOM_VAR"])
	}
}

func TestEnvVars_ExtraEnvCanOverrideDefaults(t *testing.T) {
	cfg := LaunchConfig{
		ProfileName: "prod",
		Region:      "us-east-1",
		ExtraEnv:    map[string]string{"AWS_DEFAULT_REGION": "eu-west-1"},
	}
	env := cfg.envVars()

	if env["AWS_DEFAULT_REGION"] != "eu-west-1" {
		t.Errorf("ExtraEnv should override defaults; got %q", env["AWS_DEFAULT_REGION"])
	}
}

func TestEnvVars_NilExtraEnvDoesNotPanic(t *testing.T) {
	cfg := LaunchConfig{ProfileName: "prod", Region: "us-east-1", ExtraEnv: nil}
	// Should not panic.
	_ = cfg.envVars()
}

func TestEnvVars_EmptyProfileAndRegion(t *testing.T) {
	cfg := LaunchConfig{}
	env := cfg.envVars()
	// AWS_SDK_LOAD_CONFIG must always be present.
	if env["AWS_SDK_LOAD_CONFIG"] != "1" {
		t.Error("AWS_SDK_LOAD_CONFIG must always be set to 1")
	}
	// AWS_PROFILE and AWS_DEFAULT_REGION must NOT be injected when empty,
	// otherwise the AWS CLI fails with "profile () could not be found".
	if _, ok := env["AWS_PROFILE"]; ok {
		t.Error("AWS_PROFILE must not be injected when ProfileName is empty")
	}
	if _, ok := env["AWS_DEFAULT_REGION"]; ok {
		t.Error("AWS_DEFAULT_REGION must not be injected when Region is empty")
	}
}

// --- LaunchConfig.Command ---

func TestLaunchConfig_CommandIsOptional(t *testing.T) {
	cfg := LaunchConfig{ProfileName: "prod", Region: "us-east-1"}
	if cfg.Command != "" {
		t.Error("Command should default to empty string")
	}
}

func TestLaunchConfig_CommandDoesNotAffectEnvVars(t *testing.T) {
	// Setting a Command should not change the env vars map.
	cfg := LaunchConfig{
		ProfileName: "prod",
		Region:      "us-east-1",
		Command:     "aws sso login --sso-session test",
	}
	env := cfg.envVars()
	if _, ok := env["Command"]; ok {
		t.Error("Command should not appear in envVars map")
	}
	if env["AWS_PROFILE"] != "prod" {
		t.Errorf("AWS_PROFILE should still be set correctly; got %q", env["AWS_PROFILE"])
	}
}
