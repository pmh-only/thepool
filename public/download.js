const MB = 1024 * 1024

parallelEl.addEventListener('input', () => {
  parallelCountEl.textContent = parallelEl.value
  submit.innerText = parallelEl.value > 1 ? 'Parallel Download' : 'Download'
})

loadCollectionConfig()
async function loadCollectionConfig () {
  const collectionId = window.location.pathname.replace('/', '')
  document.getElementById('collection').innerText = collectionId


  log(`Fetching collection: ${collectionId}`)
  const meta = await fetch(`/api/collections/${encodeURIComponent(collectionId)}`)
    .then((res) => res.json())
    .catch(() => undefined)

  if (meta === undefined || !meta.success) {
    const alert = document.createElement('div')
    alert.innerHTML = `
      <div role="alert" class="alert alert-error alert-soft">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 shrink-0 stroke-current" fill="none" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <span><b>Oops.</b> Collection not found or some file chunks have expired.</span>
      </div>
    `

    panel.prepend(alert)
    return
  }

  const name = meta.data.originalName
  const chunks = meta.data.chunks
  const total = chunks.length

  log(`File: ${name} • ${total} chunk(s)`)

  const box = document.createElement('div')
  box.innerHTML = `
    <div class="flex gap-6 items-center">
      <div id="fprog" class="radial-progress" style="--value: 0;" role="progressbar">0%</div>
      <div class="flex-1 grow">
          <p><strong>${name}</strong> • ${total} chunk(s)</p>
          <p id="ftext">0% • 0 / ${formatBytes(meta.data.totalSize)}</p>
      </div>
    </div>
    <ul id="rows" class="list">
      <li class="p-4 pb-2 text-xs opacity-60 tracking-wide">File chunks</li>
    </ul>
  `

  panel.prepend(box)
  const fprog = box.querySelector('#fprog')
  const ftext = box.querySelector('#ftext')
  const rows  = box.querySelector('#rows')

  const row = chunks.map((chunk, i) => {
    const li = document.createElement('li')
    li.innerHTML = `
      <div class="text-4xl font-thin opacity-30 tabular-nums">${(i+1).toString().padStart(2, '0')}</div>
      <div class="list-col-grow">
        <div><span>Chunk #${i+1} (${chunk.id}): 0 / ${formatBytes(chunk.size)} (0%)</span></div>
        <progress class="progress w-full" max="${chunk.size}" value="0"></progress>
      </div>
    `

    rows.appendChild(li)
    li.classList.add('list-row')

    return {
      prog: li.querySelector('progress'),
      txt: li.querySelector('span')
    }
  })

  document
    .getElementById('form')
    .addEventListener('submit', async (e) => {
      e.preventDefault()

      setEnable(false)
      
      await downloadCollectionParallel(collectionId, meta, fprog, ftext, rows, row)
        .catch((err) => {
          log(`Error for file: ${err?.message ?? err}`, { error: true })
        })

      log('All done.')
      setEnable(true)
    })

  setEnable(true)
}

async function downloadCollectionParallel (collectionId, meta, fprog, ftext, rows, row) {
  const workers = +parallelEl.value
  const name = meta.data.originalName
  const mime = meta.data.mimeType
  const chunks = meta.data.chunks
  const total = chunks.length

  const slices = name.split('.')
  const exts = '.' + slices[slices.length - 1]

  const saveHandle = await window.showSaveFilePicker({
    suggestedName: name,
    types: [{
      accept: {
        [mime]: [exts ??'.bin']
      }
    }]
  }).catch(() => undefined)

  if (saveHandle === undefined) {
    setEnable(false)
    log('Picker canceled.')
    return
  }

  const root = await navigator.storage.getDirectory()
  const tmpDirName = `dl-${collectionId}-${Date.now()}`
  const tmpDir = await root.getDirectoryHandle(tmpDirName, { create: true })

  let next = 0
  const tempFiles = new Array(total)

  let fileDownloaded = 0
  let lastReport = 0
  const REPORT_MS = 500

  function maybeReport () {
    const now = performance.now()

    if (now - lastReport >= REPORT_MS){
      ftext.textContent = `${formatBytes(fileDownloaded)} downloaded`
      lastReport = now
    }
  }

  const { download } = await loadConfig()

  function bumpFile (delta) {
    fileDownloaded += delta
    const pct = Math.floor(fileDownloaded / meta.data.totalSize * 100)
    fprog.style.cssText = `--value: ${pct}`
    fprog.textContent = `${pct}%`
    ftext.textContent = `${pct}% • ${formatBytes(fileDownloaded)} / ${formatBytes(meta.data.totalSize)}`
  }

  async function worker () {
    while (true) {
      const i = next++
      if (i >= total) return

      const chunk = chunks[i]

      log(`start ${i+1}/${total} ${chunk.id}...`)

      const res = await fetch(`${download}/${encodeURIComponent(chunk.id)}`)
      if(!res.ok) throw new Error(`Chunk ${chunk.id} missing or expired`)

      const fh = await tmpDir.getFileHandle(`part-${String(i).padStart(8,'0')}`, { create: true })
      const ws = await fh.createWritable()
      const reader = res.body.getReader()
      let chunkDownloaded = 0

      while (true) {
        const { done, value } = await reader.read()

        if (done) break
        
        await ws.write(value)
        
        chunkDownloaded += value.byteLength
        
        const pct = Math.floor(chunkDownloaded / chunk.size * 100)

        row[i].prog.value = chunkDownloaded
        row[i].txt.textContent = `Chunk #${i+1} (${chunk.id}): ${formatBytes(chunkDownloaded)} / ${formatBytes(chunk.size)} (${pct}%)`
        
        bumpFile(value.byteLength)
        maybeReport()
      }

      await ws.close()

      tempFiles[i] = fh
      log(`done  ${i+1}/${total}`)
    }
  }

  const workerSet = Array.from({ length: Math.min(workers, total) }, worker)
  await Promise.all(workerSet)

  
  const li = document.createElement('li')
  li.innerHTML = `
    <div class="text-4xl font-thin opacity-30 tabular-nums">FF</div>
    <div class="list-col-grow">
      <div><span>Merge chunks</span></div>
      <progress class="progress w-full" ></progress>
    </div>
  `

  rows.appendChild(li)
  li.classList.add('list-row')

  const out = await saveHandle.createWritable()

  for (const tempFile of tempFiles) {
    const file = await tempFile.getFile()
    const reader = file.stream().getReader()

    while (true) {
      const { done, value } = await reader.read()

      if (done) break
      await out.write(value)
    }

    await tmpDir.removeEntry(tempFile.name)
  }

  await out.close();
  await root.removeEntry(tmpDirName, { recursive: true })

  rows.removeChild(li)

  ftext.textContent = `${formatBytes(fileDownloaded)} downloaded`
  log(`File complete: ${name}`)
}
