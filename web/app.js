const $ = (id) => document.getElementById(id);
let projects=[], activeProject=null, config={}, currentAssets=[], knowledge={entities:[],claims:[],relations:[]};
const token=()=>localStorage.getItem('nexo_access_token')||'';
async function api(path,opts={}){opts.headers={...(opts.headers||{})};if(token())opts.headers.Authorization=`Bearer ${token()}`;if(opts.body&&!(opts.body instanceof FormData))opts.headers['Content-Type']='application/json';let r=await fetch(path,opts);if(r.status===401){const t=prompt('Este espacio requiere el token de acceso NEXO:');if(t){localStorage.setItem('nexo_access_token',t);return api(path,opts)}}if(!r.ok){let e={};try{e=await r.json()}catch{};throw new Error(e.error||`HTTP ${r.status}`)}if(r.headers.get('content-type')?.includes('text/csv'))return r.blob();return r.json()}
function toast(msg){$('toast').textContent=msg;$('toast').classList.add('show');setTimeout(()=>$('toast').classList.remove('show'),2600)}
function fmt(sec){sec=Math.max(0,Math.floor(sec||0));return `${String(Math.floor(sec/3600)).padStart(2,'0')}:${String(Math.floor(sec%3600/60)).padStart(2,'0')}:${String(sec%60).padStart(2,'0')}`}
function escapeHTML(v){return String(v||'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]))}
async function init(){try{config=await api('/api/config');$('planBadge').textContent=config.pro?'PRO':'FREE';$('upgradeBtn').hidden=config.pro;$('paymentLink').hidden=!config.payment_url;await loadProjects()}catch(e){toast(e.message)}}
async function loadProjects(){const d=await api('/api/projects');projects=d.projects||[];renderProjects();if(activeProject){const p=projects.find(x=>x.id===activeProject.id);if(p)await openProject(p);else activeProject=null}if(!activeProject){$('emptyState').hidden=false;$('projectView').hidden=true}}
function renderProjects(){$('projectList').innerHTML='';projects.forEach(p=>{const b=document.createElement('button');b.className='project-item'+(activeProject?.id===p.id?' active':'');b.innerHTML=`<span>${escapeHTML(p.name)}</span>`;b.onclick=()=>openProject(p);$('projectList').append(b)})}
async function openProject(p){
 activeProject=p;renderProjects();
 const d=await api('/api/projects/'+encodeURIComponent(p.id));
 if(activeProject?.id!==p.id)return;
 $('emptyState').hidden=true;$('projectView').hidden=false;
 $('projectTitle').textContent=p.name;
 $('projectDescription').textContent=(p.description||'')+(p.core_status==='connected'?' · PrismaTEC-Core conectado':p.core_status==='not_configured'?' · Core no configurado':'');
 $('exportBtn').hidden=!config.pro;$('results').innerHTML='';
 applyProjectAssets(d.assets||[]);
 await loadKnowledge();await loadEvents();
}
function applyProjectAssets(assets){
 currentAssets=Array.isArray(assets)?assets:[];
 renderAssets(currentAssets);
 $('mediaCount').textContent=`${currentAssets.length} fuentes`;
 $('assetCount').textContent=currentAssets.length;
 $('evidenceCount').textContent=currentAssets.reduce((n,a)=>n+(a.segments?.length||0),0);
 populateKnowledgeForms();
}
async function refreshProjectAssets(projectID){
 const d=await api('/api/projects/'+encodeURIComponent(projectID));
 if(activeProject?.id!==projectID)return;
 applyProjectAssets(d.assets||[]);
 return currentAssets;
}
async function watchTranscription(projectID,assetID,attempts=40){
 for(let i=0;i<attempts;i++){
  await new Promise(resolve=>setTimeout(resolve,1500));
  if(activeProject?.id!==projectID)return;
  const asset=await api('/api/assets/'+encodeURIComponent(assetID));
  await refreshProjectAssets(projectID);
  if(!['awaiting_transcript','transcribing'].includes(asset.analysis_status))return;
 }
 toast('La transcripción sigue en proceso. Puedes continuar trabajando; actualiza el espacio más tarde.');
}
function renderAssets(assets){$('assetList').innerHTML='';if(!assets.length){$('assetList').innerHTML='<div class="muted">Todavía no hay contenidos. Añade un vídeo, una clase o una entrevista.</div>';return}assets.forEach(a=>{const el=document.createElement('div');el.className='asset-card';el.innerHTML=`<strong>${escapeHTML(a.title)}</strong><small>${escapeHTML(a.filename)} · ${(a.size_bytes/1048576).toFixed(1)} MB</small><small>${a.segments?.length||0} evidencias · ${escapeHTML(a.analysis_status)}</small><button class="text-btn" data-id="${a.id}">Abrir fuente ↗</button>`;el.querySelector('button').onclick=()=>openPlayer(a,0);$('assetList').append(el)})}
function openProjectDialog(){$('projectDialog').showModal()}
$('newProjectBtn').onclick=openProjectDialog;$('emptyCreateBtn').onclick=openProjectDialog;
$('projectForm').addEventListener('submit',async e=>{e.preventDefault();const f=new FormData(e.currentTarget);try{const p=await api('/api/projects',{method:'POST',body:JSON.stringify({name:f.get('name'),description:f.get('description')})});$('projectDialog').close();e.currentTarget.reset();await loadProjects();await openProject(p);toast(p.core_status==='connected'?'Espacio creado y organismo registrado en PrismaTEC-Core':'Espacio creado')}catch(err){toast(err.message)}});
$('exportBtn').onclick=async()=>{try{const blob=await api('/api/export.csv?project_id='+encodeURIComponent(activeProject.id));const url=URL.createObjectURL(blob);const a=document.createElement('a');a.href=url;a.download='nexo-evidencias.csv';a.click();URL.revokeObjectURL(url)}catch(e){toast(e.message)}};$('addMediaBtn').onclick=()=>$('uploadDialog').showModal();$('closeUpload').onclick=$('cancelUpload').onclick=()=>$('uploadDialog').close();
$('uploadForm').addEventListener('submit',async e=>{
 e.preventDefault();if(!activeProject)return;
 const projectID=activeProject.id,form=new FormData(e.currentTarget);
 const transcript=String(form.get('transcript')||'').trim();form.delete('transcript');
 const submit=e.currentTarget.querySelector('button[type="submit"]');
 if(submit){submit.disabled=true;submit.textContent='Guardando…'}
 try{
  const a=await api(`/api/projects/${encodeURIComponent(projectID)}/assets`,{method:'POST',body:form});
  if(activeProject?.id===projectID){applyProjectAssets([a,...currentAssets.filter(x=>x.id!==a.id)])}
  $('uploadDialog').close();e.currentTarget.reset();
  if(transcript){
   try{await api(`/api/assets/${encodeURIComponent(a.id)}/transcript`,{method:'POST',body:JSON.stringify({text:transcript})})}
   catch(err){await refreshProjectAssets(projectID);toast('La fuente se guardó, pero falló la transcripción: '+err.message);return}
  }
  await refreshProjectAssets(projectID);
  await loadKnowledge();await loadEvents();
  toast('Fuente añadida a la memoria');
  if(!transcript&&config.transcription_configured)watchTranscription(projectID,a.id).catch(err=>toast('Error actualizando la transcripción: '+err.message));
 }catch(err){toast(err.message)}
 finally{if(submit){submit.disabled=false;submit.textContent='Guardar fuente'}}
});
$('searchBtn').onclick=runSearch;$('askBtn').onclick=askNexo;$('searchInput').addEventListener('keydown',e=>{if(e.key==='Enter')runSearch()});
async function askNexo(){if(!activeProject)return;const question=$('searchInput').value.trim();if(!question){toast('Escribe una pregunta primero');return}try{const d=await api('/api/ask',{method:'POST',body:JSON.stringify({project_id:activeProject.id,question})});const el=document.createElement('section');el.className='result-card';el.innerHTML=`<div class="result-top"><span class="result-source">RESPUESTA NEXO · ${escapeHTML(d.mode)}</span><span class="timestamp">CON EVIDENCIA</span></div><p>${escapeHTML(d.answer).replace(/\n/g,'<br>')}</p><div class="muted">${escapeHTML(d.warning||'')}</div>`;$('results').prepend(el);toast('Respuesta generada; revisa las evidencias citadas')}catch(e){toast(e.message)}}
async function runSearch(){
 if(!activeProject){toast('Selecciona un espacio de conocimiento');return}
 const q=$('searchInput').value.trim();
 if(!q){toast('Escribe una palabra o frase para buscar');return}
 try{
  const d=await api('/api/search',{method:'POST',body:JSON.stringify({project_id:activeProject.id,query:q,limit:40})});
  const results=Array.isArray(d.results)?d.results:[];
  $('results').innerHTML='';
  if(!results.length){
   const hasEvidence=currentAssets.some(a=>Array.isArray(a.segments)&&a.segments.length>0);
   $('results').innerHTML=`<div class="muted">${hasEvidence?'No encontramos coincidencias literales. Prueba con otras palabras presentes en la transcripción.':'Esta fuente todavía no tiene evidencias indexadas. Añade una transcripción temporizada o configura la transcripción automática.'}</div>`;
   toast('Búsqueda completada: 0 coincidencias');return;
  }
  results.forEach(r=>{
   const el=document.createElement('article');el.className='result-card';
   el.innerHTML=`<div class="result-top"><span class="result-source">${escapeHTML(r.asset_title)}</span><span class="timestamp">${r.timestamp_source==='estimated'?'≈':''}${fmt(r.start)} — ${fmt(r.end)}</span></div><p>${escapeHTML(r.text)}</p><div class="result-actions"><button class="text-btn">▶ Abrir en este momento</button><span class="muted">Evidencia ${escapeHTML(r.evidence_id)}</span></div>`;
   el.querySelector('button').onclick=()=>openPlayer({id:r.asset_id,title:r.asset_title},r.start,r.text);
   $('results').append(el);
  });
  toast(`${results.length} evidencias encontradas`);
 }catch(e){toast('Error en la búsqueda: '+e.message)}
}
async function openPlayer(a,sec,evidence=''){try{const full=await api('/api/assets/'+a.id);const media=await api('/api/assets/'+a.id+'/media-url');$('playerTitle').textContent=full.title;$('player').src=media.url;$('player').onloadedmetadata=()=>{$('player').currentTime=sec||0};$('playerEvidence').textContent=evidence?`Evidencia: ${evidence}`:`Fuente: ${full.filename} · enlace temporal`; $('playerDialog').showModal()}catch(e){toast(e.message)}}
$('closePlayer').onclick=()=>{$('player').pause();$('player').removeAttribute('src');$('player').load();$('playerDialog').close()};$('upgradeBtn').onclick=()=>$('licenseDialog').showModal();
$('licenseForm').addEventListener('submit',async e=>{e.preventDefault();const key=new FormData(e.currentTarget).get('key');try{await api('/api/license/activate',{method:'POST',body:JSON.stringify({key})});$('licenseDialog').close();await init();toast('Plan Pro activado')}catch(err){toast(err.message)}});
$('paymentLink').onclick=()=>{if(config.payment_url)window.open(config.payment_url,'_blank','noopener')};
init();


// NEXO 0.4: propuestas de acontecimientos con revisión humana.
let projectEvents=[];
async function loadEvents(){if(!activeProject)return;try{const d=await api(`/api/projects/${activeProject.id}/events`);projectEvents=d.events||[];renderEvents()}catch(e){toast(`No se pudieron cargar los acontecimientos: ${e.message}`)}}
function renderEvents(){
 $('eventCount').textContent=`${projectEvents.length} acontecimientos`;
 const list=$('eventList');list.innerHTML='';
 if(!projectEvents.length){list.innerHTML='<div class="knowledge-empty">No hay propuestas. Puedes generarlas desde las transcripciones disponibles.</div>';return}
 projectEvents.forEach(ev=>{const card=document.createElement('article');card.className='knowledge-card';const status=ev.status==='pending'?'Pendiente de revisión':ev.status==='accepted'?'Revisado · aceptado':'Revisado · rechazado';
 card.innerHTML=`<strong>${escapeHTML(ev.title)}</strong><p>${escapeHTML(ev.description||'')}</p><div class="knowledge-meta"><span>${escapeHTML(status)}</span><span>${fmt(ev.start_seconds)} — ${fmt(ev.end_seconds)}</span><span>${escapeHTML(ev.extractor||'manual')}</span></div>${evidenceButton(ev.evidence_ids)}${ev.status==='pending'?'<div class="knowledge-toolbar"><button class="btn btn-quiet" data-review="accept">Aceptar</button><button class="btn btn-quiet" data-review="reject">Rechazar</button></div>':''}${ev.review_note?`<p class="knowledge-meta">Revisión: ${escapeHTML(ev.review_note)}</p>`:''}`;
 card.querySelectorAll('[data-review]').forEach(b=>b.onclick=()=>reviewEvent(ev,b.dataset.review));list.append(card);
 });bindEvidenceButtons(list);
}
async function reviewEvent(ev,decision){const note=prompt(decision==='accept'?'Nota de revisión (opcional):':'Motivo del rechazo (opcional):')||'';try{await api(`/api/projects/${activeProject.id}/events/${ev.id}/review`,{method:'POST',body:JSON.stringify({decision,review_note:note})});await loadEvents();toast(decision==='accept'?'Acontecimiento aceptado':'Propuesta rechazada; se conserva la auditoría')}catch(e){toast(e.message)}}
$('extractEventsBtn').onclick=async()=>{if(!activeProject)return;try{const d=await api(`/api/projects/${activeProject.id}/events/extract`,{method:'POST',body:JSON.stringify({})});await loadEvents();toast(`${d.created} propuestas nuevas; ${d.skipped} segmentos omitidos`)}catch(e){toast(e.message)}};

// NEXO 0.3: interfaz de memoria estructurada con procedencia audiovisual.
function evidenceOptions(){
 const out=[];
 currentAssets.forEach(a=>(a.segments||[]).forEach(seg=>out.push({id:seg.id,label:`${a.title} · ${fmt(seg.start)} · ${(seg.text||'').slice(0,100)}`})));
 return out;
}
function fillSelect(select,items,placeholder='Selecciona…'){
 if(!select)return;select.innerHTML='';
 if(!select.multiple){const o=document.createElement('option');o.value='';o.textContent=placeholder;select.append(o)}
 items.forEach(item=>{const o=document.createElement('option');o.value=item.id;o.textContent=item.label;select.append(o)});
}
function selectedValues(select){return [...select.selectedOptions].map(o=>o.value).filter(Boolean)}
function evidenceButton(ids){
 const id=ids?.[0];if(!id)return '<span class="knowledge-meta">Sin evidencia</span>';
 return `<button type="button" class="knowledge-evidence" data-evidence="${escapeHTML(id)}">▶ Abrir evidencia original</button>`;
}
function bindEvidenceButtons(root){root.querySelectorAll('[data-evidence]').forEach(b=>b.addEventListener('click',()=>{
 const evidenceID=b.dataset.evidence;
 for(const asset of currentAssets){const seg=(asset.segments||[]).find(x=>x.id===evidenceID);if(seg){openPlayer(asset,seg.start,seg.text);return}}
 toast('La evidencia ya no está disponible en esta fuente');
}));}
async function loadKnowledge(){
 if(!activeProject)return;
 try{knowledge=await api(`/api/projects/${activeProject.id}/knowledge`);renderKnowledge();populateKnowledgeForms()}
 catch(e){toast(`No se pudo cargar la memoria estructurada: ${e.message}`)}
}
function renderKnowledge(){
 const entities=knowledge.entities||[],claims=knowledge.claims||[],relations=knowledge.relations||[];
 $('knowledgeCount').textContent=`${entities.length} entidades · ${claims.length} afirmaciones · ${relations.length} relaciones`;
 const entityNames=new Map(entities.map(e=>[e.id,e.name]));
 const renderEmpty=(msg)=>`<div class="knowledge-empty">${msg}</div>`;
 $('entityList').innerHTML=entities.length?entities.map(e=>`<article class="knowledge-card"><strong>${escapeHTML(e.name)}</strong><p>${escapeHTML(e.description||e.type)}</p><div class="knowledge-meta"><span>${escapeHTML(e.type)}</span><span>Confianza ${Math.round((e.confidence||0)*100)}%</span></div>${evidenceButton(e.evidence_ids)}</article>`).join(''):renderEmpty('Todavía no hay entidades registradas.');
 $('claimList').innerHTML=claims.length?claims.map(c=>`<article class="knowledge-card"><strong>${escapeHTML(c.text)}</strong><p>${(c.entity_ids||[]).map(id=>escapeHTML(entityNames.get(id)||id)).join(' · ')}</p><div class="knowledge-meta"><span class="state-${c.zyrion_state}">Zyrion ${c.zyrion_state} · ${c.zyrion_state===1?'True':c.zyrion_state===0?'False':'Indeterminado'}</span><span>Confianza ${Math.round((c.confidence||0)*100)}%</span></div>${evidenceButton(c.evidence_ids)}</article>`).join(''):renderEmpty('Todavía no hay afirmaciones registradas.');
 $('relationList').innerHTML=relations.length?relations.map(r=>`<article class="knowledge-card"><strong>${escapeHTML(entityNames.get(r.subject_entity_id)||r.subject_entity_id)} <span class="muted">${escapeHTML(r.predicate)}</span> ${escapeHTML(entityNames.get(r.object_entity_id)||r.object_entity_id)}</strong><div class="knowledge-meta"><span>Confianza ${Math.round((r.confidence||0)*100)}%</span></div>${evidenceButton(r.evidence_ids)}</article>`).join(''):renderEmpty('Todavía no hay relaciones registradas.');
 bindEvidenceButtons(document.getElementById('entityList'));bindEvidenceButtons(document.getElementById('claimList'));bindEvidenceButtons(document.getElementById('relationList'));
}
function populateKnowledgeForms(){
 const evidence=evidenceOptions().map(e=>({id:e.id,label:e.label}));
 document.querySelectorAll('.evidence-select').forEach(s=>fillSelect(s,evidence,'Selecciona la evidencia…'));
 const entities=(knowledge.entities||[]).map(e=>({id:e.id,label:`${e.name} (${e.type})`}));
 fillSelect($('claimEntities'),entities,'');fillSelect($('relationForm').elements.subject_entity_id,entities,'Entidad de origen…');fillSelect($('relationForm').elements.object_entity_id,entities,'Entidad de destino…');
 const disabled=entities.length<2 || evidence.length===0;$('addRelationBtn').disabled=disabled;$('addClaimBtn').disabled=evidence.length===0;$('addEntityBtn').disabled=evidence.length===0;
}
function openKnowledgeDialog(id){if(!activeProject){toast('Selecciona un espacio primero');return}if(!evidenceOptions().length){toast('Añade una fuente con segmentos de evidencia antes de registrar conocimiento');return}if(id==='relationDialog'&&(knowledge.entities||[]).length<2){toast('Necesitas al menos dos entidades registradas');return}$(id).showModal()}
$('addEntityBtn').onclick=()=>openKnowledgeDialog('entityDialog');$('addClaimBtn').onclick=()=>openKnowledgeDialog('claimDialog');$('addRelationBtn').onclick=()=>openKnowledgeDialog('relationDialog');
document.querySelectorAll('[data-close]').forEach(b=>b.addEventListener('click',()=>$(b.dataset.close).close()));
$('entityForm').addEventListener('submit',async e=>{e.preventDefault();const f=new FormData(e.currentTarget);try{await api(`/api/projects/${activeProject.id}/knowledge/entities`,{method:'POST',body:JSON.stringify({name:f.get('name'),type:f.get('type'),description:f.get('description'),evidence_ids:[f.get('evidence_ids')],confidence:Number(f.get('confidence'))})});$('entityDialog').close();e.currentTarget.reset();await loadKnowledge();toast('Entidad guardada con su evidencia')}catch(err){toast(err.message)}});
$('claimForm').addEventListener('submit',async e=>{e.preventDefault();const f=new FormData(e.currentTarget);try{await api(`/api/projects/${activeProject.id}/knowledge/claims`,{method:'POST',body:JSON.stringify({text:f.get('text'),kind:f.get('kind'),zyrion_state:Number(f.get('zyrion_state')),entity_ids:selectedValues($('claimEntities')),evidence_ids:[f.get('evidence_ids')],confidence:Number(f.get('confidence'))})});$('claimDialog').close();e.currentTarget.reset();$('claimForm').elements.zyrion_state.value='2';$('claimForm').elements.kind.value='statement';$('claimForm').elements.confidence.value='0.8';await loadKnowledge();toast('Afirmación guardada como conocimiento trazable')}catch(err){toast(err.message)}});
$('relationForm').addEventListener('submit',async e=>{e.preventDefault();const f=new FormData(e.currentTarget);try{await api(`/api/projects/${activeProject.id}/knowledge/relations`,{method:'POST',body:JSON.stringify({subject_entity_id:f.get('subject_entity_id'),predicate:f.get('predicate'),object_entity_id:f.get('object_entity_id'),evidence_ids:[f.get('evidence_ids')],confidence:Number(f.get('confidence'))})});$('relationDialog').close();e.currentTarget.reset();await loadKnowledge();toast('Relación guardada con su evidencia')}catch(err){toast(err.message)}});
