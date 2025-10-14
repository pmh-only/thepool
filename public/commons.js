function log (s, { error = false, success = false } = { error: false, success: false }) {
  document.getElementById('log').innerHTML += `
    <pre
      data-prefix=">"
      ${error ? 'class="text-error"' : ''}
      ${success ? 'class="text-success"' : ''}
    ><code class="ff-mono text-xl whitespace-nowrap pr-6">${s}</code></pre>`
}

function setEnable (enabled = true) {
  if (document.getElementById('file'))
    document.getElementById('file').disabled = !enabled

  document.getElementById('parallel').disabled = !enabled
  document.getElementById('submit').disabled = !enabled
}

async function loadConfig () { 
  return await fetch('/api/config')
    .then((r) => r.json())
}

function formatBytes(n){
  if (n < 1024)
    return `${n} B`

  const u = ['KB','MB','GB','TB']
  let i = -1

  do {
    n /= 1024
    i++
  } while (n >= 1024 && i < u.length - 1)
    
  return `${n.toFixed(n >= 100 ? 0 : n >= 10 ? 1 : 2)} ${u[i]}`
}

// ---

const submit = document.getElementById('submit')
const parallelEl = document.getElementById('parallel')
const parallelCountEl = document.getElementById('parallel_count')
const panel = document.getElementById('panel')

const originalFetch = window.fetch

window.fetch = (url, option) =>
  originalFetch(url, {
    ...option,
    headers: {
      ...(option?.headers ?? {}),
      Authorization: 'Bearer ' + window.sessionStorage.getItem('SESSION_TOKEN')
    }
  })
