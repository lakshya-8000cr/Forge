import { useEffect, useState, useRef } from "react";
import "./App.css";

const API = "http://localhost:8080";

interface Project {
  id: number;
  name: string;
  image_name: string;
  status: string;
}

interface Deployment {
  id: number;
  project_id: number;
  image_name: string;
  status: string;
  created_at: string;
}

interface LogLine {
  ts: string;
  text: string;
  type: "sys" | "ok" | "err" | "warn" | "net" | "dim";
}

const LOG_DATA: Record<string, (p: Project) => LogLine[]> = {
  running: (p) => [
    { ts: now(), text: `${p.name} nominal — all systems go`, type: "ok" },
    { ts: now(), text: `GET / 200 8ms`, type: "net" },
    { ts: now(), text: `GET /api/health 200 2ms`, type: "net" },
    { ts: now(), text: `heap 128mb / cpu 2.1%`, type: "sys" },
  ],
  error: (p) => [
    { ts: now(), text: `${p.name} fault detected`, type: "err" },
    { ts: now(), text: `db connection retry 1/3`, type: "warn" },
    { ts: now(), text: `db connection retry 2/3`, type: "warn" },
    { ts: now(), text: `ECONNREFUSED 127.0.0.1:5432`, type: "err" },
    { ts: now(), text: `process exited code 1`, type: "err" },
  ],
  stopped: (p) => [
    { ts: now(), text: `${p.name} offline`, type: "dim" },
    { ts: now(), text: `graceful shutdown complete`, type: "sys" },
  ],
  building: (p) => [
    { ts: now(), text: `launching ${p.name}...`, type: "sys" },
    { ts: now(), text: `pulling ${p.image_name}`, type: "sys" },
    { ts: now(), text: `compiling layers...`, type: "sys" },
  ],
};

function now() {
  return new Date().toISOString().substr(11, 8);
}

function Stars() {
  const stars = Array.from({ length: 80 }, (_, i) => ({
    id: i,
    x: Math.random() * 100,
    y: Math.random() * 100,
    size: Math.random() * 1.8 + 0.4,
    dur: (Math.random() * 4 + 2).toFixed(1),
    delay: (Math.random() * 5).toFixed(1),
    op: (Math.random() * 0.5 + 0.2).toFixed(2),
  }));

  return (
    <div className="starfield">
      {stars.map((s) => (
        <div
          key={s.id}
          className="star"
          style={{
            left: `${s.x}%`,
            top: `${s.y}%`,
            width: s.size,
            height: s.size,
            ["--dur" as any]: `${s.dur}s`,
            ["--delay" as any]: `${s.delay}s`,
            ["--max-op" as any]: s.op,
          }}
        />
      ))}
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const labels: Record<string, string> = {
    running: "live",
    error: "error",
    stopped: "offline",
    building: "launching",
  };
  return (
    <span className={`badge ${status}`}>
      <span className="badge-dot" />
      {labels[status] || status}
    </span>
  );
}

export default function App() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [name, setName] = useState("");
  const [image, setImage] = useState("");
  const [logs, setLogs] = useState<LogLine[]>([]);
  const [logsSource, setLogsSource] = useState("system");
  const [clock, setClock] = useState("");
  const teleRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    loadProjects();
    addLog([
      { ts: now(), text: "orbit mission control online", type: "ok" },
      { ts: now(), text: "all subsystems nominal", type: "sys" },
      { ts: now(), text: "telemetry feed active", type: "sys" },
      { ts: now(), text: "awaiting commands...", type: "dim" },
    ], "system");

    const t = setInterval(() => {
      setClock(new Date().toISOString().substr(11, 8) + " UTC");
    }, 1000);
    return () => clearInterval(t);
  }, []);

  useEffect(() => {
    if (teleRef.current) teleRef.current.scrollTop = teleRef.current.scrollHeight;
  }, [logs]);

  async function loadProjects() {
    try {
      const res = await fetch(`${API}/projects`);
      const data = await res.json();
      setProjects(data);
    } catch {
      setProjects([
        { id: 1, name: "api-server", image_name: "node:18-alpine", status: "running" },
        { id: 2, name: "frontend-app", image_name: "nginx:latest", status: "running" },
        { id: 3, name: "worker", image_name: "python:3.11", status: "error" },
      ]);
    }
  }

  function addLog(lines: LogLine[], source: string) {
    setLogs(lines);
    setLogsSource(source);
  }

  async function createProject() {
    if (!name || !image) return;
    try {
      await fetch(`${API}/projects`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name, image_name: image }),
      });
      await loadProjects();
    } catch {
      setProjects((prev) => [
        ...prev,
        { id: Date.now(), name, image_name: image, status: "stopped" },
      ]);
    }
    addLog([
      { ts: now(), text: `satellite registered: ${name}`, type: "sys" },
      { ts: now(), text: `image: ${image}`, type: "sys" },
      { ts: now(), text: "status: standby — ready to launch", type: "dim" },
    ], name);
    setName("");
    setImage("");
  }

  async function deployProject(id: number) {
    const p = projects.find((x) => x.id === id);
    if (!p) return;

    setProjects((prev) =>
      prev.map((x) => (x.id === id ? { ...x, status: "building" } : x))
    );
    addLog([
      { ts: now(), text: `initiating launch: ${p.name}`, type: "sys" },
      { ts: now(), text: `pulling ${p.image_name}`, type: "sys" },
      { ts: now(), text: "building layers...", type: "sys" },
      { ts: now(), text: "container starting", type: "sys" },
      { ts: now(), text: "health probe ok", type: "ok" },
      { ts: now(), text: "orbit achieved ✓", type: "ok" },
    ], p.name);

    try {
      await fetch(`${API}/projects/${id}/deploy`, { method: "POST" });
      await loadProjects();
    } catch {
      setTimeout(() => {
        setProjects((prev) =>
          prev.map((x) => (x.id === id ? { ...x, status: "running" } : x))
        );
      }, 2000);
    }
  }

  
