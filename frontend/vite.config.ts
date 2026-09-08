import { defineConfig } from "vite";

export default defineConfig({
  server: {
    port: 8080,
    proxy: {
      "/ws": {
        target: "http://localhost:3000",
        ws: true,
      },
      "/health": "http://localhost:3000",
    },
  },
});
