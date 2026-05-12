import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	transpilePackages: ["@leros/ui", "@leros/store"],
	allowedDevOrigins: ["172.16.0.160"],
};

export default nextConfig;
