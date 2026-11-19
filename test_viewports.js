const fs = require('fs');
const path = require('path');
const { spawn } = require('child_process');

const viewports = [
  // Mobile & Tablet Viewports
  { name: 'Mobile Small', width: 320, height: 640, dpr: 2, mobile: true },
  { name: 'iPhone SE (375x667)', width: 375, height: 667, dpr: 2, mobile: true },
  { name: 'iPhone 12/13/14/15', width: 390, height: 844, dpr: 3, mobile: true },
  { name: 'Android Standard', width: 412, height: 915, dpr: 2.625, mobile: true },
  { name: 'iPhone Plus / Max', width: 414, height: 896, dpr: 3, mobile: true },
  { name: 'iPad Mini / Tablet Portrait', width: 768, height: 1024, dpr: 2, mobile: true },

  // 19 Desktop / Laptop / TV viewports
  { name: 'Minimum desktop', width: 1024, height: 768, dpr: 1, mobile: false },
  { name: 'Netbook', width: 1280, height: 720, dpr: 1, mobile: false },
  { name: 'Small laptop 13"', width: 1280, height: 800, dpr: 1, mobile: false },
  { name: 'Small laptop 14"', width: 1366, height: 768, dpr: 1, mobile: false },
  { name: 'MacBook Air 13" older', width: 1440, height: 900, dpr: 2, mobile: false },
  { name: 'MacBook Air 13" M2/M3', width: 1470, height: 956, dpr: 2, mobile: false },
  { name: 'MacBook Pro 14"', width: 1512, height: 982, dpr: 2, mobile: false },
  { name: 'Windows laptop 15"', width: 1536, height: 864, dpr: 1.25, mobile: false },
  { name: 'MacBook Air 15"', width: 1710, height: 1107, dpr: 2, mobile: false },
  { name: 'MacBook Pro 16"', width: 1728, height: 1117, dpr: 2, mobile: false },
  { name: '24" Full HD', width: 1920, height: 1080, dpr: 1, mobile: false },
  { name: '72" TV scaled', width: 1920, height: 1080, dpr: 2, mobile: false },
  { name: 'iMac 24"', width: 2240, height: 1260, dpr: 2, mobile: false },
  { name: '27" QHD', width: 2560, height: 1440, dpr: 1, mobile: false },
  { name: 'iMac 27" 5K', width: 2560, height: 1440, dpr: 2, mobile: false },
  { name: '32" 4K scaled', width: 2560, height: 1440, dpr: 1.5, mobile: false },
  { name: 'Ultrawide 34"', width: 3440, height: 1440, dpr: 1, mobile: false },
  { name: '43" 4K monitor', width: 3840, height: 2160, dpr: 1, mobile: false },
  { name: '72" TV 4K raw', width: 3840, height: 2160, dpr: 1, mobile: false },

  // In-between widths to verify no breakpoint collisions
  { name: 'In-between 1100', width: 1100, height: 800, dpr: 1, mobile: false },
  { name: 'In-between 1200', width: 1200, height: 800, dpr: 1, mobile: false },
  { name: 'In-between 1400', width: 1400, height: 900, dpr: 1, mobile: false },
  { name: 'In-between 1600', width: 1600, height: 900, dpr: 1, mobile: false },
  { name: 'In-between 1800', width: 1800, height: 1000, dpr: 1, mobile: false },
  { name: 'In-between 2000', width: 2000, height: 1100, dpr: 1, mobile: false },
  { name: 'In-between 3000', width: 3000, height: 1300, dpr: 1, mobile: false }
];

