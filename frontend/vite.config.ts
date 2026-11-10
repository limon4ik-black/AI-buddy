import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: { "/api": { target: "http://127.0.0.1:8080" } },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/test-setup.ts"],
    coverage: {
      provider: "v8",
      include: ["src/**/*.{ts,tsx}"],
      exclude: ["src/**/*.test.*", "src/test-setup.ts", "src/test-fixtures.ts"],
      reporter: ["text", "json-summary", "html", "lcov"],
      reportsDirectory: "../reports/frontend",
      thresholds: { lines: 81, statements: 81 },
    },
  },
});
