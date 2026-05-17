import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";
import path from "path";
import wails from "@wailsio/runtime/plugins/vite";

export default defineConfig({
  base: "./",
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 5173,
    strictPort: true,
  },
  plugins: [tailwindcss(), svelte(), wails("./bindings")],
  resolve: {
    alias: {
      "$lib": path.resolve("./src/lib"),
    },
  },
});
