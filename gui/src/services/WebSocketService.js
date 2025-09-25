export default class WebSocketService {
  constructor(url) {
    this.url = url;
    this.socket = null;
  }

  connect() {
    // No-op stub for production build
    return Promise.resolve();
  }

  disconnect() {
    if (this.socket) {
      try { this.socket.close(); } catch (_) {}
    }
  }
}
