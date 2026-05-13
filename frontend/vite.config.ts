import { execSync } from "child_process";
import { readFileSync } from "fs";
import path from "path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { tanstackRouter } from "@tanstack/router-vite-plugin";
import { defineConfig } from "vite";

function getPackageVersion(): string {
  const packageJson = JSON.parse(
    readFileSync(new URL("./package.json", import.meta.url), "utf-8"),
  ) as { version?: unknown };

  return typeof packageJson.version === "string" ? packageJson.version : "";
}

function getGitVersion(): string {
  try {
    return execSync("git describe --tags --always").toString().trim();
  } catch {
    return getPackageVersion();
  }
}

function getVendorChunk(id: string): string | undefined {
  const normalizedId = id.replaceAll("\\", "/");

  if (!normalizedId.includes("node_modules")) {
    return undefined;
  }

  if (normalizedId.includes("/react/") || normalizedId.includes("/react-dom/")) {
    return "vendor-react";
  }
  if (normalizedId.includes("/@tanstack/")) {
    return "vendor-tanstack";
  }
  if (normalizedId.includes("/@radix-ui/") || normalizedId.includes("/radix-ui/")) {
    return "vendor-radix";
  }
  if (normalizedId.includes("/lucide-react/")) {
    return "vendor-icons";
  }

  return undefined;
}

// https://vite.dev/config/
export default defineConfig({
  define: {
    __APP_VERSION__: JSON.stringify(getGitVersion()),
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: getVendorChunk,
      },
    },
  },
  plugins: [react(), tanstackRouter(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
