#!/usr/bin/env python3
"""Generate index.html for GitHub Pages from reports/ directory."""

import json
import os
from datetime import datetime, timezone
from pathlib import Path

REPORTS_DIR = Path(os.environ.get("REPORTS_DIR", "reports"))

MODULE_STYLE = {
    "eest": {"label": "EEST (EVM)", "color": "#1e40af", "bg": "#eff6ff"},
    "execution-apis": {"label": "RPC", "color": "#92400e", "bg": "#fffbeb"},
    "op-e2e": {"label": "op-e2e", "color": "#166534", "bg": "#f0fdf4"},
    "op-acceptance": {"label": "acceptance", "color": "#7e22ce", "bg": "#fdf4ff"},
    "proxyd": {"label": "proxyd", "color": "#0369a1", "bg": "#f0f9ff"},
}

def get_style(name):
    return MODULE_STYLE.get(name, {"label": name, "color": "#666", "bg": "#f5f5f5"})

def main():
    os.chdir(REPORTS_DIR)

    # Collect modules
    modules = {}
    for d in sorted(Path(".").iterdir()):
        if not d.is_dir() or d.name.startswith(".") or d.name == "plans":
            continue
        files = sorted(
            [f for f in d.iterdir() if f.is_file() and not f.name.startswith(".")],
            key=lambda f: f.name, reverse=True
        )
        if files:
            modules[d.name] = [f"{d.name}/{f.name}" for f in files]

    # Collect plans
    plans = {}
    plans_dir = Path("plans")
    if plans_dir.is_dir():
        for pf in plans_dir.iterdir():
            if pf.suffix == ".json":
                plans[pf.stem] = json.loads(pf.read_text())

    # Plans only come from plans/*.json — no auto-detection from filenames

    # Build sidebar HTML
    sidebar_items = []

    # Plans section
    if plans:
        items = ""
        for i, pname in enumerate(sorted(plans.keys(), reverse=True)):
            desc = plans[pname].get("description", pname)
            latest = ' <span class="tag">latest</span>' if i == 0 else ""
            items += f'<li><a href="#" onclick="showPlan(\'{pname}\')" data-search="{pname} {desc}">{pname}{latest}</a></li>\n'
        sidebar_items.append(("Test Plans", "#991b1b", "#fef2f2", items, len(plans)))

    # Module sections
    for mod_name, files in sorted(modules.items()):
        s = get_style(mod_name)
        items = ""
        for i, f in enumerate(files):
            fname = f.split("/")[-1]
            display = fname.replace(".html", "")
            latest = ' <span class="tag">latest</span>' if i == 0 else ""
            items += f'<li><a href="#" onclick="showReport(\'{f}\')" data-search="{mod_name} {fname}">{display}{latest}</a></li>\n'
        sidebar_items.append((s["label"], s["color"], s["bg"], items, len(files)))

    # Render
    sidebar_html = ""
    for i, (label, color, bg, items, count) in enumerate(sidebar_items):
        collapsed = "collapsed" if i > 0 else ""
        items_style = ' style="display:none"' if i > 0 else ""
        sidebar_html += f"""
        <div class="group">
          <div class="group-header {collapsed}" onclick="toggle(this)">
            <svg class="chevron" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"/></svg>
            <span class="badge" style="background:{bg};color:{color}">{label}</span>
            <span class="count">{count}</span>
          </div>
          <ul class="items"{items_style}>{items}</ul>
        </div>"""

    all_files_json = json.dumps(modules)
    now = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")

    html = f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Mantle Test Reports</title>
