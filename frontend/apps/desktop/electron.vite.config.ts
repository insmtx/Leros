import { resolve } from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, externalizeDepsPlugin } from "electron-vite";

export default defineConfig({
	main: {
		plugins: [externalizeDepsPlugin()],
	},
	preload: {
		plugins: [externalizeDepsPlugin()],
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
