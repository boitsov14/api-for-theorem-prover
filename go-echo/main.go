package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"sync"

	_ "github.com/joho/godotenv/autoload"
	"github.com/juunini/simple-go-line-notify/notify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	mutex sync.Mutex // Mutex for exclusive control of execution of prover
)

func main() {
	// create echo instance
	e := echo.New()
	e.Use(middleware.CORS())

	e.POST("/web", func(c echo.Context) error {
		// get request
		req := new(WebReq)
		if err := c.Bind(req); err != nil {
			return err
		}
		txt := req.Txt
		notifyLine("Web: " + txt)

		// process request
		webRes, err := processWebReq(txt)
		if err != nil {
			notifyLine("Unexpected error has occurred.\n" + err.Error())
			return c.String(http.StatusInternalServerError, "")
		}

		// return response
		return c.JSON(http.StatusOK, webRes)
	})

	// get port
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// start server
	e.Logger.Fatal(e.Start(":" + port))
}

type WebReq struct {
	Txt string `json:"txt"`
}

type WebRes struct {
	Msg string `json:"msg"`
	Img string `json:"img"`
	Tex string `json:"tex"`
}

func processWebReq(txt string) (*WebRes, error) {
	// lock
	mutex.Lock()
	defer mutex.Unlock()

	// create temp dir
	dir, err := os.MkdirTemp(".", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	// change dir
	if err := os.Chdir(dir); err != nil {
		return nil, err
	}
	defer os.Chdir("..")

	// symlink ../prover
	if err := os.Symlink("../prover", "prover"); err != nil {
		return nil, err
	}

	// run prover
	msg, err := prove(txt, "500m", 10)
	if err != nil {
		return nil, err
	}
	// make dvi
	msgDVI, err := makeDVI()
	if err != nil {
		return nil, err
	}
	// make png
	msgPNG, err := makePNG()
	if err != nil {
		return nil, err
	}
	msg += msgDVI + msgPNG

	// create response
	res := &WebRes{
		Msg: msg,
	}

	// if out.png exists
	if _, err := os.Stat("out.png"); err == nil {
		notifyLineWithImage(msg)
		// read out.png
		img, err := os.ReadFile("out.png")
		if err != nil {
			return nil, err
		}
		res.Img = base64.StdEncoding.EncodeToString(img)
		// read out.tex
		tex, err := os.ReadFile("out.tex")
		if err != nil {
			return nil, err
		}
		res.Tex = string(tex)
	} else {
		notifyLine(msg)
	}

	return res, nil
}

func notifyLine(msg string) {
	if err := notify.SendText(os.Getenv("LINE_ACCESS_TOKEN"), msg); err != nil {
		fmt.Println("LINE Notification Error: ", err)
	}
}

func notifyLineWithImage(msg string) {
	if err := notify.SendLocalImage(os.Getenv("LINE_ACCESS_TOKEN"), msg, "out.png"); err != nil {
		fmt.Println("LINE Notification Error: ", err)
	}
}
