# Get started with multiple 1Password tenants in Splunk

This guide describes how to use the **multi-tenant** build of [1Password Events Reporting for Splunk](https://github.com/scottisloud/events-api-splunk). It extends the standard setup documented on the [1Password Support site](https://support.1password.com/events-reporting-splunk/).

If you are using the unmodified Splunkbase add-on, follow the official guide only. Multi-tenant support is available in this fork and requires building or installing a release from this repository.

## What’s different from the standard integration?

| Topic | Standard add-on | Multi-tenant build |
| --- | --- | --- |
| 1Password accounts | One token, one Events API endpoint | Multiple tokens, one per 1Password tenant |
| Splunk indexes | Per event type (or `main`) | **Shared indexes** across all tenants |
| Identifying events | Not needed | Every event includes a **`tenant_id`** field |
| Setup UI | Single setup wizard | Initial wizard + **Manage Tenants** |
| Credentials | One storage password | One storage password **per tenant** |

Each 1Password tenant is a separate 1Password Business account (or separate Events API integration) with its own bearer token. The JWT `aud` claim identifies the Events API host for that tenant; the add-on derives an internal `tenant_key` from it automatically.

## Before you begin

Complete the same prerequisites as the [official guide](https://support.1password.com/events-reporting-splunk/):

- Splunk Enterprise or Splunk Cloud (Victoria Experience)
- [1Password Business](https://1password.com/business-security)
- Splunk admin access to install and configure the add-on
- Owner or administrator access in each 1Password account you want to connect

You must be able to issue an Events API bearer token for **each** 1Password tenant.

## Step 1: Create indexes for event types (unchanged)

Follow [Step 1 on the support site](https://support.1password.com/events-reporting-splunk/#step-1-create-an-index-for-each-event-type), or index everything on `main`.

**Multi-tenant note:** All tenants use the **same** index per event type. You do not create separate indexes per 1Password account. Use `tenant_id` in searches and dashboards to separate tenants.

## Step 2: Install the add-on and set up the first tenant

### Install the multi-tenant build

1. Build or download a package from this repository (see the [repository README](../README.md)).
2. [Install the add-on](https://docs.splunk.com/Documentation/AddOns/released/Overview/Installingadd-ons) in Splunk Web, same as the standard add-on.
3. Select **Set up now** and complete the initial wizard. This configures the **default** tenant and matches the [official Step 2](https://support.1password.com/events-reporting-splunk/#step-2-set-up-the-1password-add-on):

   1. Select **Generate an Events API token** (opens 1Password.com).
   2. Sign in to your **first** 1Password account.
   3. Create the integration and issue a token with the event types you need.
   4. Copy the token into Splunk and select **Next**.
   5. Choose which index to use for each event type (these indexes apply to **all** tenants).
   6. Select **Submit**, then **Finish**.

The default tenant is stored as:

- Config stanza: `[config]` in `$SPLUNK_HOME/etc/apps/onepassword_events_api/local/events_reporting.conf`
- Storage password: `events_reporting_realm` / `events_api_token`
- Events field: `tenant_id=default` (unless you set a custom label in config)

Existing deployments of the standard add-on upgrade without re-setup; they continue as the default tenant.

### Enable scripted inputs

As in the official guide, enable the 1Password scripted inputs under **Settings → Data Inputs → Scripts** (or use the indexes chosen during setup). Forwarder installs follow the same [distributed install](https://docs.splunk.com/Documentation/AddOns/released/Overview/Distributedinstall) guidance as the standard add-on.

## Step 3: Add additional 1Password tenants

After the first tenant is configured:

1. Sign in to **Splunk Web**.
2. Open the **1Password Events Reporting for Splunk** app from the Apps menu.
3. Select **Manage Tenants** in the navigation bar.

### Add a tenant

1. In **1Password.com**, sign in to the **next** 1Password account (not the one used for initial setup).
2. Go to [Integrations](https://start.1password.com/integrations/active) and open or create an **Events Reporting** integration for Splunk.
3. Issue a new bearer token with the required event types. Copy the token.
4. In Splunk **Manage Tenants**, paste the token into **Events API Token**.
5. Optionally enter a **tenant_id label** (for example `acme-corp` or `eu-security`).  
   - Use letters, numbers, underscores, and hyphens only (max 64 characters).  
   - If you leave this blank, the add-on uses a slug derived from the token’s API host.
6. Select **Add tenant**.

Splunk stores:

| Item | Location |
| --- | --- |
| Tenant settings | `[tenant.<tenant_key>]` in `local/events_reporting.conf` |
| Token | Storage password `events_api_token_<tenant_key>` in realm `events_reporting_realm` |
| Cursors | `local/signin_cursor_store_<tenant_key>`, etc. |

The internal `tenant_key` is computed from the JWT audience and must be unique. You cannot add two tokens for the same 1Password Events API endpoint under different labels.

### Remove a tenant

1. Open **Manage Tenants**.
2. Select **Remove** next to the tenant (not available for the **default** tenant).
3. Confirm removal.

This deletes only that tenant’s config stanza and storage password. Other tenants and the default `[config]` setup are untouched.

## Step 4: Create search macros (unchanged)

Follow [Step 3 on the support site](https://support.1password.com/events-reporting-splunk/#step-3-create-a-search-macro) to create `1password_signin_attempts_index`, `1password_item_usages_index`, and `1password_audit_events_index` macros. Definitions are the same because indexes are shared.

## Search and dashboard examples

Filter by tenant using the `tenant_id` field injected at ingest time:

```spl
`1password_signin_attempts_index` tenant_id=acme-corp
```

```spl
`1password_item_usages_index` tenant_id=default
```

```spl
index=main sourcetype="1password:insights:audit_events" tenant_id=eu-security
```

Count sign-in attempts per tenant:

```spl
`1password_signin_attempts_index`
| stats count by tenant_id
```

The field alias `onepassword_tenant_id` is also available on all three sourcetypes.

## Update or rotate a token

### Default tenant (initial setup)

Same as the [official “Update a bearer token in Splunk”](https://support.1password.com/events-reporting-splunk/#update-a-bearer-token-in-splunk) appendix:

1. Issue a replacement token in 1Password.com.
2. In Splunk, open the add-on → **Launch app** → setup flow → **I already have my Events API token**.
3. Paste the new token and submit. Only the default tenant’s credential is updated.

### Additional tenants

1. Issue a new token in the correct 1Password account.
2. Open **Manage Tenants**.
3. Add the tenant again with the same **tenant_id** label if you want to keep search filters stable, **or** remove the old tenant entry first, then add with the new token.

**Important:** Revoking a token in 1Password stops ingestion for that tenant until Splunk is updated. Issue a replacement before revoking, as described in the [official revoke guidance](https://support.1password.com/events-reporting-splunk/#revoke-a-bearer-token).

## Security notes

- Tokens are stored in Splunk **storage/passwords** (encrypted by Splunk). Each tenant has its own password stanza; updating one tenant does not delete others.
- Any Splunk role with access to the app namespace can read stored tokens. Restrict app permissions as you would for the standard add-on.
- Use `tenant_id` labels for searches only. File paths and secrets use the internal `tenant_key`, not the display label.
- See [About 1Password Events Reporting security](https://support.1password.com/events-reporting-security/) for general Events API security practices.

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| No events for one tenant | Token valid in 1Password; correct account; scripted inputs enabled; `splunkd.log` for `tenant "<name>" failed` |
| `panic: could not decode toml` on `tenantId = mspc` | Splunk wrote an unquoted `tenantId`; use `tenantId = "mspc"` or upgrade to a build that normalizes this automatically |
| Duplicate tenant error on add | That 1Password endpoint is already configured (same JWT `aud`) |
| All tenants stopped | Splunk or forwarder connectivity; not specific to multi-tenant |
| Wrong tenant in results | Verify `tenant_id` in raw event JSON; check label used when adding tenant |

Enable debug logging for the app and inspect:

```text
$SPLUNK_HOME/var/log/splunk/splunkd.log
```

Look for lines containing `Processing tenant` or `tenant ... failed`.

## Appendix: Configuration reference

### `events_reporting.conf`

**Legacy default tenant** (`[config]`):

```toml
[config]
limit = 100
startAt = 2020-01-01T00:00:00Z
signInCursorFile = "/etc/apps/onepassword_events_api/local/signin_cursor_store"
itemUsageCursorFile = "/etc/apps/onepassword_events_api/local/itemusage_cursor_store"
auditEventsCursorFile = "/etc/apps/onepassword_events_api/local/auditevents_cursor_store"
```

**Additional tenant** (`[tenant.<tenant_key>]`):

```toml
[tenant.acme_corp]
tenantId = "acme-corp"
enabled = true
limit = 100
startAt = 2020-01-01T00:00:00Z
```

### Storage passwords

| Tenant | Name |
| --- | --- |
| Default (initial setup) | `events_api_token` |
| Additional | `events_api_token_<tenant_key>` |

Realm for all: `events_reporting_realm`.

## Learn more

- [Official setup (single tenant)](https://support.1password.com/events-reporting-splunk/)
- [Splunk Cloud Classic Experience](https://support.1password.com/events-reporting-splunk-classic/)
- [1Password Events API reference](https://1password.dev/events-api/reference/)
- [Events Reporting security](https://support.1password.com/events-reporting-security/)
- [Repository README](../README.md) — build and development
