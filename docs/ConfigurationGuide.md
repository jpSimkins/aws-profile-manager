# Configuration Guide

The AWS Profile Manager configuration file is a JSON document that describes your AWS organizations, accounts, and roles. You give it to `aws-profile-manager install` and it generates all your AWS CLI profiles automatically.

## Table of Contents
- [Quick Start](#quick-start)
- [Schema Overview](#schema-overview)
- [Top-Level Fields](#top-level-fields)
- [SSO Organizations](#sso-organizations)
- [IAM Users](#iam-users)
- [AssumeRole Chains](#assumerole-chains)
- [Generic Profiles](#generic-profiles)
- [Presets](#presets)
- [Profile Naming Reference](#profile-naming-reference)
- [Full Example](#full-example)

---

## Quick Start

The minimal config for an SSO-based organization:

```json
{
  "version": "2.0",
  "managed": {
    "organizations": {
      "my-org": {
        "name": "My Organization",
        "partitions": {
          "commercial": {
            "url": "https://my-org.awsapps.com/start",
            "default_region": "us-east-1",
            "regions": ["us-east-1"],
            "accounts": [
              {
                "alias": "dev",
                "name": "Development",
                "id": "123456789012"
              }
            ],
            "roles": ["Developer", "ReadOnly"]
          }
        }
      }
    }
  }
}
```

Running `aws-profile-manager install --config my-config.json` with this file produces:

```ini
[profile commercial-dev-Developer]
sso_session   = my-org-commercial
sso_account_id = 123456789012
sso_role_name  = Developer
region         = us-east-1

[profile commercial-dev-ReadOnly]
sso_session   = my-org-commercial
sso_account_id = 123456789012
sso_role_name  = ReadOnly
region         = us-east-1

[sso-session my-org-commercial]
sso_start_url    = https://my-org.awsapps.com/start
sso_region       = us-east-1
sso_registration_scopes = sso:account:access
```

---

## Schema Overview

The config file uses a hierarchical structure:

```
Schema
└── managed
    ├── organizations       (SSO authentication)
    │   └── <org-alias>
    │       └── partitions
    │           └── commercial / govcloud
    │               ├── accounts[]
    │               └── roles[]
    ├── iam_users[]         (static credentials or credential_process)
    ├── assume_role_chains[] (role assumption)
    └── generic_profiles[]  (custom profiles)
```

All fields under `managed` are **replaced** every time you run `install` or `sync fetch`. Your personal profiles (outside the managed section) are always preserved.

---

## Top-Level Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Schema version. Use `"2.0"`. |
| `managed` | object | Yes* | Company-controlled profiles, replaced on every install. |
| `presets` | object | No | Named filter configurations for the GUI installer. |

\* At least one of `managed` or `unmanaged` must be present. In practice, all distributed config files use `managed`.

---

## SSO Organizations

This is the primary section for AWS IAM Identity Center (SSO) users. Each entry under `organizations` represents one SSO portal and generates SSO profiles and sessions.

### Organization Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable display name. |
| `description` | string | No | Optional description shown in the GUI. |
| `partitions` | object | Yes | AWS partitions. Keys must be `"commercial"` or `"govcloud"`. |

### Partition Fields

A partition is either the standard commercial AWS partition or GovCloud. Each partition has its own SSO portal URL, regions, accounts, and roles.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | string | Yes | SSO start URL from the AWS access portal. |
| `default_region` | string | Yes | Default region for profiles in this partition. |
| `regions` | array | Yes | All regions to generate profiles for. Must include `default_region`. |
| `accounts` | array | Yes | AWS accounts in this partition. |
| `roles` | array | Yes | IAM roles to generate profiles for. Every account gets every role. |

**Regions and profile generation**: A profile is only generated for a non-default region if that region appears in `regions`. The default region produces a profile without a region suffix; other regions produce a profile with `--<region>` appended.

```json
"default_region": "us-east-1",
"regions": ["us-east-1", "us-west-2"]
```

With two accounts and one role this produces:
- `commercial-dev-Developer` (us-east-1, the default — no suffix)
- `commercial-dev-Developer--us-west-2`
- `commercial-prod-Developer`
- `commercial-prod-Developer--us-west-2`

### Account Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `alias` | string | Yes | Short identifier used in profile names (e.g., `"prod"`, `"dev-sandbox"`). Use lowercase and hyphens. |
| `name` | string | Yes | Display name shown in the GUI (e.g., `"Production Environment"`). |
| `id` | string | Yes | 12-digit AWS account ID (e.g., `"123456789012"`). |

**Choosing a good alias**: The alias becomes part of every profile name for that account, so keep it short and descriptive. `prod`, `dev`, `staging`, `dev-sandbox` are typical patterns.

### SSO Example

```json
"organizations": {
  "my-company": {
    "name": "My Company",
    "description": "Main company AWS organization",
    "partitions": {
      "commercial": {
        "url": "https://my-company.awsapps.com/start",
        "default_region": "us-east-1",
        "regions": ["us-east-1", "us-west-2"],
        "accounts": [
          {
            "alias": "prod",
            "name": "Production",
            "id": "111122223333"
          },
          {
            "alias": "staging",
            "name": "Staging",
            "id": "444455556666"
          },
          {
            "alias": "dev",
            "name": "Development",
            "id": "777788889999"
          }
        ],
        "roles": ["Developer", "ReadOnly", "SystemAdmin"]
      }
    }
  }
}
```

### GovCloud Partition

GovCloud requires a separate SSO portal URL and uses `us-gov-*` regions. You can have both partitions under the same organization:

```json
"my-company": {
  "name": "My Company",
  "partitions": {
    "commercial": {
      "url": "https://my-company.awsapps.com/start",
      "default_region": "us-east-1",
      "regions": ["us-east-1"],
      "accounts": [{ "alias": "prod", "name": "Production", "id": "111122223333" }],
      "roles": ["Developer"]
    },
    "govcloud": {
      "url": "https://start.us-gov-home.awsapps.com/directory/my-company",
      "default_region": "us-gov-west-1",
      "regions": ["us-gov-west-1"],
      "accounts": [{ "alias": "prod", "name": "Production (GovCloud)", "id": "000011112222" }],
      "roles": ["Developer"]
    }
  }
}
```

---

## IAM Users

Use `iam_users` for profiles that authenticate with static access keys or a credential process command. Unlike SSO profiles, each IAM user produces exactly one profile with the name you specify.

### IAM User Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `profile_name` | string | Yes | Exact AWS CLI profile name to create. |
| `region` | string | No | Default region for this profile. |
| `aws_access_key_id` | string | No* | Static AWS access key ID. |
| `aws_secret_access_key` | string | No* | Static AWS secret access key. |
| `credential_process` | string | No* | Shell command that returns credentials as JSON. |

\* At least one of (`aws_access_key_id` + `aws_secret_access_key`) or `credential_process` must be provided.

> **Note**: Storing static credentials in a distributed config file is a security risk. Prefer `credential_process` for any config that will be shared or stored remotely.

### IAM User Example (credential_process)

```json
"iam_users": [
  {
    "profile_name": "ci-runner",
    "region": "us-east-1",
    "credential_process": "aws-vault exec ci-runner --json"
  }
]
```

### IAM User Example (static credentials)

```json
"iam_users": [
  {
    "profile_name": "legacy-service",
    "region": "us-east-1",
    "aws_access_key_id": "AKIAIOSFODNN7EXAMPLE",
    "aws_secret_access_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  }
]
```

---

## AssumeRole Chains

Use `assume_role_chains` for profiles that assume an IAM role using credentials from another profile. These are common in multi-account setups where you log into a central account and then assume roles in child accounts.

### AssumeRole Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `profile_name` | string | Yes | Exact AWS CLI profile name to create. |
| `source_profile` | string | Yes | Existing AWS CLI profile to use as the credential source. |
| `role_arn` | string | Yes | Full ARN of the IAM role to assume (e.g., `arn:aws:iam::123456789012:role/MyRole`). |
| `region` | string | No | Default region for this profile. |
| `mfa_serial` | string | No | ARN of the MFA device, if required by the role's trust policy. |
| `external_id` | string | No | External ID for cross-account role assumptions. |
| `session_name` | string | No | Custom role session name for audit logging. |

### AssumeRole Example

```json
"assume_role_chains": [
  {
    "profile_name": "prod-admin",
    "source_profile": "commercial-hub-Developer",
    "role_arn": "arn:aws:iam::111122223333:role/AdminRole",
    "region": "us-east-1"
  },
  {
    "profile_name": "partner-account",
    "source_profile": "commercial-hub-Developer",
    "role_arn": "arn:aws:iam::444455556666:role/CrossAccountAccess",
    "external_id": "UniqueExternalID"
  }
]
```

---

## Generic Profiles

Use `generic_profiles` for profiles that don't fit the SSO, IAM, or AssumeRole patterns. The `properties` map is written directly to the AWS config file as key-value pairs.

### Generic Profile Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `profile_name` | string | Yes | Exact AWS CLI profile name to create. |
| `properties` | object | Yes | Key-value pairs written directly to the profile block. |

### Generic Profile Example

```json
"generic_profiles": [
  {
    "profile_name": "localstack",
    "properties": {
      "region": "us-east-1",
      "endpoint_url": "http://localhost:4566",
      "aws_access_key_id": "test",
      "aws_secret_access_key": "test"
    }
  }
]
```

This produces:

```ini
[profile localstack]
region            = us-east-1
endpoint_url      = http://localhost:4566
aws_access_key_id = test
aws_secret_access_key = test
```

---

## Presets

Presets are named filter configurations that appear in the GUI installer. They let users quickly install a subset of profiles without manually selecting filters each time. Presets do not affect CLI `install` operations unless `--preset` is specified.

### Preset Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `label` | string | Yes | Display name shown in the GUI (e.g., `"Developer"`). |
| `description` | string | No | Longer explanation of what this preset installs. |
| `organizations` | array | No | Organization aliases to include. Empty = all organizations. |
| `partitions` | array | No | Partition names to include. Empty = all partitions. |
| `accounts` | array | No | Account aliases to include. Empty = all accounts. |
| `roles` | array | No | Role names to include. Empty = all roles. |
| `regions` | array | No | Specific regions to include. Empty = default region only. |
| `all_regions` | bool | No | If `true`, include all configured regions (overrides `regions`). |

**Empty filter = include all**: An omitted or empty array means "no restriction on this dimension." To install only the `Developer` role across all organizations and accounts:

```json
"presets": {
  "developer": {
    "label": "Developer",
    "description": "Developer role only, all accounts",
    "roles": ["Developer"]
  }
}
```

### Preset Examples

```json
"presets": {
  "developer": {
    "label": "Developer",
    "description": "Developer role, single org, default region only",
    "organizations": ["my-org"],
    "roles": ["Developer"]
  },
  "admin": {
    "label": "Admin",
    "description": "All admin roles across all organizations",
    "roles": ["SystemAdmin", "NetworkAdmin"]
  },
  "all-regions": {
    "label": "All Regions",
    "description": "Developer role with profiles for every configured region",
    "roles": ["Developer"],
    "all_regions": true
  },
  "break-glass": {
    "label": "Break Glass",
    "description": "Emergency access — all accounts and partitions",
    "roles": ["BreakGlass"]
  }
}
```

---

## Profile Naming Reference

AWS Profile Manager uses a consistent naming pattern so profiles are easy to identify and filter.

### SSO Profiles

```
<partition>-<account-alias>-<role>
<partition>-<account-alias>-<role>--<region>
```

The region suffix is only added when the region differs from the partition's `default_region`.

| Organization | Partition | Account Alias | Role | Default Region | Profile Region | Generated Profile Name |
|---|---|---|---|---|---|---|
| my-org | commercial | prod | Developer | us-east-1 | us-east-1 | `commercial-prod-Developer` |
| my-org | commercial | prod | Developer | us-east-1 | us-west-2 | `commercial-prod-Developer--us-west-2` |
| my-org | govcloud | prod | SystemAdmin | us-gov-west-1 | us-gov-west-1 | `govcloud-prod-SystemAdmin` |

### SSO Sessions

Each organization+partition combination gets one shared SSO session:

```
<org-alias>-<partition>
```

Examples: `my-org-commercial`, `my-org-govcloud`, `company-commercial`

All profiles within the same org+partition share this session, so you only authenticate once per session.

### IAM and AssumeRole Profiles

These use the exact `profile_name` you specify — no automatic naming is applied.

---

## Full Example

A complete config combining SSO, IAM, AssumeRole, generic profiles, and presets:

```json
{
  "version": "2.0",
  "managed": {
    "organizations": {
      "my-company": {
        "name": "My Company",
        "description": "Main company AWS organization",
        "partitions": {
          "commercial": {
            "url": "https://my-company.awsapps.com/start",
            "default_region": "us-east-1",
            "regions": ["us-east-1", "us-west-2"],
            "accounts": [
              {
                "alias": "prod",
                "name": "Production",
                "id": "111122223333"
              },
              {
                "alias": "staging",
                "name": "Staging",
                "id": "444455556666"
              },
              {
                "alias": "dev",
                "name": "Development",
                "id": "777788889999"
              }
            ],
            "roles": ["Developer", "ReadOnly", "SystemAdmin"]
          },
          "govcloud": {
            "url": "https://start.us-gov-home.awsapps.com/directory/my-company",
            "default_region": "us-gov-west-1",
            "regions": ["us-gov-west-1"],
            "accounts": [
              {
                "alias": "prod",
                "name": "Production (GovCloud)",
                "id": "000011112222"
              }
            ],
            "roles": ["Developer", "SystemAdmin"]
          }
        }
      }
    },
    "iam_users": [
      {
        "profile_name": "ci-runner",
        "region": "us-east-1",
        "credential_process": "aws-vault exec ci-runner --json"
      }
    ],
    "assume_role_chains": [
      {
        "profile_name": "partner-readonly",
        "source_profile": "commercial-prod-Developer",
        "role_arn": "arn:aws:iam::999988887777:role/ReadOnlyAccess",
        "region": "us-east-1"
      }
    ],
    "generic_profiles": [
      {
        "profile_name": "localstack",
        "properties": {
          "region": "us-east-1",
          "endpoint_url": "http://localhost:4566",
          "aws_access_key_id": "test",
          "aws_secret_access_key": "test"
        }
      }
    ]
  },
  "presets": {
    "developer": {
      "label": "Developer",
      "description": "Developer role across all commercial accounts",
      "partitions": ["commercial"],
      "roles": ["Developer"]
    },
    "admin": {
      "label": "Admin",
      "description": "All admin roles, all partitions",
      "roles": ["SystemAdmin"]
    },
    "break-glass": {
      "label": "Break Glass",
      "roles": ["BreakGlass"]
    }
  }
}
```
