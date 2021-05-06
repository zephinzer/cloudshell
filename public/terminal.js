(function() {
  var terminal = new Terminal({
    screenKeys: true,
    useStyle: true,
    cursorBlink: true,
    fullscreenWin: true,
    maximizeWin: true,
    cols: 128,
  });
  terminal.open(document.getElementById("terminal"));
  var url = "ws://" + "localhost" + ":8376" + "/xterm.js"
  var ws = new WebSocket(url);
  var attachAddon = new AttachAddon.AttachAddon(ws);
  var fitAddon = new FitAddon.FitAddon();
  terminal.loadAddon(fitAddon);
  ws.onopen = function() {
    terminal.loadAddon(attachAddon);
    terminal._initialized = true;
    terminal.focus();
    fitAddon.fit();
    terminal.onResize(function(event) {
      var rows = event.rows;
      var cols = event.cols;
      var size = JSON.stringify({cols: cols, rows: rows});
      var send = new TextEncoder().encode("\x01" + size);
      console.log('resizing to', size);
      ws.send(send);
    });
    terminal.onTitleChange(function(event) {
      console.log(event);
    });
    window.onresize = function() {
      fitAddon.fit();
    };
  };  
})();
