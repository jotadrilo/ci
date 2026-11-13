// Launches the Chromium the workflow installed, to prove the Playwright step works.
import { chromium } from "playwright";

const browser = await chromium.launch();
const page = await browser.newPage();
await page.setContent("<h1>ok</h1>");
const text = await page.textContent("h1");
await browser.close();
if (text !== "ok") {
  console.error(`unexpected heading: ${text}`);
  process.exit(1);
}
console.log("chromium ok");
