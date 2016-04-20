package control

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"
)

import (
	"github.com/mikydna/x/managed"
	"golang.org/x/net/websocket"
)

var (
	CookieName = "_z12345"
)

func ForwardLog(udpPort uint) http.HandlerFunc {
	conns := managed.NewWebsockets()

	go func() {
		en0, _ := net.InterfaceByName("en0")
		log.Println(en0.MulticastAddrs())

		addrStr := fmt.Sprintf("224.0.0.251:%d", udpPort) // wat?
		addr, _ := net.ResolveUDPAddr("udp", addrStr)

		log.Println("udp=", addr)

		udp, err := net.ListenMulticastUDP("udp", nil, addr)
		if err != nil {
			log.Println(err)
			return
		}

		alive := true
		buf := make([]byte, 1024)
		for alive {
			n, err := udp.Read(buf)
			if err != nil {
				alive = false
				log.Println(err)
			}

			if n > 0 {
				conns.Broadcast(buf[:n])
			}

		}

	}()

	usingWebsocket := func(conn *websocket.Conn) {
		cookie, err := conn.Request().Cookie(CookieName)
		if err != nil {
			log.Println(err)
			return
		}

		connId := cookie.Value
		conns.Set(connId, conn)
		conns.Wait(connId)
	}

	return websocket.Handler(usingWebsocket).ServeHTTP
}

func ShowLog(host string) http.HandlerFunc {

	return func(w http.ResponseWriter, req *http.Request) {
		var connId string
		if cookie, err := req.Cookie(CookieName); err == nil {
			connId = cookie.Value

		} else {
			connId = fmt.Sprintf("%d", rand.Uint32())

		}

		cookie := http.Cookie{
			Name:    CookieName,
			Value:   connId,
			Expires: time.Now().Add(5 * time.Minute),
		}

		http.SetCookie(w, &cookie)

		t, _ := template.New("show").Parse(LogShow)
		t.Execute(w, struct {
			Host   string
			Cookie string
		}{
			Host:   host,
			Cookie: CookieName,
		})

	}

}
