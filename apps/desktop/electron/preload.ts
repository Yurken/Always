import { contextBridge, ipcRenderer } from "electron";

contextBridge.exposeInMainWorld("always", {
  version: process.env.APP_VERSION || "0.1.0",
  moveWindow: (x: number, y: number) => ipcRenderer.invoke("window-move", { x, y }),
  setIgnoreMouseEvents: (ignore: boolean) =>
    ipcRenderer.invoke("window-ignore-mouse", { ignore }),
});
