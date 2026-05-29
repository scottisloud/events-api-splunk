const replace = require("replace-in-file");
const package = require("../package.json");

const appConfVersion = {
  files: "./default/app.conf",
  from: /version \= .*/,
  to: `version = ${package.version}`,
};

const appConfBuild = {
  files: "./default/app.conf",
  from: /build \= .*/,
  to: `build = ${package.version.replace(/\./g, "")}`,
};

const wizardVersion = {
  files: "./appserver/static/javascript/views/setup_page.js",
  from: /export const VERSION \= ".*";/,
  to: `export const VERSION = "${package.version}";`,
};

replace(appConfVersion);
replace(appConfBuild);
replace(wizardVersion);
