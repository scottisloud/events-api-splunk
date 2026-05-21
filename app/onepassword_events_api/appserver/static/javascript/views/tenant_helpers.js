const audienceDEPRECATED = "com.1password.streamingservice";
const TENANT_ID_PATTERN = /^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$/;

export function parseJWTPayload(token) {
  const parts = token.split(".");
  if (parts.length !== 3) {
    return { error: "This doesn't look like a valid JSON Web Token." };
  }
  try {
    return { payload: JSON.parse(atob(parts[1])) };
  } catch (e) {
    return { error: "This doesn't look like a valid JSON Web Token." };
  }
}

export function validateJWT(token) {
  const parsed = parseJWTPayload(token);
  if (parsed.error) {
    return parsed.error;
  }
  const payload = parsed.payload;
  if (!payload.aud || payload.aud.length !== 1) {
    return "This doesn't look like a valid JSON Web Token.";
  }
  if (payload.aud[0] === audienceDEPRECATED) {
    return "Please generate a new token.";
  }
}

export function tenantKeyFromAudience(audience) {
  let slug = audience.replace(/[^a-zA-Z0-9]/g, "_").replace(/_+/g, "_");
  slug = slug.replace(/^_+|_+$/g, "");
  if (!slug) {
    return "t_" + hashAudience(audience).slice(0, 16);
  }
  if (slug.length > 48) {
    return slug.slice(0, 32) + "_" + hashAudience(audience).slice(0, 8);
  }
  return slug;
}

function hashAudience(audience) {
  // Simple deterministic hash for fallback keys (setup-time only).
  let h = 0;
  for (let i = 0; i < audience.length; i++) {
    h = (h << 5) - h + audience.charCodeAt(i);
    h |= 0;
  }
  return Math.abs(h).toString(16);
}

export function validateTenantId(tenantId) {
  if (!TENANT_ID_PATTERN.test(tenantId)) {
    return "tenant_id may only contain letters, numbers, underscores, and hyphens (max 64 characters).";
  }
}

export function secretNameForTenantKey(tenantKey) {
  if (tenantKey === "default") {
    return "events_api_token";
  }
  return "events_api_token_" + tenantKey;
}

export function defaultTenantOptions(tenantKey) {
  const suffix = tenantKey === "default" ? "" : "_" + tenantKey;
  return {
    limit: 100,
    startAt: "2020-01-01T00:00:00Z",
    signInCursorFile: `"/etc/apps/onepassword_events_api/local/signin_cursor_store${suffix}"`,
    itemUsageCursorFile: `"/etc/apps/onepassword_events_api/local/itemusage_cursor_store${suffix}"`,
    auditEventsCursorFile: `"/etc/apps/onepassword_events_api/local/auditevents_cursor_store${suffix}"`,
  };
}
