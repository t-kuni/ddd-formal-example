// Run against a test app and Chrome with remote debugging enabled.
// BASE_URL and CHROME_DEBUG_URL can override the local defaults.
import assert from 'node:assert/strict';

const base = process.env.BASE_URL ?? 'http://127.0.0.1:8088';
const chrome = process.env.CHROME_DEBUG_URL ?? 'http://127.0.0.1:19222';
const tab = await (await fetch(chrome + '/json/new?' + encodeURIComponent(base), {method: 'PUT'})).json();
const socket = new WebSocket(tab.webSocketDebuggerUrl);
try {
  await new Promise((resolve, reject) => {
    socket.addEventListener('open', resolve, {once: true});
    socket.addEventListener('error', reject, {once: true});
  });
  const result = await new Promise((resolve, reject) => {
    const timeout = setTimeout(() => reject(new Error('Chrome test timed out')), 30000);
    socket.addEventListener('message', event => {
      const message = JSON.parse(event.data);
      if (message.id !== 1) return;
      clearTimeout(timeout);
      if (message.error || message.result.exceptionDetails) {
        reject(new Error(JSON.stringify(message.error ?? message.result.exceptionDetails)));
      } else {
        resolve(message.result.result.value);
      }
    });
    socket.send(JSON.stringify({id: 1, method: 'Runtime.evaluate', params: {
      awaitPromise: true, returnByValue: true, expression: '(' + verifyCartClicks + ')()',
    }}));
  });
  assert.deepEqual(result.actual, result.expected);
  console.log('PASS 連続クリック4回をカートに反映');
} finally {
  socket.close();
  await fetch(chrome + '/json/close/' + tab.id);
}

async function verifyCartClicks() {
  const sleep = ms => new Promise(resolve => setTimeout(resolve, ms));
  for (let i = 0; i < 80 && !document.querySelector('[data-add]'); i++) await sleep(100);

  const headers = {'X-Actor-ID': document.querySelector('#actor').value};
  const originalFetch = window.fetch;
  const catalog = async () => (await originalFetch('/api/catalog', {headers})).json();
  const {inventory, cart, now} = await catalog();
  const ids = inventory.filter(row => row.body.onHand - row.body.reserved >= 2)
    .slice(0, 3).map(row => row.id);
  if (ids.length !== 3) throw new Error('在庫のある検証用商品が3種類必要です');
  const before = cart?.body.expires > now ? cart.body.items : {};
  const expected = ids.map((id, index) => (before[id] || 0) + (index === 0 ? 2 : 1));
  const click = id => document.querySelector('[data-add="' + id + '"]').click();
  let release;
  const holdStatus = new Promise(resolve => { release = resolve; });
  let checkStarted;
  const checking = new Promise(resolve => { checkStarted = resolve; });
  window.fetch = async (path, options) => {
    if (String(path).startsWith('/api/commands?id=')) {
      checkStarted();
      await holdStatus;
    }
    return originalFetch(path, options);
  };

  click(ids[0]);
  await checking;
  click(ids[1]);
  click(ids[2]);
  click(ids[0]);
  release();

  let actual;
  for (let i = 0; i < 80; i++) {
    const state = await catalog();
    actual = ids.map(id => state.cart?.body.items[id] || 0);
    if (actual.every((quantity, index) => quantity === expected[index])) break;
    await sleep(150);
  }
  return {actual, expected};
}
