fetch('/healthz').then(response => response.json()).then(data => {
  document.querySelector('#status').textContent = `${data.service}: ${data.status}`;
}).catch(() => {
  document.querySelector('#status').textContent = 'Service unavailable';
});
