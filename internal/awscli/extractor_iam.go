package awscli

import "aws-profile-manager/internal/logging"

// extractIamFields extracts and classifies IAM and AssumeRole profile fields.
//
// This function processes AWS IAM user and AssumeRole profile properties and
// sets the appropriate fields in the AwsCliProfile struct. It also updates
// the profile type based on detected authentication methods.
//
// Profile Type Detection:
//   - IAM: Profiles with aws_access_key_id, aws_secret_access_key, or credential_process
//   - AssumeRole: Profiles with role_arn and source_profile
//
// IAM Authentication Methods:
//   - Static credentials: aws_access_key_id + aws_secret_access_key
//   - Credential process: External command that provides credentials
//
// AssumeRole Configuration:
//   - role_arn: ARN of the role to assume
//   - source_profile: Profile providing base credentials for assumption
//
// Parameters:
//   - profile: Profile struct to update with IAM/AssumeRole fields
//   - key: Config property name
//   - value: Config property value
func (e *Extractor) extractIamFields(profile *AwsCliProfile, key, value string) {
	switch key {
	case "aws_access_key_id":
		profile.HasAccessKey = true
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeIAM
		}
		logging.Debug.Logf("\t🔹 Detected IAM access key for profile: %s", profile.Name)

	case "aws_secret_access_key":
		profile.HasSecretKey = true
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeIAM
		}
		logging.Debug.Logf("\t🔹 Detected IAM secret key for profile: %s", profile.Name)

	case "credential_process":
		profile.HasCredentialProc = true
		profile.CredentialProcess = value
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeIAM
		}
		logging.Debug.Logf("\t🔹 Detected credential process for profile %s: %s", profile.Name, value)

	case "role_arn":
		// This could be AssumeRole profile
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeAssumeRole
		}
		logging.Debug.Logf("\t🔹 Detected role ARN (assume role) for profile %s: %s", profile.Name, value)

	case "source_profile":
		// This is part of AssumeRole configuration
		if profile.Type == ProfileTypeUnknown {
			profile.Type = ProfileTypeAssumeRole
		}
		logging.Debug.Logf("\t🔹 Detected source profile (assume role) for profile %s: %s", profile.Name, value)
	}
}
