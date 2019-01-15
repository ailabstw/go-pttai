const { app,
        Menu,
        dialog,
        ipcMain,
        BrowserWindow } = require('electron')
const { autoUpdater }   = require('electron-updater')
const path              = require('path')
const spawn             = require('child_process').spawn


let gptt_node   = null
let gptt_win    = null;
let first_load  = true;
let force_quit  = false;

autoUpdater.autoDownload = false
autoUpdater.autoInstallOnAppQuit = false

const template = [
  {
    label: 'Edit',
    submenu: [
      { role: 'undo' },
      { role: 'redo' },
      { type: 'separator' },
      { role: 'cut' },
      { role: 'copy' },
      { role: 'paste' },
      { role: 'pasteandmatchstyle' },
      { role: 'delete' },
      { role: 'selectall' }
    ]
  },
  {
    label: 'View',
    submenu: [
      { role: 'reload' },
      { role: 'forcereload' },
      { role: 'toggledevtools' },
      { type: 'separator' },
      { role: 'resetzoom' },
      { role: 'zoomin' },
      { role: 'zoomout' },
      { type: 'separator' },
      { role: 'togglefullscreen' }
    ]
  },
  {
    role: 'window',
    submenu: [
      { role: 'minimize' },
      { role: 'close' }
    ]
  },
  {
    role: 'help',
    submenu: [
      {
        label: 'Learn More',
        click () { require('electron').shell.openExternal('https://electronjs.org') }
      }
    ]
  }
]

if (process.platform === 'darwin') {
  template.unshift({
    label: app.getName(),
    submenu: [
      { role: 'about' },
      { type: 'separator' },
      { role: 'services' },
      { type: 'separator' },
      { role: 'hide' },
      { role: 'hideothers' },
      { role: 'unhide' },
      { type: 'separator' },
      { role: 'quit' }
    ]
  })

  // Edit menu
  template[1].submenu.push(
    { type: 'separator' },
    {
      label: 'Speech',
      submenu: [
        { role: 'startspeaking' },
        { role: 'stopspeaking' }
      ]
    }
  )

  // Window menu
  template[3].submenu = [
    { role: 'close' },
    { role: 'minimize' },
    { role: 'zoom' },
    { type: 'separator' },
    { role: 'front' }
  ]
}

function createGptt() {
    // run node
    run_gptt_node()

    open_window()

    setup_menu()

    init_auto_updater()
}

function run_gptt_node () {
    console.log("gptt node start");
    let filepath = null
    let gpttFile = 'gptt'

    if (process.platform === "darwin") {
      gpttFile = 'gptt'
    } else if (process.platform === "win32") {
      gpttFile = 'gptt.exe'
    } else {
      // not supported for now
    }

    if(process.env.NODE_ENV === 'dev'){
       filepath = __dirname;
    } else {
       filepath = path.join(process.resourcesPath, "app");
    }

    const tmpLog = path.join(filepath, 'log.tmp.txt')
    const errLog = path.join(filepath, 'log.err.txt')

    //const gpttCmd = `${path.join(filepath, gpttFile)} --httpdir ${path.join(filepath, 'static')} --server --testp2p --log ${tmpLog} 2> ${errLog}`
    const proc = `${path.join(filepath, gpttFile)}`

    const args = [
      '--httpdir', path.join(filepath, 'static'),
      '--server',
      '--testp2p',
      '--log', tmpLog,
      '2>', errLog
    ]

    gptt_node = spawn(proc, args, {setsid:true});
}

function load_content () {

    if (first_load) {
        setTimeout(load_content, 8000);
        let loadingUrl = require('url').format({
          protocol: 'file',
          slashes: true,
          pathname: require('path').join(__dirname, 'loading.html')
        })
        gptt_win.loadURL(loadingUrl)
        first_load = false
    } else {
        gptt_win.loadURL('http://localhost:9774')
    }
}

function open_window () {

    if (!gptt_win) {

        gptt_win = new BrowserWindow({
          width: 414,
          height: 1000,
        })

        // Continue to handle mainWindow "close" event here
        gptt_win.on('close', (e) => {
          e.preventDefault();
          gptt_win.hide();

          if (process.platform !== "darwin") {
            //spawn("taskkill", ["/PID", gptt_node.pid, "/F", "T"])
            spawn("taskkill", ["/F", "/IM", "gptt.exe", "/T"])
            spawn("taskkill", ["/F", "/IM", "Pttai.exe", "/T"])
          }
        });

        app.on('before-quit', (e) => {
          // Handle menu-item or keyboard shortcut quit here
          gptt_win = null
          if (process.platform !== "darwin") {
            spawn("taskkill", ["/F", "/IM", "gptt.exe", "/T"])
            spawn("taskkill", ["/F", "/IM", "Pttai.exe", "/T"])
          } else {
            gptt_node.kill()
          }
          app.exit()
        });

        app.on('activate-with-no-open-windows', function(){
          gptt_win.show();
        });
    }

    load_content()
}

function setup_menu() {
    const menu = Menu.buildFromTemplate(template)
    Menu.setApplicationMenu(menu)
}

function init_auto_updater() {
  autoUpdater.checkForUpdates();
}

app.on('ready', createGptt)

app.on('window-all-closed', () => {
  // On macOS it is common for applications and their menu bar
  // to stay active until the user quits explicitly with Cmd + Q
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('activate', () => {
  // On macOS it's common to re-create a window in the app when the
  // dock icon is clicked and there are no other windows open.
  // open_window()
  gptt_win.show()
})

autoUpdater.on('error', (err) => console.log(err));
autoUpdater.on('checking-for-update', () => console.log('autoUpdater: checking-for-update'));

autoUpdater.on('update-not-available', () => console.log('autoUpdater: update-not-available'));
autoUpdater.on('update-available', (info) => {

  console.log('autoUpdater: update-available ', info)
  let message = app.getName() + ' ' + info.version + ' is now available. Do you want it to be downloaded and installed?';
  // if (releaseNotes) {
  //   const splitNotes = releaseNotes.split(/[^\r]\n/);
  //   message += '\n\nRelease notes:\n';
  //   splitNotes.forEach(notes => {
  //     message += notes + '\n\n';
  //   });
  // }

  dialog.showMessageBox({
    type: 'question',
    buttons: ['Install and Relaunch', 'Later'],
    defaultId: 0,
    message: 'A new version of ' + app.getName() + ' is available',
    detail: message
  }, response => {
    if (response === 0) {
      setTimeout(() => {
        console.log('autoUpdater: download-update')
        autoUpdater.downloadUpdate()
      }, 1);
    }
  });

});

// Ask the user if update is available
autoUpdater.on('update-downloaded', (event, releaseNotes, releaseName) => {
  console.log('autoUpdater: quit-and-install')
  autoUpdater.quitAndInstall()
  app.quit();
});
