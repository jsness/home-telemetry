import React, { useEffect, useState } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

type Node = {
  id: string;
  name: string;
  last_seen: string;
  meta?: Record<string, string>;
};

const App = () => {
  const [nodes, setNodes] = useState<Node[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      try {
        const res = await fetch(`/api/v1/nodes`);
        if (!res.ok) {
          throw new Error(`Request failed: ${res.status}`);
        }
        const data = (await res.json()) as Node[];
        setNodes(data);
      } catch (e) {
        setError(e instanceof Error ? e.message : "Unknown error");
      }
    };
    load();
  }, []);

  return (
    <div className="app">
      <header className="header">
        <h1>Home Telemetry</h1>
        <p>Nodes</p>
      </header>
      <main className="content">
        {error && <div className="card">Error: {error}</div>}
        {!error && nodes.length === 0 && <div className="card">No nodes yet.</div>}
        {!error && nodes.length > 0 && (
          <div className="card">
            <ul>
              {nodes.map((n) => (
                <li key={n.id}>
                  <strong>{n.name}</strong> <span>({n.id})</span>
                  <div>Last seen: {new Date(n.last_seen).toLocaleString()}</div>
                </li>
              ))}
            </ul>
          </div>
        )}
      </main>
    </div>
  );
};

const root = createRoot(document.getElementById("root")!);
root.render(<App />);