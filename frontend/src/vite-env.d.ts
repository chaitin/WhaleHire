/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string;
  readonly VITE_APP_TITLE: string;
  readonly VITE_APP_VERSION: string;
  readonly VITE_APP_ENV: string;
  readonly VITE_MAX_FILE_SIZE: string;
  readonly VITE_ALLOWED_FILE_TYPES: string;
  readonly VITE_ENABLE_MOCK: string;
  readonly VITE_DEBUG: string;
  readonly VITE_ENABLE_DEVTOOLS: string;
  readonly VITE_DEV_SERVER_PORT: string;
  readonly VITE_DEV_SERVER_HOST: string;
  readonly VITE_BACKEND_HOST: string;
  readonly VITE_BUILD_SOURCEMAP: string;
  readonly VITE_BUILD_MINIFY: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
