package awscli

import "aws-profile-manager/internal/logging"

// extractSsoFields extracts and classifies SSO-specific profile fields.
//
// This function processes AWS SSO profile properties and sets the appropriate
// fields in the AwsCliProfile struct. It also updates the profile type to SSO
// when SSO-specific fields are detected.
//
// SSO Profile Detection:
//   - sso_account_id: AWS account ID for SSO authentication
//   - sso_role_name: IAM role name to assume via SSO
//   - sso_session: Reference to SSO session configuration (modern SSO)
//   - sso_start_url: SSO portal URL (legacy SSO profiles)
//
// Parameters:
//   - profile: Profile struct to update with SSO fields
//   - key: Config property name
//   - value: Config property value
func (e *Extractor) extractSsoFields(profile *AwsCliProfile, key, value string) {
	switch key {
	case "sso_account_id":
		profile.AccountID = value
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeSSO
		}
		logging.Debug.Logf("\t🔹 Extracted SSO account ID: %s", value)

	case "sso_role_name":
		profile.RoleName = value
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeSSO
		}
		logging.Debug.Logf("\t🔹 Extracted SSO role name: %s", value)

	case "sso_session":
		profile.SsoSession = value
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeSSO
		}
		logging.Debug.Logf("\t🔹 Extracted SSO session: %s", value)

	case "sso_start_url":
		profile.SsoStartURL = value
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeSSO
		}
		logging.Debug.Logf("\t🔹 Extracted SSO start URL: %s", value)
	}
}
