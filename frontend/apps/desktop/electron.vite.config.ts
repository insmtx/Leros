import { resolve } from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "electron-vite";

export default defineConfig({
	main: {
		// 主进程依赖直接打进产物，避免 electron-builder 在打包时沿着 workspace 依赖树继续扫描 node_modules。
	},
	preload: {
		// preload 同样内联依赖，减少 Windows 下 pnpm monorepo 的依赖收集开销。
	},
	renderer: {
		server: {
			port: Number(process.env.DESKTOP_RENDERER_PORT) || 5175,
			strictPort: true,
		},
		plugins: [react(), tailwindcss()],
		resolve: {
			alias: {
				"@": resolve("src/renderer/src"),
			},
			dedupe: ["react", "react-dom"],
		},
	},
});
