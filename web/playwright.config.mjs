import {defineConfig} from '@playwright/test';
export default defineConfig({
 testDir:'./tests',use:{baseURL:'http://localhost:3000',browserName:'chromium'},
 webServer:{command:'npm run dev',url:'http://localhost:3000',reuseExistingServer:false,timeout:120000}
});