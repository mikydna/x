package control

const LogShow = `
<html>
  <head>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/rxjs/4.1.0/rx.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/rxjs-dom/7.0.3/rx.dom.js"></script>
    <style>
      html, body {
        margin: 0;
        padding: 0;
      }

      span.line {
        display: block;
        padding-left: 0.5rem;
        padding-right: 0.5rem;
        font-family: monospace;
        font-size: 0.8rem;
        line-height: 1.2rem;
      }

      span.line:first-child {
        background-color: black;
        color: white;
        font-size: 1rem;
        line-height: 1.5rem;
      }

      span.line:nth-child(n+50) {
        color: green;
      }
    </style>
  </head>
  <body>
    <article id="logs"></article>
    <script>
      const z = document.cookie.replace(/(?:(?:^|.*;\s*)_z12345\s*\=\s*([^;]*).*$)|^.*$/, "$1");
      const el = document.getElementById("logs");
      const $log = Rx.DOM.fromWebSocket(
        'ws://{{.Host}}/x/log?id=' + z,
        null,
        new Rx.Subject(),
        new Rx.Subject());

      $log.
        map(e => (e.data)).
        filter(data => (data.length !== 0)).
        tap(str => {
            const line = document.createElement('span');
            line.className = 'line';
            line.innerHTML = str;
            el.insertBefore(line, el.firstChild);
          }).
        tap(() => {
            const lines = el.getElementsByClassName('line');
            for (let i=50; i<lines.length; i++) {
              lines[i].remove()
            }
          }).
        subscribe();

    </script>
  </body>
</html>
`
