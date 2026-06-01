<<<<<<< HEAD
import { defineConfig, createLogger } from 'vite'
import react from '@vitejs/plugin-react'

// Suppress WebSocket proxy errors when backend is down
const logger = createLogger();
const originalError = logger.error;
logger.error = (msg, options) => {
  if (msg.includes('ws proxy error')) return;
  originalError(msg, options);
};

// https://vite.dev/config/
export default defineConfig({
  customLogger: logger,
=======
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
>>>>>>> 9b97903 (Fixing conflict)
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:3939',
        changeOrigin: true,
        ws: true,
      },
    },
  },
})
