import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": { target: "http://127.0.0.1:8080" },
      "/notes-api": {
        target: "http://127.0.0.1:8079",
        rewrite: (path) => path.replace(/^\/notes-api/, "/api/v1"),
      },
      // Доски отдаёт сервис заметок; путь меняем на его собственный /api/v1.
      "/board-api": {
        target: "http://127.0.0.1:8079",
        rewrite: (path) => path.replace(/^\/board-api/, "/api/v1"),
      },
    },
  },
  test: {
    // Browser scenarios run in a CPU-limited Docker container on developer Macs.
    maxWorkers: 2,
    testTimeout: 15000,
    environment: "jsdom",
    setupFiles: ["./src/test-setup.ts"],
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.test.*",
        "src/test-setup.ts",
        "src/test-fixtures.ts",
      ],
      reporter: ["text", "json-summary", "html", "lcov"],
      reportsDirectory: "../reports/frontend",
      thresholds: { lines: 81, statements: 81 },
    },
  },
});
