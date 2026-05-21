# Overview

This application includes a scripted input to ingest data into Splunk from the Events Reporting API. After completing setup you will be able to monitor and alert on important 1Password event data.

# Binary File Declaration

bin/item_usages - binary source code has been included in the lib folder
bin/signin_attempts - binary source code has also been included in the lib folder

The source code is actually the same for these two binaries, but has been provided twice to meet the Splunk naming specification.

## Program Flow

This program starts in the `onepassword_events_api/default/app.conf`, where the `[install]` stanza's `is_configured` property is set to `false`. This causes Splunk to redirect to it's setup page that is specified so that an admin/user can configure it for use.

In the `onepassword_events_api/default/app.conf`'s, `[ui]` stanza there is a `setup_page` property that points to which resource should be used for the setup page. In this case it's pointing to `default/data/ui/views/setup_page_dashboard.xml`.

Once setup is finished, Splunk will need to be restarted in order to be aware of the new configuration variables. On startup,
the scripted inputs (included in `onepassword_events_api/bin/`) will be triggered and Splunk will index the retrieved 1Password events.

## Multi-tenant support

This build supports multiple 1Password tenants in one Splunk add-on installation.

1. Complete **initial setup** (same as the [official Splunk guide](https://support.1password.com/events-reporting-splunk/)) for your first 1Password account — this becomes the **default** tenant (`tenant_id=default`).
2. Open **Manage Tenants** in the app navigation to add more accounts. Each needs its own Events API token from 1Password.com.
3. Search across tenants in shared indexes using `tenant_id`, for example:

   ```spl
   index=* sourcetype="1password:insights:signin_attempts" tenant_id=acme-corp
   ```

Full step-by-step instructions, token rotation, and troubleshooting: **[docs/events-reporting-splunk-multi-tenant.md](../../docs/events-reporting-splunk-multi-tenant.md)** in the repository root.

The original setup flow continues to configure the **default** tenant via the `[config]` stanza and `events_api_token` storage password for backwards compatibility.

## Setup

Click on the 1Password Application in the Apps navigation pane and follow the setup instructions for the first tenant. Once complete, you will be navigated to Search. Start seeing what data has already been ingested by filtering by the event source type, such as `index=* sourcetype="1password:insights:signin_attempts"` or `index=* sourcetype="1password:insights:item_usages"`. If you don't see any events, try increasing the length of time to "All time".

- **First tenant:** [1Password Support — Events Reporting for Splunk](https://support.1password.com/events-reporting-splunk/)
- **Additional tenants:** [Multi-tenant setup guide](../../docs/events-reporting-splunk-multi-tenant.md)

## Debugging

If you've gone through the installation steps and do not see any ingested events, take a look at the logs at `splunkd.log` to see if there are any actionable steps.

Common errors:

```
ERROR ExecProcessor - message from "/opt/splunk/etc/apps/onepassword_events_api/bin/signin_attempts" panic: introspect request failed: could not unmarshal response: 404 page not found
```

- There is something wrong with your JWT. A common issue is not copying the entire token during the setup flow.
