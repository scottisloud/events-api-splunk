import React, { useEffect, useState } from "react";
import * as Setup from "./setup_page.js";
import {
  parseJWTPayload,
  validateJWT,
  validateTenantId,
  validateTenantKey,
  tenantKeyFromAudience,
  secretNameForTenantKey,
} from "./tenant_helpers.js";

const e = React.createElement;

export default class TenantManagementPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      tenants: [],
      authToken: "",
      tenantId: "",
      loading: false,
      result: { success: false, error: "" },
    };
  }

  async componentDidMount() {
    const tenants = await Setup.listTenants(splunkjs);
    this.setState({ tenants });
  }

  refreshTenants = async () => {
    const tenants = await Setup.listTenants(splunkjs);
    this.setState({ tenants });
  };

  handleAddTenant = async () => {
    const tokenError = validateJWT(this.state.authToken);
    if (tokenError) {
      this.setState({ result: { success: false, error: tokenError } });
      return;
    }

    const parsed = parseJWTPayload(this.state.authToken);
    const audience = parsed.payload.aud[0];
    const tenantKey = await tenantKeyFromAudience(audience);
    const keyError = validateTenantKey(tenantKey);
    if (keyError) {
      this.setState({ result: { success: false, error: keyError } });
      return;
    }

    let tenantId = this.state.tenantId.trim();
    if (!tenantId) {
      tenantId = tenantKey;
    }
    const idError = validateTenantId(tenantId);
    if (idError) {
      this.setState({ result: { success: false, error: idError } });
      return;
    }

    const existing = this.state.tenants.find((t) => t.tenantKey === tenantKey);
    if (existing) {
      this.setState({
        result: {
          success: false,
          error: `A tenant is already configured for this 1Password endpoint (key: ${tenantKey}).`,
        },
      });
      return;
    }

    this.setState({ loading: true, result: { success: false, error: "" } });
    try {
      await Setup.addTenant(
        splunkjs,
        this.state.authToken,
        tenantKey,
        tenantId
      );
      await this.refreshTenants();
      this.setState({
        authToken: "",
        tenantId: "",
        loading: false,
        result: {
          success: true,
          error: "",
        },
      });
    } catch (error) {
      console.log(error);
      this.setState({
        loading: false,
        result: {
          success: false,
          error:
            error.message ||
            "Something went wrong while adding the tenant - please try again.",
        },
      });
    }
  };

  handleRemoveTenant = async (tenantKey) => {
    if (!window.confirm(`Remove tenant "${tenantKey}"?`)) {
      return;
    }
    this.setState({ loading: true, result: { success: false, error: "" } });
    try {
      await Setup.removeTenant(splunkjs, tenantKey);
      await this.refreshTenants();
      this.setState({
        loading: false,
        result: { success: true, error: "" },
      });
    } catch (error) {
      console.log(error);
      this.setState({
        loading: false,
        result: {
          success: false,
          error: "Could not remove tenant - please try again.",
        },
      });
    }
  };

  render() {
    const { tenants, authToken, tenantId, loading, result } = this.state;

    return e("div", { className: "container" }, [
      e("div", { className: "main-contents" }, [
        e("div", { className: "title" }, [
          e("h1", { className: "block" }, "Manage 1Password Tenants"),
        ]),
        e("div", { className: "content" }, [
          e(
            "div",
            { className: "description block" },
            "Add additional 1Password tenants to ingest into shared Splunk indexes. Events include a tenant_id field for filtering."
          ),
          e("div", { className: "warning block" }, [
            "Your other Splunk apps or add-ons may be able to access stored Events API tokens. Make sure you trust them before you add tokens.",
          ]),
          e("h3", { className: "block" }, "Configured tenants"),
          tenants.length === 0 &&
            e(
              "p",
              { className: "block" },
              "No additional tenants in [tenant.*] stanzas. The default tenant uses the initial setup configuration."
            ),
          e(
            "ul",
            { className: "block tenant-list" },
            tenants.map((t) =>
              e("li", { key: t.tenantKey }, [
                e("strong", null, t.tenantId),
                ` (key: ${t.tenantKey}, secret: ${secretNameForTenantKey(t.tenantKey)}) `,
                t.tenantKey !== "default" &&
                  e(
                    "button",
                    {
                      className: "btn",
                      disabled: loading,
                      onClick: () => this.handleRemoveTenant(t.tenantKey),
                    },
                    "Remove"
                  ),
              ])
            )
          ),
          e("h3", { className: "block" }, "Add tenant"),
          e("div", { className: "token block" }, [
            e("input", {
              type: "text",
              placeholder: "Events API Token",
              value: authToken,
              onChange: (ev) =>
                this.setState({ authToken: ev.target.value }),
            }),
          ]),
          e("div", { className: "token block" }, [
            e("input", {
              type: "text",
              placeholder: "tenant_id label (optional, e.g. acme-corp)",
              value: tenantId,
              onChange: (ev) =>
                this.setState({ tenantId: ev.target.value }),
            }),
          ]),
          result.error && e("div", { className: "error block" }, result.error),
          result.success &&
            e("div", { className: "success block" }, "Tenant saved."),
          e(
            "button",
            {
              className: "next btn",
              disabled: loading,
              onClick: this.handleAddTenant,
            },
            loading ? "Saving..." : "Add tenant"
          ),
        ]),
      ]),
    ]);
  }
}
