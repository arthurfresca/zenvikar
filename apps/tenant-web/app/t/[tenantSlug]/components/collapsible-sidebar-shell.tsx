"use client";

import { useState, useEffect } from "react";
import { SidebarContext } from "./sidebar-context";

type Props = {
  sidebar: React.ReactNode;
  children: React.ReactNode;
};

export function CollapsibleSidebarShell({ sidebar, children }: Props) {
  const [collapsed, setCollapsed] = useState(false);
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (localStorage.getItem("zenvikar_sidebar") === "collapsed") setCollapsed(true);
    setReady(true);
  }, []);

  const toggle = () => {
    setCollapsed((c) => {
      const next = !c;
      localStorage.setItem("zenvikar_sidebar", next ? "collapsed" : "expanded");
      return next;
    });
  };

  return (
    <SidebarContext.Provider value={{ collapsed }}>
      <div className="flex min-h-screen bg-gray-50 dark:bg-zinc-950">
        {/* Sidebar */}
        <aside
          data-collapsed={collapsed ? "true" : "false"}
          className={[
            "group relative flex flex-shrink-0 flex-col overflow-hidden border-r border-gray-200 bg-white dark:border-zinc-800 dark:bg-zinc-950",
            "lg:fixed lg:inset-y-0 lg:left-0 lg:z-40",
            ready ? "transition-[width] duration-200 ease-in-out" : "",
            collapsed ? "w-16" : "w-64",
          ].join(" ")}
        >
          {sidebar}

          {/* Toggle button — floats on the right edge, desktop only */}
          <button
            type="button"
            onClick={toggle}
            aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
            className="absolute -right-3 top-[26px] z-50 hidden h-6 w-6 items-center justify-center rounded-full border border-gray-200 bg-white shadow-sm transition-colors hover:bg-gray-100 dark:border-zinc-700 dark:bg-zinc-900 dark:hover:bg-zinc-800 lg:flex"
          >
            <svg
              className={[
                "h-3 w-3 text-gray-500 transition-transform duration-200 dark:text-zinc-400",
                collapsed ? "rotate-180" : "",
              ].join(" ")}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2.5}
            >
              <path strokeLinecap="round" strokeLinejoin="round" d="M15 19l-7-7 7-7" />
            </svg>
          </button>
        </aside>

        {/* Main content */}
        <div
          className={[
            "flex min-h-screen flex-1 flex-col",
            ready ? "transition-[padding-left] duration-200 ease-in-out" : "",
            collapsed ? "lg:pl-16" : "lg:pl-64",
          ].join(" ")}
        >
          <main className="flex-1 px-6 py-8 lg:px-8 xl:px-10">{children}</main>
        </div>
      </div>
    </SidebarContext.Provider>
  );
}
