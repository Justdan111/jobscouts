import type { Config } from "tailwindcss";
const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: { extend: {
    colors: {
      paper:"#F6F7F9", surface:"#FFFFFF", ink:"#15171C", muted:"#6B7280",
      line:"#E6E8EC", signal:"#117C6F", "signal-dim":"#E1EFEC", heat:"#E0913A",
    },
    fontFamily: {
      display:["'Space Grotesk'","system-ui","sans-serif"],
      sans:["Inter","system-ui","sans-serif"],
      mono:["'IBM Plex Mono'","ui-monospace","monospace"],
    },
  }},
  plugins: [],
};
export default config;
