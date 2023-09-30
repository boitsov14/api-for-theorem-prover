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
	e.HideBanner = true
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	e.POST("/web", func(c echo.Context) error {
		// get request
		req := new(WebReq)
		if err := c.Bind(req); err != nil {
			return err
		}
		sequent := req.Txt
		notifyLine("Web: " + sequent)

		// prove
		result, err := prove(sequent, "500m", 10, true)
		if err != nil {
			notifyLine("Unexpected error has occurred.\n" + err.Error())
			return c.String(http.StatusInternalServerError, "")
		}

		// create response
		res := &WebRes{
			Msg: result.Msg,
			Img: base64.StdEncoding.EncodeToString(result.Img),
			Tex: result.Tex,
		}

		// return response
		return c.JSON(http.StatusOK, res)
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

func notifyLine(msg string) {
	fmt.Println(msg)
	if err := notify.SendText(os.Getenv("LINE_ACCESS_TOKEN"), msg); err != nil {
		fmt.Println("LINE Notification Error: ", err)
	}
}

func notifyLineWithImage(msg string) {
	fmt.Println(msg)
	if err := notify.SendLocalImage(os.Getenv("LINE_ACCESS_TOKEN"), msg, "out.png"); err != nil {
		fmt.Println("LINE Notification Error: ", err)
	}
}
