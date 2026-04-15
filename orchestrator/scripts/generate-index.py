#!/usr/bin/env python3
"""Generate index.html for GitHub Pages from reports/ directory.

Layout supported (all co-exist; empty envs are hidden):
  reports/<module>/<env>/<file>.html   — environment-scoped reports (preferred)
  reports/<module>/<file>.html         — legacy flat reports (shown as "(root)")
  reports/plans/<name>.json            — plan metadata (drives Test Plans sidebar)

The sidebar renders as a 3-level tree:
  Test Plans → [plan names]
  <module>   → <env> → [report files]
"""

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

# Env labels get a small color tag in the sidebar so qa/sepolia/mainnet are visually distinct.
ENV_STYLE = {
    "qa":      {"color": "#0f766e", "bg": "#ccfbf1"},
    "qa3":     {"color": "#0f766e", "bg": "#ccfbf1"},
    "sepolia": {"color": "#a16207", "bg": "#fef3c7"},
    "mainnet": {"color": "#991b1b", "bg": "#fee2e2"},
    "localchain": {"color": "#4338ca", "bg": "#e0e7ff"},
}
ENV_DEFAULT = {"color": "#4b5563", "bg": "#f3f4f6"}
LEGACY_ENV_LABEL = "(root)"  # files directly under reports/<module>/


def get_module_style(name):
    return MODULE_STYLE.get(name, {"label": name, "color": "#666", "bg": "#f5f5f5"})


def get_env_style(name):
    return ENV_STYLE.get(name, ENV_DEFAULT)


def collect_modules():
    """Return {module: {env_or_empty: [rel_path, ...]}} sorted by filename desc."""
    modules = {}
    for mod_dir in sorted(Path(".").iterdir()):
        if not mod_dir.is_dir() or mod_dir.name.startswith(".") or mod_dir.name == "plans":
            continue

        env_map = {}
        for child in sorted(mod_dir.iterdir()):
            if child.name.startswith("."):
                continue
            if child.is_dir():
                # Environment subdir: reports/<module>/<env>/<file>.html
                files = sorted(
                    [f for f in child.iterdir() if f.is_file() and not f.name.startswith(".")],
                    key=lambda f: f.name, reverse=True,
                )
                if files:
                    env_map[child.name] = [f"{mod_dir.name}/{child.name}/{f.name}" for f in files]
            elif child.is_file():
                # Legacy flat file: reports/<module>/<file>.html
                env_map.setdefault("", []).append(f"{mod_dir.name}/{child.name}")

        # Sort legacy bucket filenames desc
        if "" in env_map:
            env_map[""] = sorted(env_map[""], key=lambda p: p.split("/")[-1], reverse=True)

        if env_map:
            modules[mod_dir.name] = env_map
    return modules


def collect_plans():
    plans = {}
    plans_dir = Path("plans")
    if plans_dir.is_dir():
        for pf in plans_dir.iterdir():
            if pf.suffix == ".json":
                try:
                    plans[pf.stem] = json.loads(pf.read_text())
                except json.JSONDecodeError:
                    pass
    return plans


def render_env_block(env, files, indent_extra=0):
    """Render a single env sub-group under a module."""
    est = get_env_style(env) if env else ENV_DEFAULT
    env_label = env if env else LEGACY_ENV_LABEL
    items = ""
    for i, f in enumerate(files):
        fname = f.split("/")[-1]
        display = fname.rsplit(".", 1)[0]
        latest = ' <span class="tag">latest</span>' if i == 0 else ""
        items += (
            f'<li><a href="javascript:void(0)" '
            f'onclick="showReport(\'{f}\')" '
            f'data-search="{env} {fname}">{display}{latest}</a></li>\n'
        )
    return (
        f'<div class="env-group">'
        f'  <div class="env-header" onclick="toggleEnv(this)">'
        f'    <svg class="chevron" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"/></svg>'
        f'    <span class="env-badge" style="background:{est["bg"]};color:{est["color"]}">{env_label}</span>'
        f'    <span class="count">{len(files)}</span>'
        f'  </div>'
        f'  <ul class="env-items">{items}</ul>'
        f'</div>'
    )


