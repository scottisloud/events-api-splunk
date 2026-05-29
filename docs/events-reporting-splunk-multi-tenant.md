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

Each 1Password tenant is a separate 1Password Business account (or separate Events API integration) with its own bearer token. Most tenants use the same Events API host (`events.1password.com`); regional accounts may use a different host in the JWT `aud` claim. The **tenant_id** label you choose identifies each account in Splunk; it also becomes the internal config key for that tenant’s stanza, token, and cursor files.

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
5. Enter a **tenant_id label** (for example `acme-corp` or `eu-security`). This value is required, must be unique across tenants, and is stamped on every ingested event for Splunk filtering.  
   - Use letters, numbers, underscores, and hyphens only (max 64 characters).  
   - Choose a name that is meaningful in your environment (customer name, team, region, etc.).
6. Select **Add tenant**.

Adding or removing a tenant automatically restarts the enabled 1Password scripted inputs so ingestion picks up the updated tenant list. If events from a new tenant do not appear within a few minutes, check `$SPLUNK_HOME/var/log/splunk/splunkd.log` for lines containing `skipping tenant` or `failed:` for that tenant.

Splunk stores:

| Item | Location |
| --- | --- |
| Tenant settings | `[tenant.<tenant_id>]` in `local/events_reporting.conf` |
| Token | Storage password `events_api_token_<tenant_id>` in realm `events_reporting_realm` |
| Cursors | `local/signin_cursor_store_<tenant_id>`, etc. |

Multiple tenants may use tokens with the same Events API endpoint (same JWT `aud` host). Each tenant still needs a **unique tenant_id label** and its own bearer token. Deployments created before this change may still use an audience-derived stanza name (for example `[tenant.events_1password_com]` with a different display label); those continue to work without reconfiguration.

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
- **Setup** and **Manage Tenants** views require the Splunk `admin` role. Direct edits to `events_reporting.conf` stanzas also require `admin`.
- The setup UI validates token structure (JWT format and audience). **Full token verification** (introspect against the 1Password Events API) runs server-side when scripted inputs start polling for each tenant. Check `splunkd.log` for `token introspect failed` if a stored token is invalid.
- Any Splunk role that can search the shared indexes can see events from **all** connected tenants. Dashboard `tenant_id` filters are not access controls.
- For new tenants, the **tenant_id** label is also the internal config key (stanza name, secret suffix, cursor file suffix). Older deployments may use a legacy audience-derived stanza name with a separate display label.
- See [About 1Password Events Reporting security](https://support.1password.com/events-reporting-security/) for general Events API security practices.

### Restricting access by tenant (recommended at scale)

Because all tenants share the same indexes, use Splunk role-based access in addition to dashboard filters:

1. **Separate indexes per tenant** (strongest isolation): route each tenant to its own index in `inputs.conf` and grant roles index-scoped access. This requires custom configuration beyond the default shared-index setup.
2. **Search filters on roles**: for shared indexes, assign each team a Splunk role with a `srchFilter` such as `tenant_id=acme-corp` so searches only return their tenant's events.
3. **Field-level security**: optionally hide or mask `tenant_id` values outside a role's allowed set.

These controls are Splunk platform settings, not enforced by the add-on UI alone.

## Scale and performance (many tenants)

The multi-tenant build is designed to poll **each enabled tenant concurrently** (one goroutine per tenant, per scripted input). With **19 tenants** and all three event types enabled, expect roughly:

| Resource | Approximate load |
| --- | --- |
| Scripted input processes | 3 long-running processes |
| Concurrent tenant pollers | 19 per process (57 total poll loops) |
| Idle API requests | ~5–6 requests/second across all tenants and inputs (10s idle backoff) |
| Cursor files | up to 57 under `local/` (19 tenants × 3 event types) |
| Storage passwords | 1 per tenant |

**Operational tips:**

- Confirm [1Password Events API rate limits](https://support.1password.com/events-reporting-security/) with 1Password for your expected volume before production rollout.
- Watch `$SPLUNK_HOME/var/log/splunk/splunkd.log` for `tenant "<name>" failed` after adding many tenants.
- Adding tenants triggers an app reload; batch tenant onboarding when possible.
- Tenant order in config is processed deterministically (sorted by `tenant_key`).

**Capacity testing:** before deploying 19 production tenants, run a staging Splunk instance with representative tokens and verify indexing lag, search latency, and API error rates under your expected event volume.

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

**Additional tenant** (`[tenant.<tenant_id>]`):

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
| Additional | `events_api_token_<tenant_id>` |

Realm for all: `events_reporting_realm`.

## Learn more

- [Official setup (single tenant)](https://support.1password.com/events-reporting-splunk/)
- [Splunk Cloud Classic Experience](https://support.1password.com/events-reporting-splunk-classic/)
- [1Password Events API reference](https://1password.dev/events-api/reference/)
- [Events Reporting security](https://support.1password.com/events-reporting-security/)
- [Repository README](../README.md) — build and development
