import {
  app,
  BrowserWindow,
  globalShortcut,
  ipcMain,
  Menu,
  nativeImage,
  screen,
  Tray,
} from "electron";
import path from "path";

const DEV_SERVER_URL = "http://localhost:5173";

let mainWindow: BrowserWindow | null = null;
let settingsWindow: BrowserWindow | null = null;
let tray: Tray | null = null;
let isQuitting = false;

const buildTrayIcon = () => {
  const svg = `
    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 16 16">
      <circle cx="8" cy="8" r="7" fill="#111827" />
      <circle cx="6" cy="7" r="1" fill="#ffffff" />
      <circle cx="10" cy="7" r="1" fill="#ffffff" />
      <rect x="5" y="10" width="6" height="1.5" rx="0.75" fill="#ffffff" />
    </svg>
  `;
  const icon = nativeImage.createFromDataURL(
    `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`
  );
  if (process.platform === "darwin") {
    icon.setTemplateImage(true);
  }
  return icon;
};

const buildTrayMenu = () => {
  const visible = mainWindow?.isVisible() ?? true;
  return Menu.buildFromTemplate([
    {
      label: visible ? "隐藏悬浮球" : "显示悬浮球",
      click: () => toggleMainWindow(),
    },
    {
      label: "设置",
      click: () => createSettingsWindow(),
    },
    { type: "separator" },
    {
      label: "退出",
      click: () => {
        isQuitting = true;
        app.quit();
      },
    },
  ]);
};

const updateTrayMenu = () => {
  if (tray) {
    tray.setContextMenu(buildTrayMenu());
  }
};

const showMainWindow = (focus = false) => {
  if (!mainWindow || mainWindow.isDestroyed()) {
    return;
  }
  if (!mainWindow.isVisible()) {
    if (focus) {
      mainWindow.show();
    } else {
      mainWindow.showInactive();
    }
  }
  if (focus) {
    mainWindow.focus();
  }
  updateTrayMenu();
};

const hideMainWindow = () => {
  if (!mainWindow || mainWindow.isDestroyed()) {
    return;
  }
  mainWindow.hide();
  updateTrayMenu();
};

const toggleMainWindow = () => {
  if (!mainWindow || mainWindow.isDestroyed()) {
    return;
  }
  if (mainWindow.isVisible()) {
    hideMainWindow();
  } else {
    showMainWindow(false);
  }
};

const createWindow = () => {
  const win = new BrowserWindow({
    width: 360,
    height: 520,
    resizable: false,
    transparent: true,
    frame: false,
    alwaysOnTop: true,
    skipTaskbar: true,
    hasShadow: false,
    focusable: false,
    show: false,
    backgroundColor: "#00000000",
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      webSecurity: false, // 允许跨域请求
    },
  });

  mainWindow = win;

  win.loadURL(DEV_SERVER_URL);
  win.setIgnoreMouseEvents(true, { forward: true });

  win.once("ready-to-show", () => {
    win.showInactive();
    updateTrayMenu();
  });

  win.on("close", (event) => {
    if (!isQuitting) {
      event.preventDefault();
      win.hide();
      updateTrayMenu();
    }
  });

  win.on("hide", updateTrayMenu);
  win.on("show", updateTrayMenu);

  // 自动打开开发者工具
  win.webContents.openDevTools({ mode: "detach" });

  attachContextMenu(win);
  return win;
};

const createSettingsWindow = () => {
  if (settingsWindow && !settingsWindow.isDestroyed()) {
    settingsWindow.show();
    settingsWindow.focus();
    return;
  }
  settingsWindow = new BrowserWindow({
    width: 420,
    height: 560,
    resizable: true,
    transparent: false,
    frame: true,
    backgroundColor: "#f3f4f6",
    webPreferences: {
      preload: path.join(__dirname, "preload.js"),
      webSecurity: false, // 允许跨域请求
    },
  });

  settingsWindow.loadURL(`${DEV_SERVER_URL}/?settings=1`);
  settingsWindow.on("closed", () => {
    settingsWindow = null;
  });

  attachContextMenu(settingsWindow);
};

const createTray = () => {
  if (tray) {
    return;
  }
  tray = new Tray(buildTrayIcon());
  tray.setToolTip("Always");
  tray.setContextMenu(buildTrayMenu());
  tray.on("click", () => toggleMainWindow());
};

const registerShortcuts = () => {
  globalShortcut.register("CommandOrControl+Shift+O", () => toggleMainWindow());
};

const attachContextMenu = (win: BrowserWindow) => {
  win.webContents.on("context-menu", () => {
    const menu = Menu.buildFromTemplate([
      {
        label: mainWindow?.isVisible() ? "隐藏悬浮球" : "显示悬浮球",
        click: () => toggleMainWindow(),
      },
      {
        label: "设置",
        click: () => createSettingsWindow(),
      },
      { type: "separator" },
      {
        label: "退出",
        click: () => {
          isQuitting = true;
          app.quit();
        },
      },
    ]);
    menu.popup({ window: win });
  });
};

app.whenReady().then(() => {
  createWindow();
  createTray();
  registerShortcuts();

  ipcMain.handle("window-move", (event, { x, y }) => {
    const win = BrowserWindow.fromWebContents(event.sender);
    if (win) {
      win.setPosition(Math.round(x), Math.round(y));
    }
  });

  ipcMain.handle("window-ignore-mouse", (event, { ignore }) => {
    const win = BrowserWindow.fromWebContents(event.sender);
    if (!win) {
      return;
    }
    if (ignore) {
      win.setIgnoreMouseEvents(true, { forward: true });
    } else {
      win.setIgnoreMouseEvents(false);
    }
  });

  ipcMain.handle("window-show", (_event, { focus }) => {
    showMainWindow(Boolean(focus));
  });

  ipcMain.handle("window-hide", () => {
    hideMainWindow();
  });

  ipcMain.handle("window-toggle", () => {
    toggleMainWindow();
  });

  ipcMain.handle("window-set-focusable", (event, { focusable }) => {
    const win = BrowserWindow.fromWebContents(event.sender);
    if (!win || typeof win.setFocusable !== "function") {
      return;
    }
    win.setFocusable(Boolean(focusable));
    if (!focusable && win.isFocused()) {
      win.blur();
    }
  });

  ipcMain.handle("window-display-bounds", (event) => {
    const win = BrowserWindow.fromWebContents(event.sender);
    if (!win) {
      return null;
    }
    const bounds = win.getBounds();
    const display = screen.getDisplayMatching(bounds);
    return { bounds: display.bounds, workArea: display.workArea };
  });

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
      return;
    }
    showMainWindow(false);
  });
});

app.on("before-quit", () => {
  isQuitting = true;
});

app.on("will-quit", () => {
  globalShortcut.unregisterAll();
  if (tray) {
    tray.destroy();
    tray = null;
  }
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
