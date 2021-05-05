async function start() {
  const config = {
    iceServers: [
      {
        urls: 'stun:stun.l.google.com:19302'
      }
    ]
  };
  const api = new GoLiveWebRTC('http://localhost:8080', config);
  api.conn.addEventListener('track', e => {
    const el = document.getElementById('live')
    el.srcObject = e.streams[0]
  });
  api.conn.addTransceiver('video', { 'direction': 'recvonly' });
  api.conn.addTransceiver('audio', { 'direction': 'recvonly' });
  await api.connect();
}
