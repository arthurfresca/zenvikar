const path = require("path");

/** @type {import('next').NextConfig} */
const nextConfig = {
  output: "standalone",
  outputFileTracingRoot: path.join(__dirname, "../../"),
  transpilePackages: ["@zenvikar/ui", "@zenvikar/types", "@zenvikar/config"],
};

module.exports = nextConfig;
