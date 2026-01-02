import { contextBridge } from "electron";

contextBridge.exposeInMainWorld("luma", {
  version: "0.1.0",
});
