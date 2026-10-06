import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api/v1": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/files": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/healthz": "http://localhost:8080",
      "/readyz": "http://localhost:8080",
    },
  },
});
