import "@testing-library/jest-dom/vitest";

// jsdom doesn't implement crypto.randomUUID or WebSocket; the pieces of the
// app under test only need a stable UUID and a no-op socket constructor.
if (!globalThis.crypto.randomUUID) {
  globalThis.crypto.randomUUID = () => "00000000-0000-4000-8000-000000000000";
}

class MockWebSocket {
  constructor() {
    this.readyState = 0;
  }
  close() {}
  send() {}
}
globalThis.WebSocket = globalThis.WebSocket || MockWebSocket;
