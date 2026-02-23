<script lang="ts">
  import Router, { link } from "svelte-spa-router";
  import active from "svelte-spa-router/active";
  import WorkHistory from "./pages/WorkHistory.svelte";
  import Skills from "./pages/Skills.svelte";
  import Education from "./pages/Education.svelte";
  import Summaries from "./pages/Summaries.svelte";
  import Descriptors from "./pages/Descriptors.svelte";
  import CoreExpertise from "./pages/CoreExpertise.svelte";
  import Export from "./pages/Export.svelte";
  import Applications from "./pages/Applications.svelte";
  import Lenses from "./pages/Lenses.svelte";
  import Settings from "./pages/Settings.svelte";
  import TemplateList from "./pages/TemplateList.svelte";
  import TemplateBuilder from "./pages/TemplateBuilder.svelte";
  import Toast from "./components/Toast.svelte";
  import StatusBar from "./components/StatusBar.svelte";
  import ZoomWidget from "./components/ZoomWidget.svelte";

  const routes = {
    "/": Applications,
    "/work-history": WorkHistory,
    "/skills": Skills,
    "/education": Education,
    "/summaries": Summaries,
    "/descriptors": Descriptors,
    "/core-expertise": CoreExpertise,
    "/lenses": Lenses,
    "/export": Export,
    "/applications": Applications,
    "/templates": TemplateList,
    "/templates/:id/builder": TemplateBuilder,
    "/settings": Settings,
  };

  interface NavItem {
    path: string;
    label: string;
    primary?: boolean;
  }

  const navItems: NavItem[] = [
    { path: "/", label: "Applications", primary: true },
    { path: "/templates", label: "Templates" },
    { path: "/work-history", label: "Work History" },
    { path: "/skills", label: "Skills" },
    { path: "/education", label: "Education" },
    { path: "/summaries", label: "Summaries" },
    { path: "/descriptors", label: "Role Descriptors" },
    { path: "/core-expertise", label: "Core Expertise" },
    { path: "/lenses", label: "Lenses" },
    { path: "/export", label: "Export" },
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
            class:primary-link={item.primary}
            href={"#" + item.path}
            use:link
            use:active={{ path: item.path, className: "active" }}
          >
            {item.label}
          </a>
        </li>
        {#if item.primary}
          <li class="nav-divider" aria-hidden="true"></li>
        {/if}
      {/each}
    </ul>
  </nav>
  <div class="main-column">
    <main class="content">
      <Router {routes} />
    </main>
    <StatusBar />
  </div>
  <Toast />
  <ZoomWidget />
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

  .nav-list a.primary-link {
    font-weight: 700;
    color: #dce7f5;
    font-size: 0.94rem;
  }

  .nav-divider {
    border-top: 1px solid #2a3a4a;
    margin: 8px 12px;
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

  .main-column {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
</style>
