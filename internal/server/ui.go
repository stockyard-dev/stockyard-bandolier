package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashHTML))
}

const dashHTML = `<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Bandolier</title>
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}.hdr h1 span{color:var(--rust)}
.main{padding:1.5rem;max-width:960px;margin:0 auto}
.stats{display:grid;grid-template-columns:repeat(3,1fr);gap:.5rem;margin-bottom:1rem}
.st{background:var(--bg2);border:1px solid var(--bg3);padding:.6rem;text-align:center}
.st-v{font-size:1.2rem;font-weight:700}.st-l{font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.15rem}
.toolbar{display:flex;gap:.5rem;margin-bottom:1rem;align-items:center;flex-wrap:wrap}
.search{flex:1;min-width:180px;padding:.4rem .6rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.search:focus{outline:none;border-color:var(--leather)}
.filter-sel{padding:.4rem .5rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.65rem}
.secret{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;margin-bottom:.5rem;transition:border-color .2s}
.secret:hover{border-color:var(--leather)}
.secret-top{display:flex;justify-content:space-between;align-items:flex-start;gap:.5rem}
.secret-key{font-size:.8rem;font-weight:700;color:var(--gold)}
.secret-val{font-size:.65rem;color:var(--cm);margin-top:.2rem;font-family:var(--mono);background:var(--bg);padding:.2rem .4rem;border:1px solid var(--bg3);cursor:pointer;position:relative}
.secret-val.masked{color:var(--bg3);user-select:none}
.secret-meta{font-size:.55rem;color:var(--cm);margin-top:.3rem;display:flex;gap:.5rem;flex-wrap:wrap;align-items:center}
.secret-actions{display:flex;gap:.3rem;flex-shrink:0}
.env-badge{font-size:.5rem;padding:.12rem .35rem;text-transform:uppercase;letter-spacing:1px;border:1px solid var(--green);color:var(--green)}
.proj-badge{font-size:.5rem;padding:.1rem .3rem;background:var(--bg3);color:var(--cd)}
.sensitive-icon{color:var(--red);font-size:.6rem}
.btn{font-size:.6rem;padding:.25rem .5rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd);transition:all .2s}
.btn:hover{border-color:var(--leather);color:var(--cream)}.btn-p{background:var(--rust);border-color:var(--rust);color:#fff}
.btn-sm{font-size:.55rem;padding:.2rem .4rem}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.65);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:460px;max-width:92vw}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust);letter-spacing:1px}
.fr{margin-bottom:.6rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.2rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.4rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.fr input:focus,.fr select:focus{outline:none;border-color:var(--leather)}
.row2{display:grid;grid-template-columns:1fr 1fr;gap:.5rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:1rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1><span>&#9670;</span> BANDOLIER</h1><button class="btn btn-p" onclick="openForm()">+ Add Secret</button></div>
<div class="main">
<div class="stats" id="stats"></div>
<div class="toolbar">
<input class="search" id="search" placeholder="Search keys, projects..." oninput="render()">
<select class="filter-sel" id="env-filter" onchange="render()"><option value="">All Environments</option></select>
</div>
<div id="secrets"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)closeModal()"><div class="modal" id="mdl"></div></div>
<script>
var A='/api',items=[],editId=null,revealed={};

async function load(){var r=await fetch(A+'/env_vars').then(function(r){return r.json()});items=r.env_vars||[];renderStats();buildEnvFilter();render();}

function renderStats(){
var total=items.length;
var envs={};items.forEach(function(s){if(s.environment)envs[s.environment]=true});
var sensitive=items.filter(function(s){return s.sensitive}).length;
document.getElementById('stats').innerHTML=[
{l:'Secrets',v:total},{l:'Environments',v:Object.keys(envs).length},{l:'Sensitive',v:sensitive,c:sensitive>0?'var(--red)':''}
].map(function(x){return '<div class="st"><div class="st-v" style="'+(x.c?'color:'+x.c:'')+'">'+x.v+'</div><div class="st-l">'+x.l+'</div></div>'}).join('');
}

function buildEnvFilter(){
var envs={};items.forEach(function(s){if(s.environment)envs[s.environment]=true});
var sel=document.getElementById('env-filter');var cur=sel.value;
sel.innerHTML='<option value="">All Environments</option>';
Object.keys(envs).sort().forEach(function(e){sel.innerHTML+='<option value="'+esc(e)+'"'+(cur===e?' selected':'')+'>'+esc(e)+'</option>';});
}

function render(){
var q=(document.getElementById('search').value||'').toLowerCase();
var ef=document.getElementById('env-filter').value;
var f=items;
if(ef)f=f.filter(function(s){return s.environment===ef});
if(q)f=f.filter(function(s){return(s.key||'').toLowerCase().includes(q)||(s.project||'').toLowerCase().includes(q)||(s.description||'').toLowerCase().includes(q)});
if(!f.length){document.getElementById('secrets').innerHTML='<div class="empty">No secrets stored.</div>';return;}
var h='';f.forEach(function(s){
var masked=s.sensitive&&!revealed[s.id];
h+='<div class="secret"><div class="secret-top"><div style="flex:1">';
h+='<div class="secret-key">'+(s.sensitive?'<span class="sensitive-icon">&#9679; </span>':'')+esc(s.key)+'</div>';
if(s.value)h+='<div class="secret-val'+(masked?' masked':'')+'" onclick="toggleReveal(''+s.id+'')">'+(masked?'&#8226;&#8226;&#8226;&#8226;&#8226;&#8226;&#8226;&#8226;':esc(s.value))+'</div>';
h+='</div><div class="secret-actions">';
h+='<button class="btn btn-sm" onclick="openEdit(''+s.id+'')">Edit</button>';
h+='<button class="btn btn-sm" onclick="del(''+s.id+'')" style="color:var(--red)">&#10005;</button>';
h+='</div></div>';
h+='<div class="secret-meta">';
if(s.environment)h+='<span class="env-badge">'+esc(s.environment)+'</span>';
if(s.project)h+='<span class="proj-badge">'+esc(s.project)+'</span>';
if(s.description)h+='<span>'+esc(s.description)+'</span>';
h+='</div></div>';
});
document.getElementById('secrets').innerHTML=h;
}

function toggleReveal(id){revealed[id]=!revealed[id];render();}
async function del(id){if(!confirm('Delete?'))return;await fetch(A+'/env_vars/'+id,{method:'DELETE'});load();}

function formHTML(secret){
var i=secret||{key:'',value:'',environment:'',project:'',sensitive:0,description:''};
var isEdit=!!secret;
var h='<h2>'+(isEdit?'EDIT SECRET':'ADD SECRET')+'</h2>';
h+='<div class="fr"><label>Key *</label><input id="f-key" value="'+esc(i.key)+'" placeholder="DATABASE_URL"></div>';
h+='<div class="fr"><label>Value</label><input id="f-value" value="'+esc(i.value)+'" placeholder="secret value"></div>';
h+='<div class="row2"><div class="fr"><label>Environment</label><input id="f-env" value="'+esc(i.environment)+'" placeholder="production"></div>';
h+='<div class="fr"><label>Project</label><input id="f-proj" value="'+esc(i.project)+'"></div></div>';
h+='<div class="fr"><label>Description</label><input id="f-desc" value="'+esc(i.description)+'"></div>';
h+='<div class="fr" style="display:flex;align-items:center;gap:.5rem"><input type="checkbox" id="f-sens" '+(i.sensitive?'checked':'')+'><label style="margin:0">Sensitive</label></div>';
h+='<div class="acts"><button class="btn" onclick="closeModal()">Cancel</button><button class="btn btn-p" onclick="submit()">'+(isEdit?'Save':'Add')+'</button></div>';
return h;
}

function openForm(){editId=null;document.getElementById('mdl').innerHTML=formHTML();document.getElementById('mbg').classList.add('open');}
function openEdit(id){var s=null;for(var j=0;j<items.length;j++){if(items[j].id===id){s=items[j];break;}}if(!s)return;editId=id;document.getElementById('mdl').innerHTML=formHTML(s);document.getElementById('mbg').classList.add('open');}
function closeModal(){document.getElementById('mbg').classList.remove('open');editId=null;}

async function submit(){
var key=document.getElementById('f-key').value.trim();
if(!key){alert('Key is required');return;}
var body={key:key,value:document.getElementById('f-value').value,environment:document.getElementById('f-env').value.trim(),project:document.getElementById('f-proj').value.trim(),description:document.getElementById('f-desc').value.trim(),sensitive:document.getElementById('f-sens').checked?1:0};
if(editId){await fetch(A+'/env_vars/'+editId,{method:'PUT',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
else{await fetch(A+'/env_vars',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(body)});}
closeModal();load();
}

function esc(s){if(!s)return'';var d=document.createElement('div');d.textContent=s;return d.innerHTML;}
document.addEventListener('keydown',function(e){if(e.key==='Escape')closeModal();});
load();
</script></body></html>`
