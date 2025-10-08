parallelEl.addEventListener('input', () => {
  parallelCountEl.textContent = parallelEl.value
  submit.innerText = parallelEl.value > 1 ? 'Parallel Upload' : 'Upload'
})


document
  .getElementById('form')
  .addEventListener('submit', async (e) => {
    e.preventDefault()

    const files = document.getElementById('file').files
    if (!files || files.length === 0) {
      log('No files.', { error: true })
      return
    }

    setEnable(false)

    const config = await loadConfig()
    const workers = parallelEl.value
    const chunkSize = (config.chunkSize ?? 5) * 1024 * 1024

    for (const file of files){
      await uploadFileParallelStreaming(file, chunkSize, workers)
        .catch((err) => {
          log(`Error for file ${file.name}: ${err?.message ?? err}`, { error: true })
        })
    }
    
    log('All done.')
    setEnable(true)
  })

function planChunks (size, chunkSize) {
  const arr = []

  let start = 0
  let idx = 0

  while (start < size) {
    const end = Math.min(start + chunkSize, size)

    arr.push({
      index:idx++,
      start,
      end,
      size: end - start
    })

    start = end
  }

  return arr
}

async function uploadFileParallelStreaming (file, chunkSize, workers) {
  log(`File: ${file.name} (${formatBytes(file.size)})`)

  const tasks = planChunks(file.size, Math.min(file.size / workers, chunkSize))
  const total = tasks.length

  const box = document.createElement('div')
  box.innerHTML = `
    <div tabindex="0" class="collapse collapse-arrow bg-base-100 border-base-300 border">
      <div class="collapse-title font-semibold flex gap-6">
        <div id="fprog" class="radial-progress" style="--value:0;" role="progressbar">0%</div>
        <div>
          <p><strong>${file.name}</strong> • ${total} chunk(s)</p>
          <p id="ftext">0% • 0 / ${formatBytes(file.size)}</p>
          <p id="flink"><span class="loading loading-dots loading-xs"></span></p>

          <button class="btn" id="fbtn" disabled>Copy Link<button>
        </div>
      </div>
      <div class="collapse-content text-sm">
        <div id="rows" class="flex flex-wrap gap-6"></div>
      </div>
    </div>
  `

  panel.prepend(box)
  const fprog = box.querySelector('#fprog')
  const ftext = box.querySelector('#ftext')
  const flink = box.querySelector('#flink')
  const fbtn = box.querySelector('#fbtn')
  const rows  = box.querySelector('#rows')

  const row = tasks.map((t,i) => {
    const div = document.createElement('div')
    div.innerHTML = `
      <div>
        <span>Chunk #${i+1}: 0 / ${formatBytes(t.size)} (0%)</span>
        <progress class="progress w-56" max="${t.size}" value="0"></progress>
      </div>
    `
    rows.appendChild(div)

    return {
      prog: div.querySelector('progress'),
      txt: div.querySelector('span')
    }
  })

  const results = new Array(total)
  let fileUploaded = 0
  let next = 0

  function bumpFile (delta) {
    fileUploaded += delta
    const pct = Math.floor(fileUploaded / file.size * 100)
    fprog.style.cssText = `--value: ${pct}`
    fprog.textContent = `${pct}%`
    ftext.textContent = `${pct}% • ${formatBytes(fileUploaded)} / ${formatBytes(file.size)}`
  }

  async function worker () {
    while (true) {
      const i = next++
      if (i >= total) return
      const t = tasks[i]
      const blob = file.slice(t.start, t.end)

      let loaded = 0
      const metered = blob.stream().pipeThrough(new TransformStream({
        transform(chunk, ctl){
          loaded += chunk.byteLength
          row[i].prog.value = loaded
          const pct = Math.floor(loaded / t.size * 100)
          row[i].txt.textContent = `Chunk #${i+1}: ${formatBytes(loaded)} / ${formatBytes(t.size)} (${pct}%)`
          bumpFile(chunk.byteLength)
          ctl.enqueue(chunk)
        }
      }))

      const res = await fetch('/api/chunks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/octet-stream'
        },
        body: metered,
        duplex: 'half'
      })

      if (!res.ok) throw new Error(`HTTP ${res.status}`)
      const json = await res.json().catch(()=>({}))
      results[i] = json.id || null
    }
  }

  await Promise.all(Array.from({length: Math.min(workers, total)}, worker))

  const collections = await fetch('/api/collections', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      collectionId: '',
      originalName: file.name,
      mimeType: file.type,
      chunkIds: results.join('')
    })
  }).then((res) => res.json())

  const link = `${window.location.protocol}//${window.location.host}/${collections.id}`

  log(`File complete: ${file.name}`, { success: true })
  log(`<a class="link link-hover" href="${link}">${link}</a>`)

  flink.innerText = link
  fbtn.disabled = false
  fbtn.onclick = async () => {
    await navigator.clipboard.writeText(link)
    fbtn.innerText = 'Copied!'
  }
}
