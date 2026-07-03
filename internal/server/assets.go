package server

// UI assets are kept as string constants (rather than an embedded FS) so the server
// package doesn't collide with the internal/embed package name and the whole binary
// stays a single self-contained artifact.

// dashboardHTML is the management dashboard (html/template). It renders the current
// configuration as editable JSON, lets the owner add knowledge, shows the embed
// snippet, and includes a live test-chat panel.
const dashboardHTML = `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.OfficerName}} — {{.BusinessName}} console</title>
<style>
 body{font:15px/1.5 system-ui,sans-serif;margin:0;background:#f6f7f9;color:#1a1a1a}
 header{background:#0f172a;color:#fff;padding:16px 24px}
 header b{font-size:18px} header span{opacity:.7;margin-left:8px;font-size:13px}
 main{max-width:920px;margin:24px auto;padding:0 16px;display:grid;gap:20px}
 .card{background:#fff;border:1px solid #e5e7eb;border-radius:10px;padding:18px}
 .card h2{margin:0 0 10px;font-size:15px;text-transform:uppercase;letter-spacing:.04em;color:#475569}
 textarea{width:100%;box-sizing:border-box;font:13px/1.45 ui-monospace,monospace;border:1px solid #cbd5e1;border-radius:8px;padding:10px}
 input,select{font:14px system-ui;border:1px solid #cbd5e1;border-radius:8px;padding:8px}
 button{background:#0f172a;color:#fff;border:0;border-radius:8px;padding:9px 16px;cursor:pointer;font:14px system-ui}
 button:hover{background:#1e293b}
 .row{display:flex;gap:8px;align-items:center;flex-wrap:wrap}
 .pill{display:inline-block;padding:2px 9px;border-radius:999px;font-size:12px;background:#e2e8f0;color:#334155}
 .ok{background:#dcfce7;color:#166534} .warn{background:#fef3c7;color:#92400e}
 code{background:#f1f5f9;padding:2px 6px;border-radius:5px}
 #log{max-height:280px;overflow:auto;display:flex;flex-direction:column;gap:8px;margin-bottom:10px}
 .msg{padding:8px 12px;border-radius:10px;max-width:80%}
 .me{align-self:flex-end;background:#0f172a;color:#fff}
 .bot{align-self:flex-start;background:#f1f5f9}
 .cite{font-size:12px;color:#64748b;margin-top:3px}
 .esc{border-left:3px solid #f59e0b;padding-left:8px}
 .note{color:#64748b;font-size:13px}
</style></head><body>
<header><b>{{.OfficerName}}</b><span>front-office console · {{.BusinessName}}</span></header>
<main>
 <div class="card"><h2>Status</h2>
  <div class="row">
   <span class="pill {{if .LLMReady}}ok{{else}}warn{{end}}">LLM: {{.LLMMode}}</span>
   <span class="pill">Embedder: {{.EmbedName}}</span>
   <span class="pill">Knowledge: {{.KBCount}} passage(s)</span>
   <span class="pill">Grounding ≥ {{.MinScore}}</span>
  </div>
  {{if not .LLMReady}}<p class="note">No LLM key detected — answers use retrieved text verbatim. Set the provider's API key (e.g. <code>GROQ_API_KEY</code>) and restart to enable natural drafting.</p>{{end}}
 </div>

 <div class="card"><h2>Test {{.OfficerName}}</h2>
  <div id="log"></div>
  <form id="chat" class="row"><input id="q" placeholder="Ask a question…" style="flex:1" autocomplete="off"><button>Send</button></form>
 </div>

 <div class="card"><h2>Add knowledge</h2>
  <form id="kb">
   <div class="row" style="margin-bottom:8px"><input id="src" placeholder="source label, e.g. kb:menu.md" style="flex:1"></div>
   <textarea id="txt" rows="3" placeholder="Paste a fact, FAQ answer, or policy passage…"></textarea>
   <div class="row" style="margin-top:8px"><button>Add & re-index</button><span id="kbmsg" class="note"></span></div>
  </form>
 </div>

 <div class="card"><h2>Configuration (JSON)</h2>
  <form id="cfg"><textarea id="cfgjson" rows="16">{{.ConfigJSON}}</textarea>
   <div class="row" style="margin-top:8px"><button>Save & re-index</button><span id="cfgmsg" class="note"></span></div></form>
 </div>

 <div class="card"><h2>Embed on your website</h2>
  <p class="note">Paste this one line before <code>&lt;/body&gt;</code>:</p>
  <textarea rows="2" readonly>&lt;script src="{{.Base}}/widget.js" defer&gt;&lt;/script&gt;</textarea>
  <p class="note">Standalone test page: <a href="/widget" target="_blank">{{.Base}}/widget</a></p>
 </div>
</main>
<script>
const log=document.getElementById('log');
function add(cls,html){const d=document.createElement('div');d.className='msg '+cls;d.innerHTML=html;log.appendChild(d);log.scrollTop=log.scrollHeight;}
fetch('/api/greeting').then(r=>r.json()).then(d=>add('bot',esc(d.greeting)));
document.getElementById('chat').onsubmit=async e=>{e.preventDefault();const q=document.getElementById('q');const t=q.value.trim();if(!t)return;add('me',esc(t));q.value='';
 const r=await fetch('/api/chat',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({message:t})});const d=await r.json();
 let h=esc(d.text);if(d.citations)d.citations.forEach(c=>h+='<div class="cite">↳ '+esc(c.source)+' ('+c.score.toFixed(2)+')</div>');
 if(d.escalation)h='<div class="esc">'+h+'<div class="cite">handoff → '+esc(d.escalation.contact||'—')+' ('+esc(d.escalation.reason)+')</div></div>';
 add('bot',h);};
document.getElementById('kb').onsubmit=async e=>{e.preventDefault();const src=document.getElementById('src').value,txt=document.getElementById('txt').value.trim();if(!txt)return;
 const r=await fetch('/admin/knowledge',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({source:src,text:txt})});
 document.getElementById('kbmsg').textContent=r.ok?'added — reload to refresh count':'error: '+await r.text();if(r.ok)document.getElementById('txt').value='';};
document.getElementById('cfg').onsubmit=async e=>{e.preventDefault();
 const r=await fetch('/admin/config',{method:'POST',headers:{'content-type':'application/json'},body:document.getElementById('cfgjson').value});
 document.getElementById('cfgmsg').textContent=r.ok?'saved & re-indexed':'error: '+await r.text();};
function esc(s){const d=document.createElement('div');d.textContent=s==null?'':s;return d.innerHTML;}
</script></body></html>`