async function main() {
  const screenshotDir = path.join(__dirname, 'screenshots');
  if (!fs.existsSync(screenshotDir)) {
    fs.mkdirSync(screenshotDir, { recursive: true });
  }

  const chromePath = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
  const chromeProcess = spawn(chromePath, [
    '--headless=new',
    '--remote-debugging-port=9222',
    '--disable-gpu',
    '--no-first-run',
    '--no-default-browser-check',
    '--user-data-dir=' + path.join(__dirname, '.chrome-temp'),
    'about:blank'
  ]);

  // Wait for Chrome to be ready
  let ready = false;
  for (let i = 0; i < 30; i++) {
    try {
      const res = await fetch('http://127.0.0.1:9222/json/version');
      if (res.ok) {
        ready = true;
        break;
      }
    } catch (e) {}
    await new Promise(r => setTimeout(r, 200));
  }

  if (!ready) {
    throw new Error('Chrome failed to start on port 9222');
  }

  // Create or get page in Chrome
  let wsUrl;
  try {
    const newTargetRes = await fetch('http://127.0.0.1:9222/json/new?http://localhost:5000/', { method: 'PUT' });
    const target = await newTargetRes.json();
    wsUrl = target.webSocketDebuggerUrl;
  } catch (e) {
    const listRes = await fetch('http://127.0.0.1:9222/json/list');
    const list = await listRes.json();
    wsUrl = list[0].webSocketDebuggerUrl;
  }

  const ws = new WebSocket(wsUrl);
  let msgId = 0;
  const callbacks = new Map();
  const consoleMessages = [];

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.method === 'Runtime.consoleAPICalled') {
      const msg = data.params.type + ': ' + (data.params.args[0]?.value || '');
      if (!consoleMessages.includes(msg)) consoleMessages.push(msg);
    }
    if (data.id && callbacks.has(data.id)) {
      const cb = callbacks.get(data.id);
      callbacks.delete(data.id);
      cb(data);
    }
  };

  await new Promise((resolve) => { ws.onopen = resolve; });

  function send(method, params = {}) {
    return new Promise((resolve) => {
      const id = ++msgId;
      callbacks.set(id, resolve);
      ws.send(JSON.stringify({ id, method, params }));
    });
  }

  await send('Page.enable');
  await send('DOM.enable');
  await send('Runtime.enable');

  // Wait for initial render and fonts
  await new Promise(r => setTimeout(r, 2000));

  const results = [];

  for (const vp of viewports) {
    const filename = `${vp.width}x${vp.height}@${vp.dpr}.png`;
    const filepath = path.join(screenshotDir, filename);

    // Apply viewport emulation
    await send('Emulation.setDeviceMetricsOverride', {
      width: vp.width,
      height: vp.height,
      deviceScaleFactor: vp.dpr,
      mobile: vp.mobile || false
    });

    // Wait for layout reflow
    await new Promise(r => setTimeout(r, 400));

    // Evaluate layout checks
    const evalRes = await send('Runtime.evaluate', {
      expression: `(() => {
        const doc = document.documentElement;
        const body = document.body;
        const hero = document.querySelector('.hero');
        const container = document.querySelector('.section');
        const innerWidth = window.innerWidth;
        const scrollWidth = doc.scrollWidth;

        const allElements = Array.from(document.querySelectorAll('*'));
        const overflowing = allElements
          .filter(el => {
            const r = el.getBoundingClientRect();
            return r.right > innerWidth + 1;
          })
          .map(el => el.tagName + (el.className ? '.' + el.className.split(' ').join('.') : ''));

        return {
          innerWidth,
          scrollWidth,
          hasHorizontalOverflow: scrollWidth > innerWidth,
          overflowingElements: overflowing.slice(0, 3),
          heroHeight: hero ? hero.offsetHeight : 0,
          containerWidth: container ? container.offsetWidth : 0,
          bodyFontSize: getComputedStyle(body).fontSize
        };
      })()`,
      returnByValue: true
    });

    const check = evalRes?.result?.result?.value || evalRes?.result?.value || {};

    // Capture viewport screenshot
    const ssRes = await send('Page.captureScreenshot', {
      format: 'png',
      captureBeyondViewport: false
    });

    if (ssRes.result?.data) {
      fs.writeFileSync(filepath, Buffer.from(ssRes.result.data, 'base64'));
    }

    const pass = !check.hasHorizontalOverflow && (check.overflowingElements?.length === 0);

    results.push({
      device: vp.name,
      viewport: `${vp.width}x${vp.height}@${vp.dpr}`,
      scrollWidth: check.scrollWidth,
      innerWidth: check.innerWidth,
      hasHorizontalOverflow: check.hasHorizontalOverflow,
      containerWidth: check.containerWidth,
      heroHeight: check.heroHeight,
      bodyFontSize: check.bodyFontSize,
      status: pass ? 'PASS' : 'FAIL',
      file: filename
    });

    console.log(`[${pass ? 'PASS' : 'FAIL'}] ${vp.name.padEnd(25)} ${vp.width}x${vp.height}@${vp.dpr} | container: ${check.containerWidth}px | hero: ${check.heroHeight}px | overflow: ${check.hasHorizontalOverflow}`);
  }

  // Clean up
  try {
    await send('Page.close');
    ws.close();
  } catch (e) {}

  chromeProcess.kill();

  fs.writeFileSync(
    path.join(screenshotDir, 'test_results.json'),
    JSON.stringify({ results, consoleErrors: consoleMessages }, null, 2)
  );

  console.log(`\nAll ${results.length} viewports completed. Output in screenshots/test_results.json`);
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