def main():
    os.chdir(REPORTS_DIR)

    modules = collect_modules()
    plans = collect_plans()

    sidebar_items = []

    # Plans section
    if plans:
        items = ""
        for i, pname in enumerate(sorted(plans.keys(), reverse=True)):
            p = plans[pname]
            desc = p.get("description") or pname
            env_meta = p.get("environment") or ""
            latest = ' <span class="tag">latest</span>' if i == 0 else ""
            env_tag = f' <span class="env-chip">{env_meta}</span>' if env_meta else ""
            items += (
                f'<li><a href="javascript:void(0)" onclick="showPlan(\'{pname}\')" '
                f'data-search="{pname} {desc} {env_meta}">{pname}{env_tag}{latest}</a></li>\n'
            )
        sidebar_items.append(("Test Plans", "#991b1b", "#fef2f2", items, len(plans), True))

    # Module sections (each module has env sub-groups)
    for mod_name in sorted(modules.keys()):
        env_map = modules[mod_name]
        s = get_module_style(mod_name)
        total = sum(len(v) for v in env_map.values())
        # Order envs: known envs first in insertion order, legacy "" bucket last.
        env_order = sorted([e for e in env_map if e], key=lambda e: (e != "qa", e != "qa3", e))
        if "" in env_map:
            env_order.append("")
        inner_html = "".join(render_env_block(e, env_map[e]) for e in env_order)
        sidebar_items.append((s["label"], s["color"], s["bg"], inner_html, total, False))

    # Build sidebar HTML
    sidebar_html = ""
    for i, (label, color, bg, inner, count, is_plan_group) in enumerate(sidebar_items):
        collapsed = "collapsed" if i > 0 else ""
        inner_style = ' style="display:none"' if i > 0 else ""
        # For plans group: inner is <li> list; wrap in <ul>. For modules: inner is env-group blocks.
        if is_plan_group:
            inner_markup = f'<ul class="items"{inner_style}>{inner}</ul>'
        else:
            inner_markup = f'<div class="env-groups"{inner_style}>{inner}</div>'
        sidebar_html += f"""
        <div class="group">
          <div class="group-header {collapsed}" onclick="toggle(this)">
            <svg class="chevron" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M5.23 7.21a.75.75 0 011.06.02L10 11.168l3.71-3.938a.75.75 0 111.08 1.04l-4.25 4.5a.75.75 0 01-1.08 0l-4.25-4.5a.75.75 0 01.02-1.06z"/></svg>
            <span class="badge" style="background:{bg};color:{color}">{label}</span>
            <span class="count">{count}</span>
          </div>
          {inner_markup}
        </div>"""

    # Flatten for client-side plan filtering: {module: {env: [files]}}
    all_files_json = json.dumps(modules)
    now = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
    total_plans = len(plans)
    total_modules = len(modules)
    total_reports = sum(sum(len(v) for v in em.values()) for em in modules.values())

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