// widgetJS is the embeddable chat widget (text/template; {{.Base}} is the server
// origin). It injects a floating bubble and panel and talks to /api/chat.
const widgetJS = `(function(){
 var BASE={{.Base}};
 var css=".alx-btn{position:fixed;right:20px;bottom:20px;width:56px;height:56px;border-radius:50%;background:#0f172a;color:#fff;border:0;font-size:24px;cursor:pointer;box-shadow:0 4px 14px rgba(0,0,0,.25);z-index:2147483000}"
 +".alx-panel{position:fixed;right:20px;bottom:88px;width:340px;max-width:92vw;height:460px;max-height:70vh;background:#fff;border-radius:14px;box-shadow:0 10px 40px rgba(0,0,0,.25);display:none;flex-direction:column;overflow:hidden;z-index:2147483000;font:14px/1.5 system-ui,sans-serif}"
 +".alx-hd{background:#0f172a;color:#fff;padding:12px 14px;font-weight:600}"
 +".alx-log{flex:1;overflow:auto;padding:12px;display:flex;flex-direction:column;gap:8px}"
 +".alx-m{padding:8px 11px;border-radius:12px;max-width:82%}.alx-me{align-self:flex-end;background:#0f172a;color:#fff}.alx-bot{align-self:flex-start;background:#f1f5f9;color:#111}"
 +".alx-cite{font-size:11px;color:#64748b;margin-top:2px}.alx-in{display:flex;border-top:1px solid #eee}.alx-in input{flex:1;border:0;padding:12px;font:14px system-ui;outline:none}.alx-in button{border:0;background:#0f172a;color:#fff;padding:0 16px;cursor:pointer}";
 var s=document.createElement('style');s.textContent=css;document.head.appendChild(s);
 var btn=document.createElement('button');btn.className='alx-btn';btn.textContent='💬';document.body.appendChild(btn);
 var p=document.createElement('div');p.className='alx-panel';
 p.innerHTML='<div class="alx-hd">Chat</div><div class="alx-log"></div><form class="alx-in"><input placeholder="Type a message…" autocomplete="off"><button>➤</button></form>';
 document.body.appendChild(p);
 var log=p.querySelector('.alx-log'),input=p.querySelector('input'),greeted=false;
 function esc(t){var d=document.createElement('div');d.textContent=t==null?'':t;return d.innerHTML;}
 function add(cls,html){var d=document.createElement('div');d.className='alx-m '+cls;d.innerHTML=html;log.appendChild(d);log.scrollTop=log.scrollHeight;}
 btn.onclick=function(){var open=p.style.display==='flex';p.style.display=open?'none':'flex';
  if(!open&&!greeted){greeted=true;fetch(BASE+'/api/greeting').then(function(r){return r.json();}).then(function(d){add('alx-bot',esc(d.greeting));});}};
 p.querySelector('form').onsubmit=function(e){e.preventDefault();var t=input.value.trim();if(!t)return;add('alx-me',esc(t));input.value='';
  fetch(BASE+'/api/chat',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({message:t})})
   .then(function(r){return r.json();}).then(function(d){var h=esc(d.text);
    if(d.citations)d.citations.forEach(function(c){h+='<div class="alx-cite">↳ '+esc(c.source)+'</div>';});
    if(d.escalation)h+='<div class="alx-cite">connecting you with a team member…</div>';add('alx-bot',h);})
   .catch(function(){add('alx-bot','Sorry, I had trouble reaching the server.');});};
})();`

// chatPageHTML is a standalone page that just loads the widget — a quick full-page test.
const chatPageHTML = `<!doctype html><html><head><meta charset="utf-8"><title>Chat test</title>
<style>body{font:16px system-ui;margin:0;display:grid;place-items:center;height:100vh;background:#eef2f7;color:#334155}</style></head>
<body><div>Test page — click the chat bubble in the corner.</div><script src="/widget.js" defer></script></body></html>`
