<script lang="ts">
  import Router, { link } from "svelte-spa-router";
  import active from "svelte-spa-router/active";
  import WorkHistory from "./pages/WorkHistory.svelte";
  import Placeholder from "./pages/Placeholder.svelte";
  import Toast from "./components/Toast.svelte";

  const routes = {
    "/": WorkHistory,
    "/work-history": WorkHistory,
    "/skills": Placeholder,
    "/education": Placeholder,
    "/certifications": Placeholder,
    "/summaries": Placeholder,
    "/descriptors": Placeholder,
    "/lenses": Placeholder,
    "/export": Placeholder,
    "/cover-letters": Placeholder,
    "/applications": Placeholder,
    "/settings": Placeholder,
  };

  interface NavItem {
    path: string;
    label: string;
  }

  const navItems: NavItem[] = [
    { path: "/work-history", label: "Work History" },
    { path: "/skills", label: "Skills" },
    { path: "/education", label: "Education" },
    { path: "/certifications", label: "Certifications" },
    { path: "/summaries", label: "Summaries" },
    { path: "/descriptors", label: "Role Descriptors" },
    { path: "/lenses", label: "Lenses" },
    { path: "/export", label: "Export" },
    { path: "/cover-letters", label: "Cover Letters" },
    { path: "/applications", label: "Applications" },
    { path: "/settings", label: "Settings" },
  ];
</script>

<div class="app-layout">
  <nav class="sidebar">
    <div class="sidebar-header">
      <h1>Cut the BS</h1>
    </div>
    <ul class="nav-list">
      {#each navItems as item (item.path)}
        <li>
          <a
            href={"#" + item.path}
            use:link
            use:active={{ path: item.path, className: "active" }}
          >
            {item.label}
          </a>
        </li>
      {/each}
    </ul>
  </nav>
  <main class="content">
    <Router {routes} />
  </main>
  <Toast />
</div>

<style>
  .app-layout {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .sidebar {
    width: 220px;
    min-width: 220px;
    background-color: #1a2332;
    border-right: 1px solid #2a3a4a;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
  }

  .sidebar-header {
    padding: 20px 16px 12px;
    border-bottom: 1px solid #2a3a4a;
  }

  .sidebar-header h1 {
    margin: 0;
    font-size: 1.2rem;
    font-weight: 700;
    letter-spacing: 0.02em;
    color: #e0e0e0;
  }

  .nav-list {
    list-style: none;
    margin: 0;
    padding: 8px 0;
  }

  .nav-list li {
    margin: 0;
  }

  .nav-list a {
    display: block;
    padding: 10px 16px;
    color: #a0b0c0;
    text-decoration: none;
    font-size: 0.9rem;
    transition:
      background-color 0.15s,
      color 0.15s;
  }

  .nav-list a:hover {
    background-color: #223344;
    color: #e0e0e0;
  }

  .nav-list a:global(.active) {
    background-color: #2a4060;
    color: #ffffff;
    font-weight: 600;
  }

  .content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    background-color: #1b2636;
  }
</style>
