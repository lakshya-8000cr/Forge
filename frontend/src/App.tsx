import { useEffect, useState, useRef } from "react";
import "./App.css";

const API = "/api";

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

type View = "overview" | "detail";

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

function StatusBadge({ status }: { status: string }) {
  const labels: Record<string, string> = {
    running: "Running",
    error: "Failed",
    stopped: "Stopped",
    building: "Deploying",
  };
  return (
    <span className={`badge ${status}`}>
      <span className="badge-dot" />
      {labels[status] || status}
    </span>
  );
}

export default function App() {
  const [view, setView] = useState<View>("overview");
  const [projects, setProjects] = useState<Project[]>([]);
  const [activeProjectId, setActiveProjectId] = useState<number | null>(null);
  const [deployments, setDeployments] = useState<Deployment[]>([]);
  const [logs, setLogs] = useState<LogLine[]>([]);
  const [logsSource, setLogsSource] = useState("system");
  const [clock, setClock] = useState("");
  const teleRef = useRef<HTMLDivElement>(null);
  const [message, setMessage] = useState("");

  // Create-service modal state
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [modalName, setModalName] = useState("");
  const [modalImage, setModalImage] = useState("");
  const [modalError, setModalError] = useState("");
  const [creating, setCreating] = useState(false);

  useEffect(() => {
    loadProjects();
    addLog([
      { ts: now(), text: "orbit dashboard ready", type: "ok" },
      { ts: now(), text: "connected to local API", type: "sys" },
      { ts: now(), text: "watching for events...", type: "dim" },
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

  function openCreateModal() {
    setModalName("");
    setModalImage("");
    setModalError("");
    setShowCreateModal(true);
  }

  function closeCreateModal() {
    if (creating) return;
    setShowCreateModal(false);
  }

  async function submitCreateProject() {
    const trimmedName = modalName.trim();
    const trimmedImage = modalImage.trim();

    if (!trimmedName || !trimmedImage) {
      setModalError("Service name and Docker image are both required.");
      return;
    }

    setCreating(true);
    setModalError("");

    try {
      await fetch(`${API}/projects`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: trimmedName, image_name: trimmedImage }),
      });
      await loadProjects();
    } catch {
      setProjects((prev) => [
        ...prev,
        { id: Date.now(), name: trimmedName, image_name: trimmedImage, status: "stopped" },
      ]);
    }

    addLog([
      { ts: now(), text: `service created: ${trimmedName}`, type: "sys" },
      { ts: now(), text: `image: ${trimmedImage}`, type: "sys" },
      { ts: now(), text: "status: stopped — ready to deploy", type: "dim" },
    ], trimmedName);

    setCreating(false);
    setShowCreateModal(false);
  }

  async function deployProject(id: number) {
    setMessage("Deploying...");

    try {
      const res = await fetch(`${API}/projects/${id}/deploy`, {
        method: "POST",
      });

      const data = await res.json();

      if (!res.ok) {
        setMessage(`Failed: ${data.error} — ${data.details}`);
        await loadProjects();
        return;
      }

      setMessage(`${data.message}`);
      await loadProjects();
    } catch {
      setMessage("Failed: could not reach the API");
    }
  }

async function launchProject(id: number) {
  const res = await fetch(`${API}/projects/${id}/url`);
  const data = await res.json();

  if (data.url) {
    window.open(data.url, "_blank");
  } else {
    alert("App URL not found");
  }
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
      { ts: now(), text: `deleting ${p?.name}`, type: "warn" },
      { ts: now(), text: "container stopped", type: "sys" },
      { ts: now(), text: "resources freed", type: "dim" },
    ], "system");

    // If we deleted the project currently open in detail view, bounce back to overview
    if (activeProjectId === id) {
      setActiveProjectId(null);
      setView("overview");
    }
  }

  async function getDeployments(id: number) {
    try {
      const res = await fetch(`${API}/projects/${id}/deployments`);
      const data = await res.json();
      setDeployments(data);
    } catch {
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

  function openProjectDetail(id: number) {
    setActiveProjectId(id);
    setView("detail");
    setMessage("");
    getLogs(id);
  }

  function backToOverview() {
    setView("overview");
    setActiveProjectId(null);
  }

  const running = projects.filter((p) => p.status === "running").length;
  const errors = projects.filter((p) => p.status === "error").length;
  const activeProject = projects.find((p) => p.id === activeProjectId) || null;

  return (
    <div className="orbit-root">
      {/* Top bar with breadcrumb */}
      <div className="topbar">
        <div className="topbar-left">
          <div className="brand-mark">
            <span className="brand-icon">◆</span>
            <span className="brand-name">ORBIT</span>
          </div>
          <span className="crumb-sep">/</span>
          {view === "overview" ? (
            <span className="crumb-item active">Projects</span>
          ) : (
            <>
              <span className="crumb-item crumb-link" onClick={backToOverview}>
                Projects
              </span>
              <span className="crumb-sep">/</span>
              <span className="crumb-item active">{activeProject?.name ?? "Service"}</span>
            </>
          )}
        </div>
        <div className="topbar-right">
          <span className="topbar-clock">{clock || "--:--:-- UTC"}</span>
          {view === "overview" && (
            <button className="btn-outline" onClick={openCreateModal}>
              + Create a Service
            </button>
          )}
        </div>
      </div>

      <div className="layout">
        {/* Sidebar */}
        <div className="sidebar">
          <div className="sidebar-section">
            <div
              className={`sidebar-item ${view === "overview" ? "active" : ""}`}
              onClick={backToOverview}
            >
              <span className="sidebar-icon">▤</span> Projects
            </div>
            <div className="sidebar-item">
              <span className="sidebar-icon">⎘</span> Blueprints
            </div>
            <div className="sidebar-item">
              <span className="sidebar-icon">▣</span> Environment Groups
            </div>
          </div>
          <div className="sidebar-label">Monitor</div>
          <div className="sidebar-section">
            <div className="sidebar-item">
              <span className="sidebar-icon">▥</span> Logs
            </div>
            <div className="sidebar-item">
              <span className="sidebar-icon">◷</span> Events
            </div>
          </div>
        </div>

        {/* Main */}
        <div className="orbit-main">
          {view === "overview" ? (
            <>
              <h1 className="page-title">Overview</h1>

              {/* Stats */}
              <div className="stats-row">
                <div className="stat-card">
                  <div className="stat-label">Services</div>
                  <div className="stat-val">{projects.length}</div>
                </div>
                <div className="stat-card">
                  <div className="stat-label">Running</div>
                  <div className="stat-val stat-green">{running}</div>
                </div>
                <div className="stat-card">
                  <div className="stat-label">Failed</div>
                  <div className="stat-val stat-red">{errors}</div>
                </div>
              </div>

              {/* Projects Table */}
              <div className="panel">
                <div className="panel-head">
                  <div className="panel-title">Services</div>
                  <button className="btn-ghost" onClick={loadProjects}>
                    Refresh
                  </button>
                </div>
                {projects.length === 0 ? (
                  <div className="empty-state">
                    <p>No services yet.</p>
                    <button className="btn-primary empty-cta" onClick={openCreateModal}>
                      Create a Service
                    </button>
                  </div>
                ) : (
                  <table>
                    <thead>
                      <tr>
                        <th>Service</th>
                        <th>Image</th>
                        <th>Status</th>
                        <th></th>
                      </tr>
                    </thead>
                    <tbody>
                      {projects.map((p) => (
                        <tr key={p.id} className="row-clickable" onClick={() => openProjectDetail(p.id)}>
                          <td>
                            <div className="td-name">{p.name}</div>
                            <div className="td-id">#{p.id}</div>
                          </td>
                          <td className="td-mono">{p.image_name}</td>
                          <td>
                            <StatusBadge status={p.status} />
                          </td>
                          <td onClick={(e) => e.stopPropagation()}>
                            <div className="act-row">
                              <button className="act-btn" onClick={() => deployProject(p.id)}>
                                Deploy
                              </button>
                              <button className="act-btn" onClick={() => getDeployments(p.id)}>
                                History
                              </button>
                              <button className="act-btn danger" onClick={() => deleteProject(p.id)}>
                                Delete
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
                <div className="panel">
                  <div className="panel-head">
                    <div className="panel-title">Deploy history</div>
                    <button className="btn-ghost" onClick={() => setDeployments([])}>
                      Clear
                    </button>
                  </div>
                  <table>
                    <thead>
                      <tr>
                        <th>Deploy</th>
                        <th>Service</th>
                        <th>Image</th>
                        <th>Status</th>
                        <th>Created</th>
                      </tr>
                    </thead>
                    <tbody>
                      {deployments.map((d) => (
                        <tr key={d.id}>
                          <td className="td-id">#{d.id}</td>
                          <td className="td-id">#{d.project_id}</td>
                          <td className="td-mono">{d.image_name}</td>
                          <td>
                            <StatusBadge status={d.status} />
                          </td>
                          <td className="td-mono">{new Date(d.created_at).toLocaleString()}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </>
          ) : (
            activeProject && (
              <>
                {/* Service detail header */}
                <div className="detail-head">
                  <div className="detail-kicker">
                    <span className="detail-kicker-icon">◆</span> SERVICE
                  </div>
                  <div className="detail-title-row">
                    <h1 className="page-title detail-title">{activeProject.name}</h1>
                    <StatusBadge status={activeProject.status} />
                  </div>
                  <div className="detail-meta">
                    <span className="detail-meta-label">Image</span>
                    <span className="td-mono">{activeProject.image_name}</span>
                    <span className="detail-meta-sep">·</span>
                    <span className="detail-meta-label">ID</span>
                    <span className="td-id">#{activeProject.id}</span>
                  </div>
                </div>

                {/* Action buttons */}
                <div className="panel detail-actions-panel">
                  <div className="panel-body detail-actions-body">
                    <button className="btn-primary" onClick={() => deployProject(activeProject.id)}>
                      Deploy
                    </button>
                    <button className="btn-outline" onClick={() => launchProject(activeProject.id)}>
                      Open
                    </button>
                    <button
                      className="btn-outline danger-outline"
                      onClick={() => deleteProject(activeProject.id)}
                    >
                      Delete
                    </button>
                  </div>
                </div>

                {message && (
                  <div className={`toast-msg ${message.startsWith("Failed") ? "toast-err" : "toast-ok"}`}>
                    {message}
                  </div>
                )}

                {/* Events log */}
                <div className="panel">
                  <div className="panel-head">
                    <div className="panel-title">
                      Events <span className="panel-sub">— {logsSource}</span>
                    </div>
                    <button className="btn-ghost" onClick={() => getLogs(activeProject.id)}>
                      Refresh
                    </button>
                  </div>
                  <div className="tele-body" ref={teleRef}>
                    {logs.length === 0 ? (
                      <div className="tele-empty">No events from the past 90 days match the selected filters.</div>
                    ) : (
                      logs.map((l, i) => (
                        <div key={i} className={`tele-line ${l.type}`}>
                          <span className="tele-ts">{l.ts}</span>
                          {l.text}
                        </div>
                      ))
                    )}
                  </div>
                </div>
              </>
            )
          )}
        </div>
      </div>

      {/* Create Service Modal */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={closeCreateModal}>
          <div className="modal-card" onClick={(e) => e.stopPropagation()}>
            <div className="modal-head">
              <h2 className="modal-title">Create a Service</h2>
              <button className="modal-close" onClick={closeCreateModal} aria-label="Close">
                ×
              </button>
            </div>
            <p className="modal-sub">Deploy a containerized service from a Docker image.</p>

            <div className="modal-body">
              <div className="form-field">
                <label className="form-label">Service name</label>
                <input
                  className="orbit-input"
                  placeholder="api-server"
                  value={modalName}
                  autoFocus
                  onChange={(e) => setModalName(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && submitCreateProject()}
                />
              </div>
              <div className="form-field">
                <label className="form-label">Docker image</label>
                <input
                  className="orbit-input"
                  placeholder="e.g. node:18-alpine or ghcr.io/you/app:latest"
                  value={modalImage}
                  onChange={(e) => setModalImage(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && submitCreateProject()}
                />
              </div>

              {modalError && <div className="modal-error">{modalError}</div>}
            </div>

            <div className="modal-actions">
              <button className="btn-outline" onClick={closeCreateModal} disabled={creating}>
                Cancel
              </button>
              <button className="btn-primary" onClick={submitCreateProject} disabled={creating}>
                {creating ? "Creating..." : "Create Service"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}