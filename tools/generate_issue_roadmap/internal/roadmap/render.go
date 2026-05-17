package roadmap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
)

func writeHTML(path string, data htmlData) error {
	funcMap := template.FuncMap{
		"colorClass": func(n int) string {
			return fmt.Sprintf("c%d", (n-1)%8+1)
		},
		"jsStr": func(s string) template.JS {
			b, _ := json.Marshal(s)
			return template.JS(b)
		},
	}

	tmpl, err := template.New("roadmap").Funcs(funcMap).Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("HTMLテンプレート解析失敗: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("HTML生成失敗: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("HTML書き込み失敗: %w", err)
	}
	return nil

}

const htmlTemplate = `<!doctype html>
<html lang="ja">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{{.ProjectTitle}}進捗</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; }
    body {
      margin: 0;
      padding: 20px 24px 40px;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Hiragino Sans", "Yu Gothic", sans-serif;
      font-size: 13px;
      color: #1f2937;
      background: #f8fafc;
    }
    h1 { margin: 0 0 4px; font-size: 22px; }
    .meta { color: #374151; margin-bottom: 3px; }
    .chart-wrap {
      margin-top: 16px;
      background: #fff;
      border: 1px solid #e5e7eb;
      border-radius: 12px;
      padding: 16px 16px 12px;
      overflow-x: auto;
    }
    /* ガントチャート */
	.gantt { display: table; width: 100%; border-collapse: collapse; min-width: 700px; position: relative; table-layout: fixed; }
    .gantt-row { display: table-row; }
    .gantt-row:hover .gantt-bar-cell { background: #f0fdf4; }
    .gantt-label {
      display: table-cell;
      width: 210px;
      max-width: 210px;
      padding: 3px 10px 3px 0;
      vertical-align: middle;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 11px;
      color: #374151;
      border-bottom: 1px solid #f3f4f6;
    }
    .gantt-label.is-phase {
      font-weight: 700;
      color: #0f172a;
      font-size: 13px;
      padding-top: 14px;
      padding-bottom: 2px;
    }
    .gantt-label.is-parent {
      font-weight: 600;
      color: #374151;
      font-size: 12px;
      padding-left: 14px;
      padding-top: 6px;
    }
    .gantt-label.is-ticket { padding-left: 32px; }
    .gantt-row.is-phase-row .gantt-label,
    .gantt-row.is-phase-row .gantt-bar-cell {
      border-top: 2px solid #64748b;
    }
    .gantt-bar-cell {
      display: table-cell;
      vertical-align: middle;
      padding: 3px 0;
      border-bottom: 1px solid #f3f4f6;
      position: relative;
    }
    .bar-bg {
      position: relative;
      height: 22px;
      background: #f1f5f9;
      border-radius: 4px;
    }
		.progress-line-global {
			position: absolute;
			top: 0;
			width: 3px;
			pointer-events: none;
			background: rgba(153, 27, 27, 0.98);
			border-radius: 0;
			box-shadow: 0 0 0 1px rgba(153, 27, 27, 0.18);
			z-index: 4;
		}
    .bar-parent {
      position: absolute;
      top: 3px; bottom: 3px;
      border-radius: 3px;
      background: rgba(100,116,139,0.22);
      pointer-events: none;
    }
    .bar-ticket {
      position: absolute;
      top: 0; bottom: 0;
      border-radius: 4px;
      cursor: pointer;
      transition: filter .15s;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 10px;
      font-weight: 700;
      color: rgba(255,255,255,0.92);
      text-shadow: 0 1px 2px rgba(0,0,0,.3);
      overflow: hidden;
      text-decoration: none;
    }
    .bar-ticket:hover { filter: brightness(1.12); }
    /* 実行順カラーパレット (vegalite版と近似) */
    .c1  { background: #4e79a7; }
    .c2  { background: #f28e2b; }
    .c3  { background: #e15759; }
    .c4  { background: #76b7b2; }
    .c5  { background: #59a14f; }
    .c6  { background: #edc948; }
    .c7  { background: #b07aa1; }
    .c8  { background: #ff9da7; }
    /* X軸ラベル */
    .axis-row { display: table-row; }
    .axis-label-cell { display: table-cell; }
    .axis-bar-cell  { display: table-cell; }
    .x-axis {
      position: relative;
      height: 24px;
      margin-top: 2px;
    }
    .x-tick {
      position: absolute;
      top: 0;
      transform: translateX(-50%);
      font-size: 10px;
      color: #6b7280;
      white-space: nowrap;
    }
    .x-tick::before {
      content: '';
      position: absolute;
      top: -4px;
      left: 50%;
      transform: translateX(-50%);
      width: 1px;
      height: 4px;
      background: #d1d5db;
    }
    /* ツールチップ */
    #tt {
      position: fixed;
      display: none;
      z-index: 9999;
      background: #1e293b;
      color: #f1f5f9;
      border-radius: 8px;
      padding: 10px 14px;
      font-size: 12px;
      line-height: 1.6;
      max-width: 320px;
      pointer-events: none;
      box-shadow: 0 8px 24px rgba(0,0,0,.3);
    }
    #tt .tt-title { font-weight: 700; margin-bottom: 4px; color: #e2e8f0; }
    #tt .tt-row   { display: flex; gap: 8px; }
    #tt .tt-key   { color: #94a3b8; min-width: 60px; }
    #tt .tt-val   { color: #f8fafc; }
  </style>
</head>
<body>
  <h1>{{.ProjectTitle}}進捗</h1>
  <div class="meta">生成日時: {{.GeneratedAt}}</div>
  <div class="meta">ソース: GitHub Project #{{.ProjectNumber}} ({{.Owner}}/{{.Repo}}) / アイテム数: {{.TotalItems}}</div>
  <div class="meta">難易度マッピング: {{.DifficultyLegend}}</div>

  <div class="chart-wrap">
    <div class="gantt" id="gantt"></div>
    <div style="display:table;width:100%;min-width:700px;">
      <div style="display:table-row;">
        <div style="display:table-cell;width:220px;min-width:160px;"></div>
        <div style="display:table-cell;">
          <div class="x-axis" id="xaxis"></div>
        </div>
      </div>
    </div>
  </div>

  <div id="tt"></div>

  <script>
  (function(){
    const MAX = {{.MaxDifficulty}};
    const lanes = [
      {{range .Lanes}}
      {
        phase: {{jsStr .Phase}},
        label: {{jsStr .Label}},
				start: {{.Start}},
        total: {{.TotalDifficulty}},
				isPhase: {{.IsPhase}},
				completed: {{.CompletedEnd}},
        tickets: [
          {{range .Tickets}}
          {
            id:      {{jsStr .ID}},
            code:    {{jsStr .TicketCode}},
            title:   {{jsStr .Title}},
            url:     {{jsStr .URL}},
            est:     {{jsStr .Estimate}},
            dep:     {{jsStr .DependsOn}},
						status:  {{jsStr .Status}},
						done:    {{.Completed}},
            diff:    {{.Difficulty}},
            start:   {{.Start}},
            end:     {{.End}},
            order:   {{.ExecutionOrder}},
          },
          {{end}}
        ],
      },
      {{end}}
    ];

    const gantt = document.getElementById('gantt');
    const tt = document.getElementById('tt');

    function pct(v){ return (v / MAX * 100).toFixed(4) + '%'; }

		var curPhase = null;
    lanes.forEach(function(lane){
			if(lane.phase !== curPhase) curPhase = lane.phase;
      // 親行
      var pr = document.createElement('div');
			pr.className = lane.isPhase ? 'gantt-row is-phase-row' : 'gantt-row';

      var pl = document.createElement('div');
			pl.className = lane.isPhase ? 'gantt-label is-phase' : 'gantt-label is-parent';
      pl.textContent = lane.label;
      pl.title = lane.label;
      pr.appendChild(pl);

      var pbc = document.createElement('div');
      pbc.className = 'gantt-bar-cell';
      var pbg = document.createElement('div');
      pbg.className = 'bar-bg';
      var pbr = document.createElement('div');
      pbr.className = 'bar-parent';
			pbr.style.left = pct(lane.start);
      pbr.style.width = pct(lane.total);
      pbg.appendChild(pbr);
      pbc.appendChild(pbg);
      pr.appendChild(pbc);
      gantt.appendChild(pr);

      // チケット行
      lane.tickets.forEach(function(tk){
        var row = document.createElement('div');
        row.className = 'gantt-row';

        var lbl = document.createElement('div');
        lbl.className = 'gantt-label is-ticket';
        lbl.textContent = '  ' + tk.title;
        lbl.addEventListener('mouseenter', function(e){
          tt.innerHTML =
            '<div class="tt-title">' + esc(tk.title) + '</div>' +
            '<div class="tt-row"><span class="tt-key">チケット</span><span class="tt-val">' + esc(tk.id) + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">状態</span><span class="tt-val">' + esc(tk.status || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">Estimate</span><span class="tt-val">' + esc(tk.est || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">難易度</span><span class="tt-val">' + tk.diff + '</span></div>';
          tt.style.display = 'block';
          moveTT(e);
        });
        lbl.addEventListener('mousemove', moveTT);
        lbl.addEventListener('mouseleave', function(){ tt.style.display = 'none'; });
        row.appendChild(lbl);

        var bc = document.createElement('div');
        bc.className = 'gantt-bar-cell';
        var bg = document.createElement('div');
        bg.className = 'bar-bg';

        var bar = document.createElement('a');
        bar.className = 'bar-ticket c' + ((tk.order - 1) % 8 + 1);
        bar.href = tk.url;
        bar.target = '_blank';
        bar.rel = 'noopener';
        bar.style.left  = pct(tk.start);
        bar.style.width = pct(tk.end - tk.start);
        bar.textContent = tk.code;

        // ツールチップ
        bar.addEventListener('mouseenter', function(e){
          tt.innerHTML =
            '<div class="tt-title">' + esc(tk.title) + '</div>' +
            '<div class="tt-row"><span class="tt-key">チケット</span><span class="tt-val">' + esc(tk.id) + '</span></div>' +
						'<div class="tt-row"><span class="tt-key">状態</span><span class="tt-val">' + esc(tk.status || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">Estimate</span><span class="tt-val">' + esc(tk.est || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">難易度</span><span class="tt-val">' + tk.diff + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">範囲</span><span class="tt-val">' + tk.start + ' - ' + tk.end + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">依存</span><span class="tt-val">' + esc(tk.dep || '-') + '</span></div>' +
            '<div class="tt-row"><span class="tt-key">実行順</span><span class="tt-val">' + tk.order + '</span></div>';
          tt.style.display = 'block';
          moveTT(e);
        });
        bar.addEventListener('mousemove', moveTT);
        bar.addEventListener('mouseleave', function(){ tt.style.display = 'none'; });

        bg.appendChild(bar);
        bc.appendChild(bg);
        row.appendChild(bc);
        gantt.appendChild(row);
      });
    });

		// 各行ではなく、ガント全体を跨ぐ1本の進捗線を描画
		// lane.completed == lane.start は「完了なし」を意味するため除外する
		var globalCompleted = 0;
		lanes.forEach(function(lane){
			if(!lane.isPhase && lane.completed > lane.start && lane.completed > globalCompleted) globalCompleted = lane.completed;
		});
		renderGlobalProgressLine(globalCompleted);

    // X軸目盛り
    var xaxis = document.getElementById('xaxis');
    var step = MAX <= 20 ? 5 : MAX <= 50 ? 10 : 25;
    for(var v = 0; v <= MAX; v += step){
      var tick = document.createElement('div');
      tick.className = 'x-tick';
      tick.style.left = pct(v);
      tick.textContent = v;
      xaxis.appendChild(tick);
    }

    function moveTT(e){
      var x = e.clientX + 14, y = e.clientY + 14;
      if(x + 330 > window.innerWidth) x = e.clientX - 330;
      if(y + 160 > window.innerHeight) y = e.clientY - 160;
      tt.style.left = x + 'px';
      tt.style.top  = y + 'px';
    }
		function renderGlobalProgressLine(completed){
			var bars = gantt.querySelectorAll('.bar-bg');
			if(!bars.length) return;

			var firstBar = bars[0];
			var lastBar = bars[bars.length - 1];
			var line = document.createElement('div');
			line.className = 'progress-line-global';
			gantt.appendChild(line);

			function positionLine(){
				var pos = completed > 0 ? completed : 0;
				var ganttRect = gantt.getBoundingClientRect();
				var firstRect = firstBar.getBoundingClientRect();
				var lastRect = lastBar.getBoundingClientRect();

				var x = (firstRect.left - ganttRect.left) + (firstRect.width * pos / MAX);
				var top = firstRect.top - ganttRect.top - 1;
				var height = (lastRect.bottom - firstRect.top) + 2;

				line.style.left = (x - 1.5) + 'px';
				line.style.top = top + 'px';
				line.style.height = height + 'px';
			}

			positionLine();
			window.addEventListener('resize', positionLine);
		}
    function esc(s){
      return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
    }
  })();
  </script>
</body>
</html>
`
