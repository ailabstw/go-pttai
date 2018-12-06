const { app, BrowserWindow, Menu } = require('electron')
const path  = require('path')

let exec    = require('child_process').execFile;

let gptt_node   = null
let gptt_win    = null;
let first_load  = true;
let force_quit = false;

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

    const menu = Menu.buildFromTemplate(template)
    Menu.setApplicationMenu(menu)
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

    gptt_node = exec(path.join(filepath, gpttFile) , ['--httpdir', path.join(filepath, 'static'), '--server'], function(err, data) {
        console.log('err:', err)
        console.log('data:',data)
    });
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
          width: 554,
          height: 1000,
        })

        // Continue to handle mainWindow "close" event here
        gptt_win.on('close', (e) => {
          e.preventDefault();
          gptt_win.hide();
        });

        app.on('before-quit', (e) => {
          // Handle menu-item or keyboard shortcut quit here
          gptt_win = null
          gptt_node.kill()
          app.exit()
        });

        app.on('activate-with-no-open-windows', function(){
          gptt_win.show();
        });
    }

    load_content()
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
  //open_window()
  gptt_win.show()
})