.sidebar{{width:320px;background:#fff;border-right:1px solid #e5e7eb;display:flex;flex-direction:column;height:100vh;overflow:hidden;position:relative;transition:width 0.2s;flex-shrink:0}}
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
.group-header.collapsed+.items,.group-header.collapsed+.env-groups{{max-height:0!important;display:none!important;overflow:hidden}}
.items li{{padding:1px 0}}
.items a{{display:block;padding:6px 12px 6px 36px;font-size:14px;color:#555;text-decoration:none;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}}
.items a:hover{{background:#f8f9fa;color:#2563eb}}
.items a.active,.env-items a.active{{background:#dbeafe;color:#1e40af;font-weight:600;border-left:3px solid #2563eb;padding-left:33px}}
.tag{{display:inline-block;padding:1px 6px;border-radius:3px;font-size:10px;font-weight:700;background:#dcfce7;color:#166534;vertical-align:middle;margin-left:4px}}
.env-chip{{display:inline-block;padding:1px 6px;border-radius:3px;font-size:10px;font-weight:600;background:#e0f2fe;color:#0369a1;vertical-align:middle;margin-left:4px}}
.hidden{{display:none!important}}

/* Env sub-groups inside a module */
.env-groups{{overflow:hidden;transition:max-height .2s ease}}
.env-group{{border-top:1px solid #fafafa}}
.env-header{{display:flex;align-items:center;padding:5px 12px 5px 32px;cursor:pointer;user-select:none;font-size:13px}}
.env-header:hover{{background:#fafafa}}
.env-header.collapsed .chevron{{transform:rotate(-90deg)}}
.env-badge{{padding:2px 8px;border-radius:8px;font-size:12px;font-weight:600}}
.env-items{{list-style:none;overflow:hidden;transition:max-height .2s ease}}
.env-header.collapsed+.env-items{{max-height:0!important;display:none!important;overflow:hidden}}
.env-items li{{padding:1px 0}}
.env-items a{{display:block;padding:5px 12px 5px 56px;font-size:13px;color:#555;text-decoration:none;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}}
.env-items a:hover{{background:#f8f9fa;color:#2563eb}}

.sidebar-footer{{padding:8px 12px;border-top:1px solid #f0f0f0;font-size:10px;color:#ccc;flex-shrink:0}}
.sidebar.collapsed-sidebar{{width:40px;min-width:40px;overflow:hidden}}
.sidebar.collapsed-sidebar .sidebar-header,.sidebar.collapsed-sidebar .search,.sidebar.collapsed-sidebar .sidebar-content,.sidebar.collapsed-sidebar .resize-handle{{display:none}}
.sidebar-collapse-btn{{width:100%;padding:10px 0;border:none;border-top:1px solid #f0f0f0;background:#fff;cursor:pointer;font-size:12px;color:#999;flex-shrink:0}}
.sidebar-collapse-btn:hover{{background:#f5f5f5;color:#333}}
.sidebar-expand-btn{{display:none;width:100%;height:100%;border:none;background:#fff;cursor:pointer;font-size:18px;color:#999;flex:1}}
.sidebar-expand-btn:hover{{background:#f5f5f5;color:#2563eb}}
.sidebar.collapsed-sidebar .sidebar-collapse-btn{{display:none}}
.sidebar.collapsed-sidebar .sidebar-expand-btn{{display:flex;align-items:center;justify-content:center}}
.resize-handle{{position:absolute;top:0;right:-4px;width:8px;height:100%;cursor:col-resize;background:transparent;z-index:10}}
.resize-handle:hover{{background:#2563eb40}}
.resizing iframe{{pointer-events:none}}

.main{{flex:1;display:flex;flex-direction:column;height:100vh;overflow:hidden}}
.toolbar{{padding:6px 16px;background:#fff;border-bottom:1px solid #e5e7eb;font-size:12px;color:#888;flex-shrink:0;display:flex;align-items:center;gap:8px}}
.toolbar .path{{color:#333;font-weight:600}}
.viewer{{flex:1;border:none;width:100%}}
.empty{{flex:1;display:flex;align-items:center;justify-content:center;color:#ccc;font-size:13px}}

.plan-section{{background:#fff;border-radius:8px;overflow:hidden;box-shadow:0 1px 3px rgba(0,0,0,0.08);margin-bottom:16px}}
.plan-section-header{{padding:10px 14px;font-weight:600;font-size:14px;background:#f8f9fa;border-bottom:1px solid #e5e7eb;display:flex;align-items:center;gap:8px}}
</style>
</head>
<body>

<div class="sidebar" id="sidebar">
  <div class="resize-handle" id="resizeHandle"></div>
  <div class="sidebar-header"><h1>Mantle Test Reports</h1></div>
  <input class="search" id="search" placeholder="Search..." type="text">
  <div class="sidebar-content">
    {sidebar_html}
  </div>
  <button class="sidebar-collapse-btn" onclick="toggleSidebar()">◀ Collapse</button>
  <button class="sidebar-expand-btn" onclick="toggleSidebar()">▶</button>
</div>

<div class="main">
  <div class="toolbar">
    <span>Report:</span>
    <span class="path" id="path">Select a report</span>
  </div>
  <div class="empty" id="empty">Select a report from the sidebar — {total_modules} modules, {total_reports} reports, {total_plans} plans</div>
  <div id="planView" style="display:none;padding:24px;overflow-y:auto;flex:1"></div>
  <iframe class="viewer" id="viewer" style="display:none"></iframe>
</div>

<script>
const allFiles={all_files_json};

let lastWidth=320;
function toggleSidebar(){{
  const sb=document.getElementById('sidebar');
  if(sb.classList.contains('collapsed-sidebar')){{
    sb.classList.remove('collapsed-sidebar');
    sb.style.width=lastWidth+'px';
  }}else{{
    lastWidth=sb.offsetWidth;
    sb.classList.add('collapsed-sidebar');
    sb.style.width='';
  }}
}}

// Drag to resize sidebar
(function(){{
  const handle=document.getElementById('resizeHandle');
  const sb=document.getElementById('sidebar');
  let startX,startW;
  handle.addEventListener('mousedown',function(e){{
    startX=e.clientX;
    startW=sb.offsetWidth;
    document.body.classList.add('resizing');
    document.addEventListener('mousemove',onDrag);
    document.addEventListener('mouseup',stopDrag);
    e.preventDefault();
  }});
  function onDrag(e){{
    const w=Math.max(220,Math.min(600,startW+(e.clientX-startX)));
    sb.style.width=w+'px';
  }}
  function stopDrag(){{
    document.body.classList.remove('resizing');
    document.removeEventListener('mousemove',onDrag);
    document.removeEventListener('mouseup',stopDrag);
  }}
}})();

function toggle(el){{
  el.classList.toggle('collapsed');
  const list=el.nextElementSibling;
  if(list)list.style.display=el.classList.contains('collapsed')?'none':'block';
}}
function toggleEnv(el){{
  el.classList.toggle('collapsed');
  const list=el.nextElementSibling;
  if(list)list.style.display=el.classList.contains('collapsed')?'none':'block';
}}

function expandTo(anchor){{
  // Walk up: expand env-group AND module group
  const envGroup=anchor.closest('.env-group');
  if(envGroup){{
    const envHeader=envGroup.querySelector('.env-header');
    const envList=envGroup.querySelector('.env-items');
    if(envHeader&&envHeader.classList.contains('collapsed')){{
      envHeader.classList.remove('collapsed');
      if(envList)envList.style.display='block';
    }}
  }}
  const group=anchor.closest('.group');
  if(group){{
    const header=group.querySelector('.group-header');
    const inner=header&&header.nextElementSibling;
    if(header&&header.classList.contains('collapsed')){{
      header.classList.remove('collapsed');
      if(inner)inner.style.display='block';
    }}
  }}
}}

function showReport(href){{
  document.querySelectorAll('.items a, .env-items a').forEach(a=>a.classList.remove('active'));
  document.querySelectorAll('.items a, .env-items a').forEach(a=>{{
    const onclick=a.getAttribute('onclick')||'';
    if(onclick.includes("'"+href+"'")){{
      a.classList.add('active');
      expandTo(a);
      a.scrollIntoView({{block:'nearest'}});
    }}
  }});
  document.getElementById('viewer').src=href;
  document.getElementById('viewer').style.display='block';
  document.getElementById('planView').style.display='none';
  document.getElementById('empty').style.display='none';
  document.getElementById('path').textContent=href;
  location.hash=href;
}}

// Load from hash — run immediately
(function(){{
  const hash=decodeURIComponent(location.hash.slice(1));
  if(hash&&!hash.startsWith('plan:')){{
    showReport(hash);
  }}else if(hash.startsWith('plan:')){{
    showPlan(hash.slice(5));
  }}
}})();

function showPlan(name){{
  // Group matching reports by env: {{env: [{{module, path, filename}}]}}
  const byEnv={{}};
  Object.keys(allFiles).sort().forEach(mod=>{{
    const envMap=allFiles[mod];
    Object.keys(envMap).forEach(env=>{{
      envMap[env].forEach(f=>{{
        const fname=f.split('/').pop();
        if(fname.startsWith(name+'-')){{
          const key=env||'(root)';
          if(!byEnv[key])byEnv[key]=[];
          byEnv[key].push({{module:mod,path:f,filename:fname}});
        }}
      }});
    }});
  }});

  const envOrder=Object.keys(byEnv).sort((a,b)=>{{
    if(a==='(root)')return 1;
    if(b==='(root)')return -1;
    return a.localeCompare(b);
  }});

  let html='<h2 style="font-size:18px;margin-bottom:16px">Plan: '+name+'</h2>';
  if(envOrder.length===0){{
    html+='<div style="padding:16px;color:#999;text-align:center;background:#fff;border-radius:8px">No reports found for this plan</div>';
  }}else{{
    envOrder.forEach(env=>{{
      const rows=byEnv[env].map(r=>
        '<tr><td style="padding:10px 14px;border-bottom:1px solid #f0f0f0;font-weight:600;font-size:14px;width:140px">'+r.module+'</td>'+
        '<td style="padding:10px 14px;border-bottom:1px solid #f0f0f0">'+
        '<a href="javascript:void(0)" onclick="showReport(\\''+r.path+'\\')" style="color:#2563eb;font-size:14px;text-decoration:none">'+r.filename+'</a></td></tr>'
      ).join('');
      html+='<div class="plan-section">'+
        '<div class="plan-section-header">'+
          '<span>Environment:</span>'+
          '<span style="padding:2px 8px;border-radius:8px;font-size:12px;background:#e0f2fe;color:#0369a1">'+env+'</span>'+
          '<span style="color:#999;font-size:12px;margin-left:auto">'+byEnv[env].length+' reports</span>'+
        '</div>'+
        '<table style="border-collapse:collapse;width:100%"><thead><tr style="background:#fafafa">'+
          '<th style="text-align:left;padding:8px 14px;border-bottom:1px solid #e5e7eb;font-size:12px;color:#666;font-weight:600">Module</th>'+
          '<th style="text-align:left;padding:8px 14px;border-bottom:1px solid #e5e7eb;font-size:12px;color:#666;font-weight:600">Report</th>'+
        '</tr></thead><tbody>'+rows+'</tbody></table>'+
        '</div>';
    }});
  }}

  const pv=document.getElementById('planView');
  pv.innerHTML=html;
  pv.style.display='block';
  document.getElementById('viewer').style.display='none';
  document.getElementById('empty').style.display='none';
  document.getElementById('path').textContent='Plan: '+name;
  location.hash='plan:'+name;
}}

document.getElementById('search').addEventListener('input',function(){{
  const q=this.value.toLowerCase();
  // Search across module links, env links, and report links.
  document.querySelectorAll('.group').forEach(g=>{{
    let visGroup=false;
    // Plans list: simple
    g.querySelectorAll('.items > li a').forEach(a=>{{
      const match=(a.dataset.search||'').toLowerCase().includes(q);
      a.parentElement.classList.toggle('hidden',!match);
      if(match)visGroup=true;
    }});
    // Env sub-groups
    g.querySelectorAll('.env-group').forEach(eg=>{{
      let visEnv=false;
      eg.querySelectorAll('.env-items a').forEach(a=>{{
        const match=(a.dataset.search||'').toLowerCase().includes(q);
        a.parentElement.classList.toggle('hidden',!match);
        if(match)visEnv=true;
      }});
      eg.classList.toggle('hidden',!visEnv);
      if(visEnv){{
        visGroup=true;
        if(q){{
          const eh=eg.querySelector('.env-header');
          if(eh){{eh.classList.remove('collapsed');const el=eh.nextElementSibling;if(el)el.style.display='block';}}
        }}
      }}
    }});
    g.classList.toggle('hidden',!visGroup);
    if(visGroup&&q){{
      const gh=g.querySelector('.group-header');
      if(gh){{gh.classList.remove('collapsed');const inner=gh.nextElementSibling;if(inner)inner.style.display='block';}}
    }}
  }});
}});
</script>
</body>
</html>"""

    Path("index.html").write_text(html)
    print(f"Generated index.html ({total_modules} modules, {total_reports} reports, {total_plans} plans)")


if __name__ == "__main__":
    main()