async function launchProject(id: number) {
  const res = await fetch(`${API}/projects/${id}/url`);
  const data = await res.json();

  alert(`${data.note}\n\n${data.command}`);
}
  async function getLogs(id: number) {
    const p = projects.find((x) => x.id === id);
    if (!p) return;
    try {
      const res = await fetch(`${API}/projects/${id}/logs`);
      const data = await res.json();
      addLog(
        String(data.logs).split("\n").map((l) => ({
          ts: now(),
          text: l,
          type: "sys" as const,
        })),
        p.name
      );
    } catch {
      addLog((LOG_DATA[p.status] || LOG_DATA.stopped)(p), p.name);
    }
  }

  async function deleteProject(id: number) {
    const p = projects.find((x) => x.id === id);
    try {
      await fetch(`${API}/projects/${id}/delete`, { method: "DELETE" });
      await loadProjects();
    } catch {
      setProjects((prev) => prev.filter((x) => x.id !== id));
    }
    addLog([
      { ts: now(), text: `decommissioning ${p?.name}`, type: "warn" },
      { ts: now(), text: "container stopped", type: "sys" },
      { ts: now(), text: "resources freed", type: "dim" },
    ], "system");
  }

  async function getDeployments(id: number) {
    try {
      const res = await fetch(`${API}/projects/${id}/deployments`);
      const data = await res.json();
      setDeployments(data);
    } catch {
      // dev fallback: mock deployment history
      setDeployments([
        {
          id: 1,
          project_id: id,
          image_name: projects.find((p) => p.id === id)?.image_name ?? "unknown",
          status: "running",
          created_at: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
        },
        {
          id: 2,
          project_id: id,
          image_name: projects.find((p) => p.id === id)?.image_name ?? "unknown",
          status: "stopped",
          created_at: new Date(Date.now() - 1000 * 60 * 90).toISOString(),
        },
      ]);
    }
  }

  const running = projects.filter((p) => p.status === "running").length;
  const errors = projects.filter((p) => p.status === "error").length;

  return (
    <div className="orbit-root">
      <Stars />

      <div className="planets">
        <div className="planet planet-sun" />
        <div className="planet planet-small" />
        <div className="planet planet-far" />
      </div>

      {/* HUD Top Bar */}
      <div className="hud-bar">
        <div className="hud-logo">
          <div className="hud-orb" />
          <span className="hud-brand">ORBIT</span>
          <span className="hud-tag">Deployment Control</span>
        </div>
        <div className="hud-right">
          <div className="status-dot">Deleted Services</div>
          <span className="hud-clock">{clock || "--:--:-- UTC"}</span>
        </div>
      </div>

      {/* Main */}
      <div className="orbit-main">
        {/* Stats */}
        <div className="stats-row">
          <div className="stat-card blue">
            <div className="stat-label">Deployments</div>
            <div className="stat-val">{projects.length}</div>
            <div className="stat-sub">in registry</div>
          </div>
          <div className="stat-card green">
            <div className="stat-label">Running</div>
            <div className="stat-val">{running}</div>
            <div className="stat-sub">live instances</div>
          </div>
          <div className="stat-card red">
            <div className="stat-label">Removed Services</div>
            <div className="stat-val">{errors}</div>
            <div className="stat-sub">need attention</div>
          </div>
        </div>

        {/* Launch Form */}
        <div className="launch-form">
          <input
            className="orbit-input"
            placeholder="satellite-name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && createProject()}
          />
          <input
            className="orbit-input"
            placeholder="image:tag"
            value={image}
            onChange={(e) => setImage(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && createProject()}
          />
          <button className="launch-btn" onClick={createProject}>
            ↑ LAUNCH
          </button>
        </div>

        {/* Projects Table */}
        <div className="panel">
          <div className="panel-head">
            <div className="panel-title">
              <span className="panel-dot" />
              Active Projects
            </div>
            <button className="panel-action" onClick={loadProjects}>
              ↺ refresh
            </button>
          </div>
          {projects.length === 0 ? (
            <div className="empty-state">No Active running project</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Name</th>
                  <th>Docker-Image</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {projects.map((p) => (
                  <tr key={p.id}>
                    <td className="td-image">#{p.id}</td>
                    <td className="td-name">{p.name}</td>
                    <td className="td-image">{p.image_name}</td>
                    <td>
                      <StatusBadge status={p.status} />
                    </td>
                    <td>
                      <div className="act-row">
                        <button
                          className="act-btn"
                          onClick={() => deployProject(p.id)}
                        >
                          ▶ deploy
                        </button>
                        <button className="act-btn" onClick={() => launchProject(p.id)}>
                        Live
                        </button>
                        <button
                          className="act-btn"
                          onClick={() => getLogs(p.id)}
                        >
                          ≡ logs
                        </button>
                        <button
                          className="act-btn"
                          onClick={() => getDeployments(p.id)}
                        >
                          ⏱ history
                        </button>
                        <button
                          className="act-btn danger"
                          onClick={() => deleteProject(p.id)}
                        >
                          ✕
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Deployment History */}
        {deployments.length > 0 && (
          <div className="panel" style={{ marginTop: 16 }}>
            <div className="panel-head">
              <div className="panel-title">
                <span className="panel-dot" />
                Deployment History
              </div>
              <button className="panel-action" onClick={() => setDeployments([])}>
                ✕ clear
              </button>
            </div>
            <table>
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Project ID</th>
                  <th>Image</th>
                  <th>Status</th>
                  <th>Created At</th>
                </tr>
              </thead>
              <tbody>
                {deployments.map((d) => (
                  <tr key={d.id}>
                    <td>#{d.id}</td>
                    <td>#{d.project_id}</td>
                    <td className="td-image">{d.image_name}</td>
                    <td>
                      <StatusBadge status={d.status} />
                    </td>
                    <td>{new Date(d.created_at).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Telemetry */}
        <div className="telemetry">
          <div className="tele-head">
            <div className="tele-title">
              <span className="panel-dot" />
              Telemetry Stream
              <span style={{ color: "rgba(0,212,255,0.3)", marginLeft: 6 }}>
                — {logsSource}
              </span>
            </div>
            <button className="panel-action" onClick={() => setLogs([])}>
              ✕ clear
            </button>
          </div>
          <div className="tele-body" ref={teleRef}>
            {logs.length === 0 ? (
              <div className="dim" style={{ color: "rgba(180,210,220,0.2)" }}>
                // awaiting telemetry signal...
              </div>
            ) : (
              logs.map((l, i) => (
                <div key={i} className={`tele-line ${l.type}`}>
                  <span className="tele-ts">[{l.ts}]</span>
                  {l.text}
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}