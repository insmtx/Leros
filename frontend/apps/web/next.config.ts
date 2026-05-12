import type { NextConfig } from "next";

const nextConfig: NextConfig = {
	transpilePackages: ["@singeros/ui", "@singeros/store"],
	allowedDevOrigins: ["172.16.0.160"],
};

export default nextConfig;
