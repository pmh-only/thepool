function log (s, { error = false, success = false } = { error: false, success: false }) {
  document.getElementById('log').innerHTML += `
    <pre
      data-prefix=">"
      ${error ? 'class="text-error"' : ''}
      ${success ? 'class="text-success"' : ''}
    ><code class="ff-mono text-xl whitespace-nowrap pr-6">${s}</code></pre>`
}

document.getElementById('back').addEventListener('click', () => {
  window.close()
})

startAuthentication()
async function startAuthentication () {
  log('Discord authentication result received')

  const url = new URL(window.location.href)
  const code = url.searchParams.get('code')
  const bc = new BroadcastChannel("token_exchange_channel")

  if (!code) {
    log('Unauthorized access', { error: true })
    document.getElementById('failed').classList.remove("hidden")
    document.getElementById('back').classList.remove("hidden")
    document.getElementById('loading').classList.add("hidden")
    
    bc.postMessage({
      success: false
    })
    return
  }

  const res = await fetch('/api/token/' + code)
    .then((res) => res.json())

  if (!res.success) {
    log('Login failed: ' + res.message, { error: true })
    document.getElementById('failed').classList.remove("hidden")
    document.getElementById('back').classList.remove("hidden")
    document.getElementById('loading').classList.add("hidden")
    
    bc.postMessage({
      success: false
    })
    return
  }

  bc.postMessage({
    success: true,
    token: res.token
  })

  log('Login success! ', { success: true })
  document.getElementById('back').classList.remove("hidden")
  document.getElementById('loading').classList.add("hidden")
}
