import path from "path"
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    allowedHosts: true,
    proxy: {
      "/auth": process.env.VITE_API_URL || "http://localhost:8080",
      "/event.v1": process.env.VITE_API_URL || "http://localhost:8080",
      "/participant.v1": process.env.VITE_API_URL || "http://localhost:8080",
      "/item.v1": process.env.VITE_API_URL || "http://localhost:8080",
    },
  },
})
