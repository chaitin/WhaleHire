import type { NextConfig } from "next";

const BACKEND_HOST = process.env.BACKEND_HOST || "http://localhost:8888";

const nextConfig: NextConfig = {
  output: "standalone",
  turbopack: {
    rules: {
      "*.svg": {
        loaders: ["@svgr/webpack"],
        as: "*.js",
      },
    },
  },
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${BACKEND_HOST}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
