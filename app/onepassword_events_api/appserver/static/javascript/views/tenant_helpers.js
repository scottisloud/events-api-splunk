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

function audienceSlug(audience) {
  let slug = audience
    .split("")
    .map((char) => (/[a-zA-Z0-9]/.test(char) ? char : "_"))
    .join("");
  slug = slug.replace(/^_+|_+$/g, "");
  while (slug.includes("__")) {
    slug = slug.replace(/__/g, "_");
  }
  return slug;
}

async function sha256HexPrefix(input, numBytes) {
  const data = new TextEncoder().encode(input);
  const digest = await crypto.subtle.digest("SHA-256", data);
  const bytes = new Uint8Array(digest, 0, numBytes);
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, "0")).join(
    ""
  );
}

// Must stay aligned with utils.TenantKeyFromAudience in Go.
export async function tenantKeyFromAudience(audience) {
  const slug = audienceSlug(audience);
  if (!slug) {
    return "t_" + (await sha256HexPrefix(audience, 8));
  }
  if (slug.length > 48) {
    return slug.slice(0, 32) + "_" + (await sha256HexPrefix(audience, 4));
  }
  return slug;
}

export function validateTenantId(tenantId) {
  if (!TENANT_ID_PATTERN.test(tenantId)) {
    return "tenant_id may only contain letters, numbers, underscores, and hyphens (max 64 characters).";
  }
}

export function validateTenantIdRequired(tenantId) {
  if (!tenantId.trim()) {
    return "Enter a tenant_id label (for example acme-corp). This is used in Splunk searches and is not derived from the API hostname.";
  }
  return validateTenantId(tenantId.trim());
}

export function validateTenantKey(tenantKey) {
  if (tenantKey === "default") {
    return;
  }
  if (!TENANT_ID_PATTERN.test(tenantKey)) {
    return "tenant_key may only contain letters, numbers, underscores, and hyphens (max 64 characters).";
  }
}

export function eventsURLFromAudience(audience) {
  if (audience === audienceDEPRECATED) {
    return null;
  }
  return `https://${audience}`;
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
