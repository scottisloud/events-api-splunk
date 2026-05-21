"use strict";

import React from "react";
import ReactDOM from "react-dom";
import SetupPage from "./views/app";
import TenantManagementPage from "./views/tenant_management_app";
import "../styles/setup_page.css";
import "../styles/switch.css";

const container = document.getElementById("main_container");
const page = container && container.dataset.page;

const Page =
  page === "tenant-management" ? TenantManagementPage : SetupPage;

ReactDOM.render(React.createElement(Page), container);