<style>
*{{margin:0;padding:0;box-sizing:border-box}}
html,body{{height:100%;overflow:hidden}}
body{{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;display:flex;background:#f5f5f5;color:#333}}

.sidebar{{width:300px;background:#fff;border-right:1px solid #e5e7eb;display:flex;flex-direction:column;height:100vh;overflow:hidden;position:relative;transition:width 0.2s;flex-shrink:0}}
.sidebar-header{{padding:14px 16px 10px;border-bottom:1px solid #f0f0f0;flex-shrink:0}}
.sidebar-header h1{{font-size:18px;font-weight:700}}
.search{{margin:10px 12px;padding:8px 12px;border:1px solid #ddd;border-radius:6px;font-size:14px;outline:none;width:calc(100% - 24px)}}
.search:focus{{border-color:#2563eb}}
.sidebar-content{{flex:1;overflow-y:auto;padding-bottom:40px}}

.group{{border-bottom:1px solid #f5f5f5}}
.group-header{{display:flex;align-items:center;padding:8px 12px;cursor:pointer;user-select:none}}
.group-header:hover{{background:#fafafa}}
.chevron{{width:16px;height:16px;margin-right:4px;transition:transform .15s;flex-shrink:0;color:#999}}
.group-header.collapsed .chevron{{transform:rotate(-90deg)}}
.badge{{padding:3px 10px;border-radius:10px;font-size:14px;font-weight:600}}
.count{{margin-left:auto;font-size:13px;color:#bbb}}

.items{{list-style:none;overflow:hidden;transition:max-height .2s ease}}
.group-header.collapsed+.items{{max-height:0!important;overflow:hidden}}
.items li{{padding:1px 0}}
.items a{{display:block;padding:6px 12px 6px 36px;font-size:14px;color:#555;text-decoration:none;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}}
.items a:hover{{background:#f8f9fa;color:#2563eb}}
.items a.active{{background:#eff6ff;color:#1e40af;font-weight:600}}
.tag{{display:inline-block;padding:1px 6px;border-radius:3px;font-size:10px;font-weight:700;background:#dcfce7;color:#166534;vertical-align:middle;margin-left:4px}}
.hidden{{display:none!important}}

.sidebar-footer{{padding:8px 12px;border-top:1px solid #f0f0f0;font-size:10px;color:#ccc;flex-shrink:0}}
.sidebar-toggle{{position:absolute;top:50%;right:-14px;width:28px;height:28px;background:#fff;border:1px solid #ddd;border-radius:50%;cursor:pointer;display:flex;align-items:center;justify-content:center;font-size:12px;color:#999;z-index:10;box-shadow:0 1px 3px rgba(0,0,0,0.1)}}
.sidebar-toggle:hover{{background:#f5f5f5;color:#333}}
.sidebar.collapsed-sidebar{{width:0;min-width:0;overflow:hidden;border-right:none}}
.sidebar.collapsed-sidebar .sidebar-header,.sidebar.collapsed-sidebar .search,.sidebar.collapsed-sidebar .sidebar-content,.sidebar.collapsed-sidebar .sidebar-footer{{display:none}}
.sidebar.collapsed-sidebar .sidebar-toggle{{right:-36px}}
.resize-handle{{position:absolute;top:0;right:0;width:4px;height:100%;cursor:col-resize;background:transparent}}
.resize-handle:hover{{background:#2563eb40}}

.main{{flex:1;display:flex;flex-direction:column;height:100vh;overflow:hidden}}
.toolbar{{padding:6px 16px;background:#fff;border-bottom:1px solid #e5e7eb;font-size:12px;color:#888;flex-shrink:0;display:flex;align-items:center;gap:8px}}
.toolbar .path{{color:#333;font-weight:600}}
.viewer{{flex:1;border:none;width:100%}}
.empty{{flex:1;display:flex;align-items:center;justify-content:center;color:#ccc;font-size:13px}}
</style>
</head>
<body>

<div class="sidebar" id="sidebar">
  <div class="sidebar-toggle" id="sidebarToggle" onclick="toggleSidebar()">◀</div>
  <div class="resize-handle" id="resizeHandle"></div>
  <div class="sidebar-header"><h1>Mantle Test Reports</h1></div>
  <input class="search" id="search" placeholder="Search..." type="text">
  <div class="sidebar-content">
    {sidebar_html}
  </div>
  <div class="sidebar-footer">Updated: {now}</div>
</div>

<div class="main">
  <div class="toolbar">
    <span>Report:</span>
    <span class="path" id="path">Select a report</span>
  </div>
  <div class="empty" id="empty">Select a report from the sidebar</div>
  <iframe class="viewer" id="viewer" style="display:none"></iframe>
</div>

<script>
const allFiles={all_files_json};

function toggleSidebar(){{
  const sb=document.getElementById('sidebar');
  const btn=document.getElementById('sidebarToggle');
  sb.classList.toggle('collapsed-sidebar');
  btn.textContent=sb.classList.contains('collapsed-sidebar')?'▶':'◀';
}}

// Drag to resize sidebar
(function(){{
  const handle=document.getElementById('resizeHandle');
  const sb=document.getElementById('sidebar');
  let startX,startW;
  handle.addEventListener('mousedown',function(e){{
    startX=e.clientX;
    startW=sb.offsetWidth;
    document.addEventListener('mousemove',onDrag);
    document.addEventListener('mouseup',stopDrag);
    e.preventDefault();
  }});
  function onDrag(e){{
    const w=Math.max(200,Math.min(600,startW+(e.clientX-startX)));
    sb.style.width=w+'px';
  }}
  function stopDrag(){{
    document.removeEventListener('mousemove',onDrag);
    document.removeEventListener('mouseup',stopDrag);
  }}
}})();

function toggle(el){{
  el.classList.toggle('collapsed');
  const list=el.nextElementSibling;
  if(list)list.style.display=el.classList.contains('collapsed')?'none':'block';
}}

function showReport(href){{
  document.querySelectorAll('.items a').forEach(a=>a.classList.remove('active'));
  if(event&&event.target)event.target.classList.add('active');
  document.getElementById('viewer').src=href;
  document.getElementById('viewer').style.display='block';
  document.getElementById('empty').style.display='none';
  document.getElementById('path').textContent=href;
  history.replaceState(null,'','#'+href);
}}

// Load from hash on page load (supports direct linking / refresh)
window.addEventListener('load',function(){{
  const hash=location.hash.slice(1);
  if(hash&&!hash.startsWith('plan:')){{
    showReport(hash);
  }}else if(hash.startsWith('plan:')){{
    showPlan(hash.slice(5));
  }}
}});

function showPlan(name){{
  const rows=[];
  Object.keys(allFiles).sort().forEach(mod=>{{
    allFiles[mod].forEach(f=>{{
      const fname=f.split('/').pop();
      if(fname.startsWith(name+'-')){{
        rows.push('<tr style="cursor:pointer" data-href="'+f+'"><td style="padding:8px 12px;border-bottom:1px solid #f0f0f0;font-weight:600">'+mod+'</td><td style="padding:8px 12px;border-bottom:1px solid #f0f0f0;color:#2563eb">'+fname+'</td></tr>');
      }}
    }});
  }});
  if(rows.length===0){{
    rows.push('<tr><td colspan="2" style="padding:12px;color:#999;text-align:center">No reports found for this plan</td></tr>');
  }}
  const html='<html><body style="font-family:-apple-system,sans-serif;padding:20px"><h2 style="font-size:16px;margin-bottom:12px">Plan: '+name+'</h2><p style="color:#888;font-size:12px;margin-bottom:12px">Click a row to view the report</p><table style="border-collapse:collapse;width:100%"><tr style="background:#f8f9fa"><th style="text-align:left;padding:8px 12px;border-bottom:2px solid #e5e7eb;font-size:12px">Module</th><th style="text-align:left;padding:8px 12px;border-bottom:2px solid #e5e7eb;font-size:12px">Report</th></tr>'+rows.join('')+'</table></body></html>';
  const blob=new Blob([html],{{type:'text/html'}});
  const frame=document.getElementById('viewer');
  frame.src=URL.createObjectURL(blob);
  frame.style.display='block';
  document.getElementById('empty').style.display='none';
  document.getElementById('path').textContent='Plan: '+name;
  history.replaceState(null,'','#plan:'+name);
  // Attach click handlers to plan rows after iframe loads
  frame.onload=function(){{
    frame.contentDocument.querySelectorAll('tr[data-href]').forEach(function(tr){{
      tr.addEventListener('click',function(){{
        parent.showReport(tr.dataset.href);
      }});
    }});
  }};
}}

document.getElementById('search').addEventListener('input',function(){{
  const q=this.value.toLowerCase();
  document.querySelectorAll('.group').forEach(g=>{{
    let vis=false;
    g.querySelectorAll('.items a').forEach(a=>{{
      const match=(a.dataset.search||'').toLowerCase().includes(q);
      a.parentElement.classList.toggle('hidden',!match);
      if(match)vis=true;
    }});
    g.classList.toggle('hidden',!vis);
    if(vis&&q)g.querySelector('.group-header').classList.remove('collapsed');
  }});
}});
</script>
</body>
</html>"""

    Path("index.html").write_text(html)
    print(f"Generated index.html ({len(modules)} modules, {len(plans)} plans)")

if __name__ == "__main__":
    main()
