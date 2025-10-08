const id = window.location.pathname.replace('/', '')

parallelEl.addEventListener('input', () => {
  parallelCountEl.textContent = parallelEl.value
  submit.innerText = parallelEl.value > 1 ? 'Parallel Download' : 'Download'
})

document.getElementById('collection').value = id
document
  .getElementById('form')
  .addEventListener('submit', async (e) => {
    e.preventDefault()

    if (!id) {
      log('No collection id.', { error: true })
      return
    }

    setEnable(false)
    
    await downloadCollectionParallel(id, parallelEl.value)
      .catch((err) => {
        log(`Error for file: ${err?.message ?? err}`, { error: true })
      })

    log('All done.')
    setEnable(true)
  })

async function downloadCollectionParallel (collectionId, workers) {
  log(`Fetching collection: ${collectionId}`)
  const meta = await fetch(`/api/collections/${encodeURIComponent(collectionId)}`)
    .then((res) => res.json())

  const name = meta.data.originalName ?? `download-${collectionId}`
  const mime = meta.data.mimeType ?? 'application/octet-stream'
  const chunkIds = meta.data.chunkIds.match(/.{1,10}/g)
  const total = chunkIds.length

  log(`File: ${name} • ${total} chunk(s)`)

  let saveHandle;
  try{
    saveHandle = await window.showSaveFilePicker({
      suggestedName: name,
      types: [{ accept: { [mime]: ['.'+name.split('.')[name.split('.').length - 1] ??'.bin'] } }],
    });
  }catch(err){
    console.log(err)
    submit.disabled = false;
    log('Picker canceled.');
    return;
  }

  const box = document.createElement('div')
  box.innerHTML = `
    <div tabindex="0" class="collapse collapse-arrow bg-base-100 border-base-300 border">
      <div class="collapse-title font-semibold flex gap-6">
        <div>
          <p><strong>${name}</strong> • ${total} chunk(s)</p>
          <p id="ftext">0 downloaded</p>
        </div>
      </div>
      <div class="collapse-content text-sm">
        <div id="rows" class="flex flex-wrap gap-6"></div>
      </div>
    </div>
  `

  panel.prepend(box)
  const ftext = box.querySelector('#ftext')
  const rows  = box.querySelector('#rows')

  const row = chunkIds.map((cid, i) => {
    const div = document.createElement('div')
    div.innerHTML = `
      <div>
        <span>Chunk #${i+1} (${cid}...): 0 B</span>
      </div>
    `

    rows.appendChild(div)

    return {
      txt: div.querySelector('span')
    }
  })

  const root = await navigator.storage.getDirectory()
  const tmpDirName = `dl-${collectionId}-${Date.now()}`
  const tmpDir = await root.getDirectoryHandle(tmpDirName, { create: true })

  let next = 0
  const tempFiles = new Array(total)

  let downloaded = 0
  let lastReport = 0
  const REPORT_MS = 500

  function maybeReport () {
    const now = performance.now()

    if (now - lastReport >= REPORT_MS){
      ftext.textContent = `${formatBytes(downloaded)} downloaded`
      lastReport = now
    }
  }

  async function worker () {
    while(true){
      const i = next++;
      if(i >= total) return;
      const cid = chunkIds[i];
      log(`start ${i+1}/${total} ${cid.slice(0,8)}…`);

      const { download } = await loadConfig()
      const res = await fetch(`${download}/${encodeURIComponent(cid)}`);
      if(!res.ok) throw new Error(`Chunk ${cid} missing or expired`);

      const fh = await tmpDir.getFileHandle(`part-${String(i).padStart(8,'0')}`, { create: true });
      const ws = await fh.createWritable();

      const reader = res.body.getReader();
      let loaded = 0;
      while(true){
        const {done, value} = await reader.read();
        if(done) break;
        await ws.write(value);
        loaded += value.byteLength;
        downloaded += value.byteLength;
        row[i].txt.textContent = `Chunk #${i+1} (${cid.slice(0,8)}…): ${formatBytes(loaded)}`;
        maybeReport();
      }
      await ws.close();
      tempFiles[i] = fh;
      log(`done  ${i+1}/${total}`);
    }
  }

  await Promise.all(Array.from({length: Math.min(workers, total)}, worker))

  const out = await saveHandle.createWritable()

  for (let i=0; i < total; i++) {
    const file = await tempFiles[i].getFile()
    const reader = file.stream().getReader()
    while(true){
      const {done, value} = await reader.read()
      if(done) break;
      await out.write(value);
    }
    await tmpDir.removeEntry(tempFiles[i].name)
  }

  await out.close();
  await root.removeEntry(tmpDirName, { recursive: true })

  ftext.textContent = `${formatBytes(downloaded)} downloaded`
  log(`File complete: ${name}`)
}
