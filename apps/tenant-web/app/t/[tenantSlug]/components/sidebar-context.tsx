"use client";

import { createContext, useContext } from "react";

type SidebarCtx = { collapsed: boolean };

export const SidebarContext = createContext<SidebarCtx>({ collapsed: false });
export const useSidebarContext = () => useContext(SidebarContext);
