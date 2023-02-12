package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
)

var chans = make(map[string]chan Msg)
var Msgs = make([]Msg, 5)

type Msg struct {
	Msg  string `json:"msg"`
	Date string `json:"date"`
	Name string `json:"name"`
	Own  bool   `json:"own"`
	ID   string `json:"id"`
}

func websock(c echo.Context) error {
	// ID生成(UUID)
	uuidObj, _ := uuid.NewRandom()
	for {
		if _, ok := chans[uuidObj.String()]; ok {
			uuidObj, _ = uuid.NewRandom()
			continue
		} else {
			chans[uuidObj.String()] = make(chan Msg, 10)
			break
		}
	}

	c.Logger().Info("接続しました", uuidObj.String())

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		err := websocket.JSON.Send(ws, Msg{Msg: "接続しました", Date: time.Now().Format("2006/01/02 15:04:05"), Name: "System", Own: false, ID: uuidObj.String()})
		if err != nil {
			c.Logger().Error(err)
		}
		// Msgsの最新を送信
		// 逆順で取り出し
		for i := len(Msgs) - 1; i >= 0; i-- {
			if Msgs[i].Msg == "" {
				continue
			}
			err := websocket.JSON.Send(ws, Msgs[i])
			if err != nil {
				c.Logger().Error(err)
			}
		}

		log.Printf("Connected: ipaddr=%s uuid=%s", c.RealIP(), uuidObj.String())

		go func() {
			for {
				var msg Msg
				err := websocket.JSON.Receive(ws, &msg)
				if err != nil && err.Error() == "EOF" {
					c.Logger().Error(err)
					break
				}
				if msg.Msg == "" {
					continue
				}
				log.Printf("Recive: ipaddr=%s uuid=%s name=%s msg=%s", c.RealIP(), uuidObj.String(), msg.Name, msg.Msg)
				msg.ID = uuidObj.String()
				c.Logger().Info("recive", msg)
				// Msgsの最新に追加
				Msgs = append([]Msg{msg}, Msgs...)
				for _, v := range chans {
					v <- msg
				}
			}
		}()

		for {
			msg := <-chans[uuidObj.String()]
			c.Logger().Info("send", msg)
			if msg.ID == uuidObj.String() {
				msg.Own = true
			} else {
				msg.Own = false
			}
			err := websocket.JSON.Send(ws, msg)
			if err != nil && err.Error() == "EOF" {
				c.Logger().Error(err)
				break
			}
		}
	}).ServeHTTP(c.Response(), c.Request())

	// 接続切れ
	delete(chans, uuidObj.String())

	return nil
}

func main() {
	e := echo.New()
	e.IPExtractor = echo.ExtractIPFromXFFHeader(
		echo.TrustLinkLocal(true),
		echo.TrustIPRange(&net.IPNet{IP: net.IPv4(127, 0, 0, 0), Mask: net.IPv4Mask(255, 0, 0, 0)}),
	)
	e.HideBanner = true
	e.HidePort = true
	e.GET("/ws", websock)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := e.Start(":8010"); err != nil && err != http.ErrServerClosed {
			e.Logger.Info("shutting down the server")
		}
	}()

	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
